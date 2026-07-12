package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"

	"github.com/google/uuid"
	"github.com/razorpay/razorpay-go"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	applogger "github.com/rohit221990/mandi-backend/pkg/logger"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/service/otp"
	"github.com/rohit221990/mandi-backend/pkg/service/sms"
	"github.com/rohit221990/mandi-backend/pkg/service/token"
	service "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type adminUseCase struct {
	adminRepo         interfaces.AdminRepository
	userRepo          interfaces.UserRepository
	authRepo          interfaces.AuthRepository
	optAuth           otp.OtpAuth
	tokenService      token.TokenService
	otpService        *otp.MobileOTPService
	smsService        *sms.TwoFactorSMSService
	skipOTPValidation bool
	config            config.Config
}

func NewAdminUseCase(repo interfaces.AdminRepository, userRepo interfaces.UserRepository, authRepo interfaces.AuthRepository, optAuth otp.OtpAuth, tokenService token.TokenService, otpService *otp.MobileOTPService, smsService *sms.TwoFactorSMSService, skipOTPValidation bool, cfg config.Config) service.AdminUseCase {

	return &adminUseCase{
		adminRepo:         repo,
		userRepo:          userRepo,
		authRepo:          authRepo,
		optAuth:           optAuth,
		tokenService:      tokenService,
		otpService:        otpService,
		smsService:        smsService,
		skipOTPValidation: skipOTPValidation,
		config:            cfg,
	}
}

func (c *adminUseCase) SignUp(ctx context.Context, signUpDetails domain.Admin) (string, error) {

	// Validate mobile number
	if signUpDetails.Mobile == "" || signUpDetails.Mobile == "null" {
		return "", fmt.Errorf("mobile number is required")
	}

	// Check if admin already exists by phone
	existAdmin, err := c.adminRepo.FindAdminByPhone(ctx, signUpDetails.Mobile)
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to check admin details already exist")
	}
	applogger.L().Debug("existing admin lookup", zap.String("event", applogger.EventSecurityLoginSuccess), zap.String("admin_id", existAdmin.ID))
	// If admin already exists, return error
	if existAdmin.ID != "" && existAdmin.VerifiedSeller {
		return "", errors.New("Admin already exists with this phone")
	}

	// Check if email is provided and already exists
	// if signUpDetails.Email != "" {
	// 	existAdminByEmail, err := c.adminRepo.FindAdminByEmail(ctx, signUpDetails.Email)
	// 	if err != nil {
	// 		return "", utils.PrependMessageToError(err, "failed to check admin email already exist")
	// 	}
	// 	if existAdminByEmail.ID != 0 {
	// 		return "", errors.New("can't save admin - an admin already exists with this email")
	// 	}
	// }

	// Save admin (hash the password first)
	hashPass, err := utils.GenerateHashFromPassword(signUpDetails.Password)
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to hash the password")
	}
	signUpDetails.Password = string(hashPass)

	savedAdmin, err := c.adminRepo.SaveAdmin(ctx, signUpDetails)
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to save admin details")
	}
	adminID := savedAdmin.ID
	applogger.L().Info("seller account created", zap.String("event", applogger.EventSecurityOTPSent), zap.String("admin_id", adminID))

	return c.issueOtpSession(ctx, adminID, signUpDetails.Mobile)
}

// SignupOtpSend sends a signup/login OTP for a seller. It is idempotent with
// respect to the admin record: if an admin already exists for the phone it
// reuses it (login / resend); otherwise it creates a new seller using the
// provided full_name/password (signup). In all cases it stores a hashed OTP
// session and sends the OTP via SMS, returning the otp_id used on verify.
func (c *adminUseCase) SignupOtpSend(ctx context.Context, details domain.Admin) (string, error) {
	if details.Mobile == "" || details.Mobile == "null" {
		return "", fmt.Errorf("mobile number is required")
	}

	adminID, err := c.findOrCreateAdminByMobile(ctx, details)
	if err != nil {
		return "", err
	}

	return c.issueOtpSession(ctx, adminID, details.Mobile)
}

// findOrCreateAdminByMobile returns the ID of the existing seller for the given
// mobile number, or creates a new minimal seller record and returns its ID.
// It is safe under concurrent requests: if two goroutines race on the same new
// number, the one that loses the INSERT recovers by re-fetching the winner's row.
func (c *adminUseCase) findOrCreateAdminByMobile(ctx context.Context, details domain.Admin) (string, error) {
	existing, err := c.adminRepo.FindAdminByPhone(ctx, details.Mobile)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", utils.PrependMessageToError(err, "failed to look up seller by mobile")
	}
	if existing.ID != "" {
		return existing.ID, nil
	}

	// New number — prepare and insert a minimal seller record.
	if details.Password != "" {
		hashPass, hashErr := utils.GenerateHashFromPassword(details.Password)
		if hashErr != nil {
			return "", utils.PrependMessageToError(hashErr, "failed to hash password")
		}
		details.Password = string(hashPass)
	}
	if details.Role == "" {
		details.Role = domain.AdminRoleSeller
	}
	if details.Status == "" {
		details.Status = "active"
	}

	saved, saveErr := c.adminRepo.SaveAdmin(ctx, details)
	if saveErr == nil {
		return saved.ID, nil
	}

	// Handle race condition: another request inserted the same mobile concurrently.
	// Recover by re-fetching the winner's row.
	if strings.Contains(saveErr.Error(), "unique constraint") ||
		strings.Contains(saveErr.Error(), "duplicate key") {
		recovered, findErr := c.adminRepo.FindAdminByPhone(ctx, details.Mobile)
		if findErr == nil && recovered.ID != "" {
			return recovered.ID, nil
		}
	}

	return "", utils.PrependMessageToError(saveErr, "failed to create seller account")
}

// issueOtpSession generates a 6-digit OTP, stores only its hash in a new OTP
// session keyed by a fresh otp_id, sends the OTP via SMS, and returns the otp_id.
func (c *adminUseCase) issueOtpSession(ctx context.Context, adminID, phone string) (string, error) {
	generatedOTP, err := c.otpService.GenerateOTP()
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to generate otp")
	}
	otpHash, err := c.otpService.HashOTP(generatedOTP)
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to hash otp")
	}

	otpID := uuid.NewString()
	otpSession := domain.OtpSession{
		OtpID:    otpID,
		OtpHash:  otpHash,
		UserID:   adminID,
		AdminID:  adminID,
		Phone:    phone,
		UserType: domain.UserTypeAdmin, // sellers are admins; satisfies otp_sessions user_type constraint
		ExpireAt: c.otpService.CalculateOTPExpiry(),
	}

	if err = c.authRepo.SaveOtpSession(ctx, otpSession); err != nil {
		return "", utils.PrependMessageToError(err, "failed to save otp session")
	}

	// Send the OTP via the 2factor.in SMS API (skipped when SKIP_OTP_VALIDATION=true)
	if !c.skipOTPValidation {
		if err = c.smsService.SendOTPSMS(phone, generatedOTP); err != nil {
			return "", utils.PrependMessageToError(err, "failed to send otp")
		}
	} else {
		log.Printf("[SendOTP admin] skipOTPValidation=true, not sending SMS, otp=%s phone=%s", generatedOTP, phone)
	}

	applogger.L().Info("OTP session issued", zap.String("event", applogger.EventSecurityOTPSent), zap.String("otp_id", otpID), zap.String("admin_id", adminID))
	return otpID, nil
}

func (c *adminUseCase) GetAdminWithShopVerificationByPhone(ctx context.Context, phone string) (domain.Admin, domain.ShopVerification, error) {
	return c.adminRepo.FindAdminWithShopVerificationByPhone(ctx, phone)
}

func (c *adminUseCase) AdminSignUpOtpVerify(ctx context.Context,
	otpVerifyDetails request.OTPVerify) (userID string, shop domain.ShopDetails, err error) {
	fmt.Printf("Starting OTP verification for OTP ID: %s\n", otpVerifyDetails.OtpID)

	// An otp_id (session id) is required to know which OTP to verify against
	if otpVerifyDetails.OtpID == "" {
		return "", domain.ShopDetails{}, ErrInvalidOtp
	}

	otpSession, err := c.authRepo.FindOtpSession(ctx, otpVerifyDetails.OtpID)
	if err != nil {
		return "", domain.ShopDetails{}, utils.PrependMessageToError(err, "failed to find otp session from database")
	}

	// No session found for this otp_id (zero-value struct) — reject
	if otpSession.OtpID == "" || otpSession.OtpHash == "" {
		return "", domain.ShopDetails{}, ErrInvalidOtp
	}

	// Reject expired OTP sessions
	if c.otpService.IsOTPExpired(otpSession.ExpireAt) {
		return "", domain.ShopDetails{}, ErrOtpExpired
	}

	// Verify the entered OTP against the stored hash (skip when SKIP_OTP_VALIDATION=true)
	log.Printf("[VerifyOTP admin] skipOTPValidation=%v otp_id=%s", c.skipOTPValidation, otpVerifyDetails.OtpID)
	if !c.skipOTPValidation {
		if err := c.otpService.VerifyOTP(otpVerifyDetails.Otp, otpSession.OtpHash); err != nil {
			return "", domain.ShopDetails{}, ErrInvalidOtp
		}
	}

	// OTP verified — invalidate the session so the same OTP cannot be reused.
	if delErr := c.authRepo.DeleteOtpSession(ctx, otpSession.OtpID); delErr != nil {
		// Non-fatal: log and continue; the session will still expire on its own.
		log.Printf("warning: failed to delete used otp session %s: %v", otpSession.OtpID, delErr)
	}

	// Try to get existing admin by phone
	admin, err := c.adminRepo.FindAdminByPhone(ctx, otpSession.Phone)
	if err == nil && admin.ID != "" {
		// Admin exists, try to get their shop
		shop, err := c.adminRepo.GetShopByOwnerID(ctx, admin.ID)
		if err == nil {
			return admin.ID, shop, nil
		}
		// Shop doesn't exist, return admin without shop
		return admin.ID, domain.ShopDetails{}, nil
	}

	// Admin doesn't exist, create new admin
	newAdmin := domain.Admin{
		Mobile:         otpSession.Phone,
		Status:         "active",
		VerifiedSeller: false,
		Role:           domain.AdminRoleSeller,
	}

	savedAdmin, err := c.adminRepo.SaveAdmin(ctx, newAdmin)
	if err != nil {
		return "", domain.ShopDetails{}, utils.PrependMessageToError(err, "failed to register admin")
	}

	if savedAdmin.ID == "" {
		return "", domain.ShopDetails{}, fmt.Errorf("failed to create admin: admin ID is empty")
	}

	return savedAdmin.ID, domain.ShopDetails{}, nil
}
func (c *adminUseCase) GenerateAccessToken(ctx context.Context, tokenParams service.GenerateTokenParams) (string, error) {
	fmt.Printf("Generating access token for userID: %s, userType: %s\n", tokenParams.UserID, tokenParams.UserType)
	tokenReq := token.GenerateTokenRequest{
		UserID:   tokenParams.UserID,
		UsedFor:  tokenParams.UserType,
		ExpireAt: time.Now().Add(AccessTokenDuration),
	}

	tokenRes, err := c.tokenService.GenerateToken(tokenReq)
	fmt.Printf("Token generation result for userID: %s, userType: %s, tokenRes: %+v, error: %v\n", tokenParams.UserID, tokenParams.UserType, tokenRes, err)
	if err != nil {
		return "", fmt.Errorf("failed to generate access token \nerror:%w", err)
	}
	return tokenRes.TokenString, err
}
func (c *adminUseCase) GenerateRefreshToken(ctx context.Context, tokenParams service.GenerateTokenParams) (string, error) {

	expireAt := time.Now().Add(RefreshTokenDuration)
	tokenReq := token.GenerateTokenRequest{
		UserID:   tokenParams.UserID,
		UsedFor:  tokenParams.UserType,
		ExpireAt: expireAt,
	}
	tokenRes, err := c.tokenService.GenerateToken(tokenReq)
	if err != nil {
		return "", err
	}

	err = c.authRepo.SaveRefreshSession(ctx, request.RefreshSession{
		UserID:       tokenParams.UserID,
		UserType:     string(tokenParams.UserType),
		TokenID:      tokenRes.TokenID,
		RefreshToken: utils.HashRefreshToken(tokenRes.TokenString),
		ExpireAt:     expireAt.Format(time.RFC3339),
	})
	if err != nil {
		return "", err
	}
	log.Printf("successfully refresh token created and refresh session stored in database")
	return tokenRes.TokenString, nil
}

func (c *adminUseCase) FindAllUser(ctx context.Context, pagination request.Pagination) (users []response.User, err error) {

	users, err = c.adminRepo.FindAllUser(ctx, pagination)

	return users, err
}

// Block User
func (c *adminUseCase) BlockOrUnBlockUser(ctx context.Context, blockDetails request.BlockUser) error {

	userToBlock, err := c.userRepo.FindUserByUserID(ctx, blockDetails.UserID)
	if err != nil {
		return fmt.Errorf("failed to find user \nerror:%w", err)
	}

	if userToBlock.BlockStatus == blockDetails.Block {
		return ErrSameBlockStatus
	}

	err = c.userRepo.UpdateBlockStatus(ctx, blockDetails.UserID, blockDetails.Block)
	if err != nil {
		return fmt.Errorf("failed to update user block status \nerror:%v", err.Error())
	}
	return nil
}

func (c *adminUseCase) GetFullSalesReport(ctx context.Context, requestData request.SalesReport) (salesReport []response.SalesReport, err error) {
	salesReport, err = c.adminRepo.CreateFullSalesReport(ctx, requestData)

	if err != nil {
		return salesReport, err
	}

	log.Printf("successfully got sales report from %v to %v of limit %v",
		requestData.StartDate, requestData.EndDate, requestData.Pagination.Limit)

	return salesReport, nil
}

func (c *adminUseCase) VerifyShop(ctx context.Context, verify request.ShopVerification) error {
	VerificationStatus := false
	if verify.Photo_Shop_Verification && verify.Business_Doc_Verification &&
		verify.Identity_Doc_Verification && verify.Address_Proof_Verification {
		VerificationStatus = true
	}
	err := c.adminRepo.VerifyShop(ctx, verify, VerificationStatus)

	if err != nil {
		return fmt.Errorf("failed to update shop verification status \nerror:%v", err.Error())
	}
	return nil
}

func (c *adminUseCase) CreateAdvertisement(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error) {
	createdAd, err := c.adminRepo.CreateAdvertisement(ctx, ad)
	if err != nil {
		return domain.Advertisement{}, fmt.Errorf("failed to create advertisement \nerror:%v", err.Error())
	}
	return createdAd, nil
}

func (c *adminUseCase) GetAllAdvertisements(ctx context.Context, pagination request.Pagination, filter domain.AdvertisementFilter) (ads []domain.Advertisement, err error) {
	ads, err = c.adminRepo.GetAllAdvertisements(ctx, pagination, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get all advertisements \nerror:%v", err.Error())
	}
	return ads, nil
}

func (c *adminUseCase) UpdateAdvertisement(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error) {
	updatedAd, err := c.adminRepo.UpdateAdvertisement(ctx, ad)
	if err != nil {
		return domain.Advertisement{}, fmt.Errorf("failed to update advertisement \nerror:%v", err.Error())
	}
	return updatedAd, nil
}

func (c *adminUseCase) DeleteAdvertisement(ctx context.Context, advertisementID string) error {
	err := c.adminRepo.DeleteAdvertisement(ctx, advertisementID)
	if err != nil {
		return fmt.Errorf("failed to delete advertisement \nerror:%v", err.Error())
	}
	return nil
}

func (c *adminUseCase) GetAdvertisementByID(ctx context.Context, advertisementID string) (domain.Advertisement, error) {
	ad, err := c.adminRepo.GetAdvertisementByID(ctx, advertisementID)
	if err != nil {
		return domain.Advertisement{}, fmt.Errorf("failed to get advertisement: %v", err)
	}
	return ad, nil
}

func (c *adminUseCase) GetActiveAdvertisements(ctx context.Context) ([]domain.Advertisement, error) {
	ads, err := c.adminRepo.GetActiveAdvertisements(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active advertisements: %v", err)
	}
	return ads, nil
}

func (c *adminUseCase) GetActiveAdvertisementsFiltered(ctx context.Context, filter domain.AdvertisementFilter) ([]domain.Advertisement, error) {
	ads, err := c.adminRepo.GetActiveAdvertisementsFiltered(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("failed to get filtered active advertisements: %v", err)
	}
	return ads, nil
}

// Advertisement Requests (seller-raised)

// advertisementRatePlans defines the quoted per-day rates (in paise) offered
// to sellers. Price is always recomputed server-side from these rates.
var advertisementRatePlans = []struct {
	Key         string
	Name        string
	Description string
	RatePerDay  int64
}{
	{Key: "high", Name: "Premium", Description: "Top banner slot with highest visibility", RatePerDay: 50000},
	{Key: "medium", Name: "Standard", Description: "Regular rotation in the banner carousel", RatePerDay: 30000},
	{Key: "low", Name: "Basic", Description: "Shown when premium slots are free", RatePerDay: 15000},
}

func advertisementDays(startDate, endDate time.Time) (int, error) {
	if startDate.IsZero() || endDate.IsZero() {
		return 0, fmt.Errorf("start_date and end_date are required")
	}
	if endDate.Before(startDate) {
		return 0, fmt.Errorf("end_date must not be before start_date")
	}
	// Inclusive day count: same-day campaign = 1 day.
	return int(endDate.Sub(startDate).Hours()/24) + 1, nil
}

func (c *adminUseCase) GetAdvertisementPricePlans(ctx context.Context, startDate, endDate time.Time) ([]domain.AdvertisementPricePlan, error) {
	days, err := advertisementDays(startDate, endDate)
	if err != nil {
		return nil, err
	}
	plans := make([]domain.AdvertisementPricePlan, 0, len(advertisementRatePlans))
	for _, p := range advertisementRatePlans {
		plans = append(plans, domain.AdvertisementPricePlan{
			PlanKey:     p.Key,
			Name:        p.Name,
			Description: p.Description,
			Days:        days,
			RatePerDay:  p.RatePerDay,
			TotalMinor:  p.RatePerDay * int64(days),
		})
	}
	return plans, nil
}

func (c *adminUseCase) CreateAdvertisementRequest(ctx context.Context, req domain.AdvertisementRequest) (domain.AdvertisementRequest, error) {
	days, err := advertisementDays(req.StartDate, req.EndDate)
	if err != nil {
		return domain.AdvertisementRequest{}, err
	}

	// Recompute the price server-side from the selected plan; never trust the
	// client-supplied amount.
	var rate int64
	for _, p := range advertisementRatePlans {
		if p.Key == req.PlanKey {
			rate = p.RatePerDay
			break
		}
	}
	if rate == 0 {
		return domain.AdvertisementRequest{}, fmt.Errorf("invalid plan_key: %s", req.PlanKey)
	}
	req.PriceMinor = rate * int64(days)
	req.Status = domain.AdvertRequestStatusPending

	// Attach the seller's shop when one exists.
	if req.ShopID == "" && req.AdminID != "" {
		if shop, shopErr := c.adminRepo.GetShopByOwnerID(ctx, req.AdminID); shopErr == nil && shop.ID != "" {
			req.ShopID = shop.ID
		}
	}

	created, err := c.adminRepo.CreateAdvertisementRequest(ctx, req)
	if err != nil {
		return domain.AdvertisementRequest{}, fmt.Errorf("failed to create advertisement request: %v", err)
	}
	return created, nil
}

func (c *adminUseCase) GetAllAdvertisementRequests(ctx context.Context, pagination request.Pagination, adminID string) ([]domain.AdvertisementRequest, error) {
	return c.adminRepo.GetAllAdvertisementRequests(ctx, pagination, adminID)
}

func (c *adminUseCase) GetAdvertisementRequestByID(ctx context.Context, requestID string) (domain.AdvertisementRequest, error) {
	return c.adminRepo.GetAdvertisementRequestByID(ctx, requestID)
}

func (c *adminUseCase) UpdateAdvertisementRequest(ctx context.Context, req domain.AdvertisementRequest) (domain.AdvertisementRequest, error) {
	if req.Status != domain.AdvertRequestStatusPending &&
		req.Status != domain.AdvertRequestStatusApproved &&
		req.Status != domain.AdvertRequestStatusRejected {
		return domain.AdvertisementRequest{}, fmt.Errorf("invalid status: %s", req.Status)
	}
	return c.adminRepo.UpdateAdvertisementRequest(ctx, req)
}

func (c *adminUseCase) DeleteAdvertisementRequest(ctx context.Context, requestID string) error {
	return c.adminRepo.DeleteAdvertisementRequest(ctx, requestID)
}

// Advertisement request payments

const advertPaymentGraceMinutes = 30 // order validity window, mirrors subscriptions

// advertisementPricingConfig returns the admin-configured charge percentages,
// falling back to sane defaults if the config row is missing.
func (c *adminUseCase) advertisementPricingConfig(ctx context.Context) domain.AdvertisementPricingConfig {
	cfg, err := c.adminRepo.GetAdvertisementPricingConfig(ctx)
	if err != nil {
		return domain.AdvertisementPricingConfig{GSTRatePercent: 18, PlatformFeePercent: 2}
	}
	return cfg
}

// GetAdvertisementRequestInvoice returns the full charge breakdown (base,
// platform fee, GST, total) for an approved request. All rates come from the
// admin-managed pricing configuration — nothing is hardcoded.
func (c *adminUseCase) GetAdvertisementRequestInvoice(ctx context.Context, requestID string) (domain.AdvertisementInvoice, error) {
	req, err := c.adminRepo.GetAdvertisementRequestByID(ctx, requestID)
	if err != nil {
		return domain.AdvertisementInvoice{}, err
	}
	if req.Status != domain.AdvertRequestStatusApproved {
		return domain.AdvertisementInvoice{}, fmt.Errorf("invoice available only for approved requests (current status: %s)", req.Status)
	}

	days, err := advertisementDays(req.StartDate, req.EndDate)
	if err != nil {
		return domain.AdvertisementInvoice{}, err
	}

	// The base amount is the price locked in at request time; the plan lookup
	// only decorates the invoice with the display name and per-day rate.
	planName := req.PlanKey
	var ratePerDay int64
	if plan, planErr := c.adminRepo.GetAdvertisementPlanByKey(ctx, req.PlanKey); planErr == nil {
		planName = plan.Name
		ratePerDay = plan.RatePerDayMinor
	} else if days > 0 {
		ratePerDay = req.PriceMinor / int64(days)
	}

	pricing := c.advertisementPricingConfig(ctx)

	base := req.PriceMinor
	platformFee := int64(float64(base) * pricing.PlatformFeePercent / 100)
	gst := int64(float64(base+platformFee) * pricing.GSTRatePercent / 100)

	return domain.AdvertisementInvoice{
		RequestID:        req.ID,
		PlanKey:          req.PlanKey,
		PlanName:         planName,
		Days:             days,
		RatePerDay:       ratePerDay,
		BaseMinor:        base,
		PlatformFeeMinor: platformFee,
		GSTRatePercent:   pricing.GSTRatePercent,
		GSTMinor:         gst,
		TotalMinor:       base + platformFee + gst,
	}, nil
}

// Advertisement pricing management (admin panel)

func (c *adminUseCase) ListAdvertisementPlanConfigs(ctx context.Context) ([]domain.AdvertisementPlanConfig, error) {
	return c.adminRepo.ListAdvertisementPlans(ctx, false)
}

func validateAdvertisementPlan(plan domain.AdvertisementPlanConfig) error {
	if strings.TrimSpace(plan.PlanKey) == "" {
		return fmt.Errorf("plan_key is required")
	}
	if strings.TrimSpace(plan.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if plan.RatePerDayMinor <= 0 {
		return fmt.Errorf("rate_per_day_minor must be greater than zero")
	}
	return nil
}

func (c *adminUseCase) CreateAdvertisementPlanConfig(ctx context.Context, plan domain.AdvertisementPlanConfig) (domain.AdvertisementPlanConfig, error) {
	if err := validateAdvertisementPlan(plan); err != nil {
		return domain.AdvertisementPlanConfig{}, err
	}
	return c.adminRepo.CreateAdvertisementPlan(ctx, plan)
}

func (c *adminUseCase) UpdateAdvertisementPlanConfig(ctx context.Context, plan domain.AdvertisementPlanConfig) (domain.AdvertisementPlanConfig, error) {
	if plan.ID == "" {
		return domain.AdvertisementPlanConfig{}, fmt.Errorf("plan id is required")
	}
	if err := validateAdvertisementPlan(plan); err != nil {
		return domain.AdvertisementPlanConfig{}, err
	}
	return c.adminRepo.UpdateAdvertisementPlan(ctx, plan)
}

func (c *adminUseCase) DeleteAdvertisementPlanConfig(ctx context.Context, planID string) error {
	return c.adminRepo.DeleteAdvertisementPlan(ctx, planID)
}

func (c *adminUseCase) GetAdvertisementPricingConfig(ctx context.Context) (domain.AdvertisementPricingConfig, error) {
	return c.advertisementPricingConfig(ctx), nil
}

func (c *adminUseCase) UpdateAdvertisementPricingConfig(ctx context.Context, cfg domain.AdvertisementPricingConfig) (domain.AdvertisementPricingConfig, error) {
	if cfg.GSTRatePercent < 0 || cfg.GSTRatePercent > 100 {
		return domain.AdvertisementPricingConfig{}, fmt.Errorf("gst_rate_percent must be between 0 and 100")
	}
	if cfg.PlatformFeePercent < 0 || cfg.PlatformFeePercent > 100 {
		return domain.AdvertisementPricingConfig{}, fmt.Errorf("platform_fee_percent must be between 0 and 100")
	}
	return c.adminRepo.UpdateAdvertisementPricingConfig(ctx, cfg)
}

// ── Shop phone number change (OTP-gated) ────────────────────────────────────

// SendShopPhoneChangeOtp sends an OTP to newPhone and returns an otp_id to
// verify it with. Rejects a number already registered to a *different*
// admin — reusing the same number the caller already has is allowed (no-op
// once verified) but shouldn't need this flow at all.
func (c *adminUseCase) SendShopPhoneChangeOtp(ctx context.Context, callerID, newPhone string) (string, error) {
	if strings.TrimSpace(newPhone) == "" {
		return "", domain.ValidationError("phone", "phone number is required")
	}

	existing, err := c.adminRepo.FindAdminByPhone(ctx, newPhone)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", domain.InternalError("failed to check phone number", err)
	}
	if existing.ID != "" && existing.ID != callerID {
		return "", domain.ConflictError("phone number already in use", "this phone number is registered to another account")
	}

	return c.issueOtpSession(ctx, callerID, newPhone)
}

// VerifyShopPhoneChangeOtp verifies the OTP issued by SendShopPhoneChangeOtp
// and, only on success, updates admins.mobile and shop_details.phone. The
// session's own AdminID/Phone are trusted over caller-supplied values — a
// verified session can only ever apply to the admin and number it was
// issued for, so a stolen otp_id can't be replayed against a different
// caller or a different phone number.
func (c *adminUseCase) VerifyShopPhoneChangeOtp(ctx context.Context, callerID, otpID, otp string) (string, error) {
	if otpID == "" {
		return "", domain.ValidationError("otp_id", "otp_id is required")
	}

	otpSession, err := c.authRepo.FindOtpSession(ctx, otpID)
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to find otp session")
	}
	if otpSession.OtpID == "" || otpSession.OtpHash == "" {
		return "", domain.ValidationError("otp", "invalid OTP session")
	}
	if otpSession.AdminID != callerID {
		return "", domain.ForbiddenError("this OTP was not issued to your account")
	}
	if c.otpService.IsOTPExpired(otpSession.ExpireAt) {
		return "", domain.ValidationError("otp", "OTP has expired — request a new one")
	}
	if !c.skipOTPValidation {
		if err := c.otpService.VerifyOTP(otp, otpSession.OtpHash); err != nil {
			return "", domain.ValidationError("otp", "incorrect OTP")
		}
	}

	if err := c.adminRepo.UpdateAdminMobile(ctx, callerID, otpSession.Phone); err != nil {
		return "", domain.InternalError("failed to update phone number", err)
	}
	if err := c.adminRepo.UpdateShopPhoneByAdminID(ctx, callerID, otpSession.Phone); err != nil {
		return "", domain.InternalError("failed to update shop phone number", err)
	}

	return otpSession.Phone, nil
}

// ── Roles & permissions ─────────────────────────────────────────────────────

func (c *adminUseCase) requireSuperAdmin(ctx context.Context, callerID string) (domain.Admin, error) {
	caller, err := c.adminRepo.GetAdminByID(ctx, callerID)
	if err != nil || caller.ID == "" {
		return domain.Admin{}, domain.NotFoundError("caller admin")
	}
	if caller.Role != domain.AdminRoleSuperAdmin {
		return domain.Admin{}, domain.ForbiddenError("only super admins can manage roles")
	}
	return caller, nil
}

func validatePermissionKeys(keys []domain.PermissionKey) error {
	for _, k := range keys {
		if !k.IsValid() {
			return domain.ValidationError("permissions", fmt.Sprintf("unknown permission key: %s", k))
		}
	}
	return nil
}

func (c *adminUseCase) CreateRole(ctx context.Context, callerID string, role domain.Role) (domain.Role, error) {
	if _, err := c.requireSuperAdmin(ctx, callerID); err != nil {
		return domain.Role{}, err
	}

	name := strings.TrimSpace(role.Name)
	if name == "" || role.Label == "" {
		return domain.Role{}, domain.ValidationError("name/label", "both are required")
	}
	if name == domain.RoleNameSeller || name == domain.RoleNameSuperAdmin {
		return domain.Role{}, domain.ValidationError("name", "this role name is reserved")
	}
	if err := validatePermissionKeys(role.Permissions); err != nil {
		return domain.Role{}, err
	}

	existing, err := c.adminRepo.GetRoleByName(ctx, name)
	if err != nil {
		return domain.Role{}, err
	}
	if existing.ID != "" {
		return domain.Role{}, domain.ConflictError("role already exists", "a role with this name already exists")
	}

	return c.adminRepo.CreateRole(ctx, domain.Role{Name: name, Label: role.Label}, role.Permissions)
}

func (c *adminUseCase) ListRoles(ctx context.Context, callerID string) ([]domain.Role, error) {
	if _, err := c.requireSuperAdmin(ctx, callerID); err != nil {
		return nil, err
	}
	return c.adminRepo.ListRoles(ctx)
}

func (c *adminUseCase) GetRole(ctx context.Context, callerID, roleID string) (domain.Role, error) {
	if _, err := c.requireSuperAdmin(ctx, callerID); err != nil {
		return domain.Role{}, err
	}
	role, err := c.adminRepo.GetRoleByID(ctx, roleID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Role{}, domain.NotFoundError("role")
	}
	return role, err
}

func (c *adminUseCase) UpdateRole(ctx context.Context, callerID, roleID string, body domain.Role) (domain.Role, error) {
	if _, err := c.requireSuperAdmin(ctx, callerID); err != nil {
		return domain.Role{}, err
	}

	existing, err := c.adminRepo.GetRoleByID(ctx, roleID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.Role{}, domain.NotFoundError("role")
	}
	if err != nil {
		return domain.Role{}, err
	}
	if existing.Name == domain.RoleNameSuperAdmin {
		return domain.Role{}, domain.ForbiddenError("super_admin permissions cannot be edited — it always has full access")
	}

	var labelPtr *string
	if body.Label != "" {
		labelPtr = &body.Label
	}

	var permsPtr *[]domain.PermissionKey
	if body.Permissions != nil {
		if err := validatePermissionKeys(body.Permissions); err != nil {
			return domain.Role{}, err
		}
		permsPtr = &body.Permissions
	}

	if labelPtr == nil && permsPtr == nil {
		return domain.Role{}, domain.ValidationError("body", "no updatable fields provided")
	}

	return c.adminRepo.UpdateRole(ctx, roleID, labelPtr, permsPtr)
}

func (c *adminUseCase) DeleteRole(ctx context.Context, callerID, roleID string) error {
	if _, err := c.requireSuperAdmin(ctx, callerID); err != nil {
		return err
	}

	role, err := c.adminRepo.GetRoleByID(ctx, roleID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return domain.NotFoundError("role")
	}
	if err != nil {
		return err
	}
	if role.IsSystem {
		return domain.ForbiddenError("built-in roles can't be deleted")
	}

	inUse, err := c.adminRepo.CountAdminsWithRole(ctx, role.Name)
	if err != nil {
		return err
	}
	if inUse > 0 {
		return domain.ConflictError("role is in use", fmt.Sprintf("%d platform user(s) currently have this role; reassign them first", inUse))
	}

	if err := c.adminRepo.DeleteRole(ctx, roleID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.NotFoundError("role")
		}
		return err
	}
	return nil
}

func (c *adminUseCase) GetRoleByName(ctx context.Context, name string) (domain.Role, error) {
	return c.adminRepo.GetRoleByName(ctx, name)
}

func (c *adminUseCase) GetPermissionsForRole(ctx context.Context, roleName string) ([]domain.PermissionKey, error) {
	if roleName == domain.RoleNameSuperAdmin || roleName == "" {
		return domain.AllPermissionKeys(), nil
	}
	role, err := c.adminRepo.GetRoleByName(ctx, roleName)
	if err != nil {
		return nil, err
	}
	if role.ID == "" {
		// Role no longer exists (deleted out from under an assigned admin) —
		// fail closed with no permissions rather than erroring the caller out.
		return []domain.PermissionKey{}, nil
	}
	return role.Permissions, nil
}

func (c *adminUseCase) GetSubscriptionGSTConfig(ctx context.Context) (domain.SubscriptionGSTConfig, error) {
	cfg, err := c.adminRepo.GetSubscriptionGSTConfig(ctx)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Same fallback as SubscriptionRepository.GetGSTRateBasisPoints —
		// the migration seeds this row, but never hard-fail a read.
		return domain.SubscriptionGSTConfig{GSTRateBasisPoints: 1800}, nil
	}
	return cfg, err
}

func (c *adminUseCase) UpdateSubscriptionGSTConfig(ctx context.Context, cfg domain.SubscriptionGSTConfig) (domain.SubscriptionGSTConfig, error) {
	if cfg.GSTRateBasisPoints < 0 || cfg.GSTRateBasisPoints > 10000 {
		return domain.SubscriptionGSTConfig{}, domain.ValidationError("gst_rate_basis_points", "must be between 0 and 10000 (0% to 100%)")
	}
	return c.adminRepo.UpdateSubscriptionGSTConfig(ctx, cfg)
}

// Feature flags

// GetFeatureFlagsObject returns all flags as a key→enabled map, the shape
// clients consume: {"advertisement": false, ...}.
func (c *adminUseCase) GetFeatureFlagsObject(ctx context.Context) (map[string]bool, error) {
	flags, err := c.adminRepo.ListFeatureFlags(ctx)
	if err != nil {
		return nil, err
	}
	obj := make(map[string]bool, len(flags))
	for _, f := range flags {
		obj[f.FlagKey] = f.Enabled
	}
	return obj, nil
}

func (c *adminUseCase) ListFeatureFlags(ctx context.Context) ([]domain.FeatureFlag, error) {
	return c.adminRepo.ListFeatureFlags(ctx)
}

func (c *adminUseCase) CreateFeatureFlag(ctx context.Context, flag domain.FeatureFlag) (domain.FeatureFlag, error) {
	flag.FlagKey = strings.TrimSpace(strings.ToLower(flag.FlagKey))
	if flag.FlagKey == "" {
		return domain.FeatureFlag{}, fmt.Errorf("flag_key is required")
	}
	return c.adminRepo.CreateFeatureFlag(ctx, flag)
}

func (c *adminUseCase) UpdateFeatureFlag(ctx context.Context, flag domain.FeatureFlag) (domain.FeatureFlag, error) {
	if flag.ID == "" {
		return domain.FeatureFlag{}, fmt.Errorf("flag id is required")
	}
	flag.FlagKey = strings.TrimSpace(strings.ToLower(flag.FlagKey))
	if flag.FlagKey == "" {
		return domain.FeatureFlag{}, fmt.Errorf("flag_key is required")
	}
	return c.adminRepo.UpdateFeatureFlag(ctx, flag)
}

func (c *adminUseCase) DeleteFeatureFlag(ctx context.Context, flagID string) error {
	return c.adminRepo.DeleteFeatureFlag(ctx, flagID)
}

// App configs

func normalizeAppConfigKey(key string) string {
	key = strings.TrimSpace(strings.ToLower(key))
	key = strings.ReplaceAll(key, " ", "_")
	key = strings.ReplaceAll(key, "-", "_")
	return key
}

func validateAppConfig(cfg domain.AppConfig) error {
	cfg.ConfigKey = normalizeAppConfigKey(cfg.ConfigKey)
	if cfg.ConfigKey == "" {
		return fmt.Errorf("config_key is required")
	}
	if cfg.Value == "" {
		return fmt.Errorf("value is required")
	}
	return nil
}

func (c *adminUseCase) ListAppConfigs(ctx context.Context) ([]domain.AppConfig, error) {
	return c.adminRepo.ListAppConfigs(ctx)
}

func (c *adminUseCase) GetAppConfigByKey(ctx context.Context, configKey string) (domain.AppConfig, error) {
	return c.adminRepo.GetAppConfigByKey(ctx, normalizeAppConfigKey(configKey))
}

func (c *adminUseCase) CreateAppConfig(ctx context.Context, cfg domain.AppConfig) (domain.AppConfig, error) {
	if err := validateAppConfig(cfg); err != nil {
		return domain.AppConfig{}, err
	}
	cfg.ConfigKey = normalizeAppConfigKey(cfg.ConfigKey)
	return c.adminRepo.CreateAppConfig(ctx, cfg)
}

func (c *adminUseCase) UpdateAppConfig(ctx context.Context, cfg domain.AppConfig) (domain.AppConfig, error) {
	if cfg.ID == "" {
		return domain.AppConfig{}, fmt.Errorf("config id is required")
	}
	if err := validateAppConfig(cfg); err != nil {
		return domain.AppConfig{}, err
	}
	cfg.ConfigKey = normalizeAppConfigKey(cfg.ConfigKey)
	return c.adminRepo.UpdateAppConfig(ctx, cfg)
}

func (c *adminUseCase) DeleteAppConfig(ctx context.Context, configID string) error {
	return c.adminRepo.DeleteAppConfig(ctx, configID)
}

// GetGlobalConfig returns the structured, admin-editable app config used by
// client apps. Falls back to DefaultGlobalAppConfig when nothing has been
// saved yet, so the portal always has something sensible to show.
func (c *adminUseCase) GetGlobalConfig(ctx context.Context) (domain.GlobalAppConfig, error) {
	row, err := c.adminRepo.GetAppConfigByKey(ctx, domain.GlobalAppConfigKey)
	if err != nil {
		return domain.DefaultGlobalAppConfig(), nil
	}
	var cfg domain.GlobalAppConfig
	if err := json.Unmarshal([]byte(row.Value), &cfg); err != nil {
		return domain.DefaultGlobalAppConfig(), fmt.Errorf("stored global config is corrupt: %w", err)
	}
	return cfg, nil
}

// UpdateGlobalConfig persists the full structured config as a single JSON
// blob, creating the underlying app_configs row on first save.
func (c *adminUseCase) UpdateGlobalConfig(ctx context.Context, cfg domain.GlobalAppConfig) (domain.GlobalAppConfig, error) {
	encoded, err := json.Marshal(cfg)
	if err != nil {
		return domain.GlobalAppConfig{}, fmt.Errorf("failed to encode config: %w", err)
	}

	row := domain.AppConfig{
		ConfigKey:   domain.GlobalAppConfigKey,
		Value:       string(encoded),
		Description: "Global app config (CDN, feature flags, HTTP, image upload, AI) — managed via admin-portal Config page",
		Enabled:     true,
	}

	existing, err := c.adminRepo.GetAppConfigByKey(ctx, domain.GlobalAppConfigKey)
	if err != nil {
		if _, err := c.adminRepo.CreateAppConfig(ctx, row); err != nil {
			return domain.GlobalAppConfig{}, err
		}
		return cfg, nil
	}

	row.ID = existing.ID
	if _, err := c.adminRepo.UpdateAppConfig(ctx, row); err != nil {
		return domain.GlobalAppConfig{}, err
	}
	return cfg, nil
}

// Help center (contact settings + FAQs)

func validateHelpSettings(s domain.HelpSettings) error {
	if strings.TrimSpace(s.SupportPhone) == "" &&
		strings.TrimSpace(s.SupportEmail) == "" &&
		strings.TrimSpace(s.WhatsAppNumber) == "" {
		return fmt.Errorf("at least one contact method (support_phone, support_email, or whatsapp_number) is required")
	}
	if s.SupportEmail != "" && !strings.Contains(s.SupportEmail, "@") {
		return fmt.Errorf("support_email is not a valid email address")
	}
	return nil
}

func validateHelpFAQ(f domain.HelpFAQ) error {
	if strings.TrimSpace(f.Question) == "" {
		return fmt.Errorf("question is required")
	}
	if strings.TrimSpace(f.Answer) == "" {
		return fmt.Errorf("answer is required")
	}
	return nil
}

func (c *adminUseCase) GetHelpSettings(ctx context.Context) (domain.HelpSettings, error) {
	return c.adminRepo.GetHelpSettings(ctx)
}

func (c *adminUseCase) UpdateHelpSettings(ctx context.Context, settings domain.HelpSettings) (domain.HelpSettings, error) {
	if err := validateHelpSettings(settings); err != nil {
		return domain.HelpSettings{}, err
	}
	return c.adminRepo.UpsertHelpSettings(ctx, settings)
}

func (c *adminUseCase) ListHelpFAQs(ctx context.Context) ([]domain.HelpFAQ, error) {
	return c.adminRepo.ListHelpFAQs(ctx)
}

func (c *adminUseCase) CreateHelpFAQ(ctx context.Context, faq domain.HelpFAQ) (domain.HelpFAQ, error) {
	if err := validateHelpFAQ(faq); err != nil {
		return domain.HelpFAQ{}, err
	}
	return c.adminRepo.CreateHelpFAQ(ctx, faq)
}

func (c *adminUseCase) UpdateHelpFAQ(ctx context.Context, faq domain.HelpFAQ) (domain.HelpFAQ, error) {
	if faq.ID == "" {
		return domain.HelpFAQ{}, fmt.Errorf("faq id is required")
	}
	if err := validateHelpFAQ(faq); err != nil {
		return domain.HelpFAQ{}, err
	}
	return c.adminRepo.UpdateHelpFAQ(ctx, faq)
}

func (c *adminUseCase) DeleteHelpFAQ(ctx context.Context, faqID string) error {
	return c.adminRepo.DeleteHelpFAQ(ctx, faqID)
}

// GetHelpCenter bundles contact settings + only ACTIVE FAQs for the single
// public read endpoint consumed by client apps (seller app today; customer
// app can reuse the same endpoint later).
func (c *adminUseCase) GetHelpCenter(ctx context.Context) (domain.HelpCenter, error) {
	settings, err := c.adminRepo.GetHelpSettings(ctx)
	if err != nil {
		return domain.HelpCenter{}, err
	}
	faqs, err := c.adminRepo.ListActiveHelpFAQs(ctx)
	if err != nil {
		return domain.HelpCenter{}, err
	}
	if faqs == nil {
		faqs = []domain.HelpFAQ{}
	}
	return domain.HelpCenter{Settings: settings, FAQs: faqs}, nil
}

// Subscription plans (admin panel)

func validateSubscriptionPlan(plan domain.SubscriptionPlan) error {
	if strings.TrimSpace(plan.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if plan.PriceMonthly.AmountMinor < 0 {
		return fmt.Errorf("price must not be negative")
	}
	if plan.DurationDays == 0 {
		return fmt.Errorf("duration_days must be greater than zero")
	}
	return nil
}

func (c *adminUseCase) ListSubscriptionPlans(ctx context.Context) ([]domain.SubscriptionPlan, error) {
	return c.adminRepo.ListSubscriptionPlans(ctx)
}

func (c *adminUseCase) CreateSubscriptionPlan(ctx context.Context, plan domain.SubscriptionPlan) (domain.SubscriptionPlan, error) {
	if err := validateSubscriptionPlan(plan); err != nil {
		return domain.SubscriptionPlan{}, err
	}
	return c.adminRepo.CreateSubscriptionPlan(ctx, plan)
}

func (c *adminUseCase) UpdateSubscriptionPlan(ctx context.Context, plan domain.SubscriptionPlan) (domain.SubscriptionPlan, error) {
	if plan.ID == "" {
		return domain.SubscriptionPlan{}, fmt.Errorf("plan id is required")
	}
	if err := validateSubscriptionPlan(plan); err != nil {
		return domain.SubscriptionPlan{}, err
	}
	return c.adminRepo.UpdateSubscriptionPlan(ctx, plan)
}

func (c *adminUseCase) DeleteSubscriptionPlan(ctx context.Context, planID string) error {
	return c.adminRepo.DeleteSubscriptionPlan(ctx, planID)
}

// CreateAdvertisementPaymentOrder creates a Razorpay order for the invoice
// total of an approved, not-yet-paid request owned by adminID.
func (c *adminUseCase) CreateAdvertisementPaymentOrder(ctx context.Context, adminID, requestID string) (orderID string, keyID string, amountMinor int64, err error) {
	req, err := c.adminRepo.GetAdvertisementRequestByID(ctx, requestID)
	if err != nil {
		return "", "", 0, err
	}
	if req.AdminID != adminID {
		return "", "", 0, fmt.Errorf("request does not belong to this seller")
	}
	if req.Status != domain.AdvertRequestStatusApproved {
		return "", "", 0, fmt.Errorf("request is not approved yet")
	}
	if req.PaymentStatus == domain.AdvertPaymentPaid {
		return "", "", 0, fmt.Errorf("request is already paid")
	}

	invoice, err := c.GetAdvertisementRequestInvoice(ctx, requestID)
	if err != nil {
		return "", "", 0, err
	}

	client := razorpay.NewClient(c.config.RazorPayKey, c.config.RazorPaySecret)
	receipt := fmt.Sprintf("advr_%s_%d", req.ID[len(req.ID)-min(8, len(req.ID)):], time.Now().Unix())
	rzpRes, err := client.Order.Create(map[string]interface{}{
		"amount":   invoice.TotalMinor,
		"currency": "INR",
		"receipt":  receipt,
	}, nil)
	if err != nil {
		return "", "", 0, fmt.Errorf("razorpay order create: %w", err)
	}
	rzpOrderID, _ := rzpRes["id"].(string)
	if rzpOrderID == "" {
		return "", "", 0, fmt.Errorf("razorpay returned empty order id")
	}

	if err := c.adminRepo.SetAdvertisementRequestPaymentOrder(ctx, req.ID, rzpOrderID); err != nil {
		return "", "", 0, fmt.Errorf("persist payment order: %w", err)
	}
	return rzpOrderID, c.config.RazorPayKey, invoice.TotalMinor, nil
}

// VerifyAdvertisementPayment verifies the Razorpay signature and marks the
// request paid. Idempotent for already-paid requests.
func (c *adminUseCase) VerifyAdvertisementPayment(ctx context.Context, adminID, orderID, paymentID, signature string) error {
	// HMAC-SHA256(order_id|payment_id, secret) must equal the signature.
	mac := hmac.New(sha256.New, []byte(c.config.RazorPaySecret))
	mac.Write([]byte(orderID + "|" + paymentID))
	expected := hex.EncodeToString(mac.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) != 1 {
		return fmt.Errorf("invalid payment signature")
	}

	req, err := c.adminRepo.GetAdvertisementRequestByPaymentOrderID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("payment order not found")
	}
	if req.AdminID != adminID {
		return fmt.Errorf("payment order does not belong to this seller")
	}
	if req.PaymentStatus == domain.AdvertPaymentPaid {
		return nil // idempotent
	}
	if time.Since(req.UpdatedAt) > advertPaymentGraceMinutes*time.Minute {
		return fmt.Errorf("payment order expired, please retry")
	}

	return c.adminRepo.MarkAdvertisementRequestPaid(ctx, req.ID, paymentID)
}

// AdvertisementPaymentFailed records a failed payment attempt.
func (c *adminUseCase) AdvertisementPaymentFailed(ctx context.Context, adminID, orderID string) error {
	req, err := c.adminRepo.GetAdvertisementRequestByPaymentOrderID(ctx, orderID)
	if err != nil {
		return nil // unknown order — nothing to record
	}
	if req.AdminID != adminID {
		return nil
	}
	log.Printf("[ADVERT_PAYMENT_FAILURE] admin_id=%s request_id=%s order_id=%s", adminID, req.ID, orderID)
	return c.adminRepo.MarkAdvertisementRequestPaymentFailed(ctx, req.ID)
}

func (c *adminUseCase) CreateShop(ctx context.Context, shop domain.ShopDetails) (domain.ShopDetails, error) {
	createdShop, err := c.adminRepo.CreateShop(ctx, shop)
	if err != nil {
		return domain.ShopDetails{}, fmt.Errorf("failed to create shop \nerror:%v", err.Error())
	}

	// Referral attach is best-effort: an invalid or missing code must never
	// fail shop creation, only surface as ReferralStatus for the client to
	// show "Invalid Referral ID" (see ShopDetails.ReferralStatus).
	createdShop.ReferralStatus = c.attachShopReferral(ctx, createdShop.ID, createdShop.AdminID, shop.ReferralCouponID)
	createdShop.ReferralCouponID = shop.ReferralCouponID

	return createdShop, nil
}

// attachShopReferral resolves referralCouponID to a Sales Executive and, on
// a match, attaches shopID to them for commission/reporting. Returns
// "attached", "invalid", or "" (no code submitted) — never an error, so a
// broken or unknown referral code never blocks shop creation.
func (c *adminUseCase) attachShopReferral(ctx context.Context, shopID, sellerAdminID, referralCouponID string) string {
	if referralCouponID == "" {
		return ""
	}

	// Idempotent: onboarding retries (or a resubmit of the same step) must
	// not create a second attachment or move the shop to a different exec.
	existing, err := c.adminRepo.GetShopReferralByShopID(ctx, shopID)
	if err != nil {
		applogger.L().Warn("referral lookup failed during shop create", zap.String("shop_id", shopID), zap.Error(err))
		return "invalid"
	}
	if existing.ID != "" {
		return "attached"
	}

	exec, err := c.adminRepo.FindAdminByReferralCouponID(ctx, referralCouponID)
	if err != nil {
		applogger.L().Warn("referral admin lookup failed during shop create", zap.String("shop_id", shopID), zap.Error(err))
		return "invalid"
	}
	if exec.ID == "" {
		return "invalid"
	}

	_, err = c.adminRepo.CreateShopReferral(ctx, domain.ShopReferral{
		ReferralCouponID: referralCouponID,
		PlatformUserID:   exec.ID,
		ShopID:           shopID,
		SellerAdminID:    sellerAdminID,
		Status:           domain.ShopReferralStatusActive,
	})
	if err != nil {
		applogger.L().Error("failed to attach shop referral", zap.String("shop_id", shopID), zap.String("platform_user_id", exec.ID), zap.Error(err))
		return "invalid"
	}
	return "attached"
}

func (c *adminUseCase) SearchShops(ctx context.Context, filter request.ShopSearch) ([]domain.ShopDetails, error) {
	if filter.Limit <= 0 || filter.Limit > 200 {
		filter.Limit = 50
	}
	return c.adminRepo.SearchShops(ctx, filter)
}

func (c *adminUseCase) GetAllShops(ctx context.Context, pagination request.Pagination) (shops []domain.ShopDetails, err error) {
	shops, err = c.adminRepo.GetAllShops(ctx, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get all shops \nerror:%v", err.Error())
	}
	return shops, nil
}

func (c *adminUseCase) GetShopByID(ctx context.Context, shopID string) (shop domain.ShopDetails, err error) {
	shop, err = c.adminRepo.GetShopByID(ctx, shopID)
	if err != nil {
		return domain.ShopDetails{}, fmt.Errorf("failed to get shop by id \nerror:%v", err.Error())
	}
	return shop, nil
}
func (c *adminUseCase) UpdateShop(ctx context.Context, shop map[string]interface{}, shopId string) (map[string]interface{}, error) {
	updatedShop, err := c.adminRepo.UpdateShop(ctx, shop, shopId)
	if err != nil {
		return nil, fmt.Errorf("failed to update shop \nerror:%v", err.Error())
	}
	return updatedShop, nil
}

func (c *adminUseCase) GetShopByOwnerID(ctx context.Context, ownerID string) (shop domain.ShopDetails, err error) {
	shop, err = c.adminRepo.GetShopByOwnerID(ctx, ownerID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ShopDetails{}, nil
		}
		return domain.ShopDetails{}, fmt.Errorf("failed to get shop by owner id \nerror:%v", err.Error())
	}
	return shop, nil
}

func (c *adminUseCase) SendNotificationToUsersInRadius(ctx context.Context, requestData request.NotificationRadiusRequest) error {
	err := c.adminRepo.SendNotificationToUsersInRadius(ctx, requestData)
	if err != nil {
		return fmt.Errorf("failed to send notification to users in radius \nerror:%v", err.Error())
	}
	return nil
}

func (c *adminUseCase) SendNotificationToUser(ctx context.Context, userID string, message string) error {
	err := c.adminRepo.SendNotificationToUser(ctx, userID, message)
	if err != nil {
		return fmt.Errorf("failed to send notification to user \nerror:%v", err.Error())
	}
	return nil
}

func (c *adminUseCase) UploadAdminProfileImage(ctx context.Context, adminID string, imagePath string, shopId string) (string, error) {
	if imagePath == "" {
		return "", fmt.Errorf("invalid image path data")
	}
	uploadedImagePath, err := c.adminRepo.UploadAdminProfileImage(ctx, adminID, imagePath, shopId)
	if err != nil {
		return "", fmt.Errorf("failed to upload admin profile image \nerror:%v", err.Error())
	}
	return uploadedImagePath, nil
}

func (c *adminUseCase) DecodeTokenData(tokenString string) string {
	return c.tokenService.DecodeTokenData(tokenString)
}

func (c *adminUseCase) UploadShopDocument(ctx context.Context, shopID string, documentType string, documentValue string) error {

	err := c.adminRepo.UploadShopDocument(ctx, shopID, documentType, documentValue)
	if err != nil {
		return fmt.Errorf("failed to upload shop document \nerror:%v", err.Error())
	}
	return nil
}
func (c *adminUseCase) UploadAddress(ctx context.Context, adminId string, address request.AddressRequest) error {
	err := c.adminRepo.UploadAddress(ctx, adminId, address)
	if err != nil {
		return fmt.Errorf("failed to upload address \nerror:%v", err.Error())
	}
	return nil
}

func (c *adminUseCase) VerifyShopDocument(ctx context.Context, otp string) error {
	return nil
}

func (c *adminUseCase) UploadAdminDocumentOtpSend(ctx context.Context, adminID string, documentType string, documentValue string) error {
	err := c.adminRepo.UploadAdminDocumentOtpSend(ctx, adminID, documentType, documentValue)
	if err != nil {
		return fmt.Errorf("failed to upload admin document \nerror:%v", err.Error())
	}
	return nil
}

func (c *adminUseCase) UploadAdminDocumentOtpVerify(ctx context.Context, otp string, documentType string, documentValue string) error {
	// For demonstration, assume OTP is always valid
	// In real implementation, verify OTP against a stored value or external service
	return nil
}

func (c *adminUseCase) GetVerificationStatus(ctx context.Context, adminId string) (admin domain.Admin, shopVerification domain.ShopVerification, err error) {
	admin, shopVerification, err = c.adminRepo.GetVerificationStatus(ctx, adminId)
	if err != nil {
		return domain.Admin{}, domain.ShopVerification{}, fmt.Errorf("failed to get admin verification status \nerror:%v", err.Error())
	}

	return admin, shopVerification, nil
}

func (c *adminUseCase) GetAllProductDetails(ctx context.Context) (products []any, err error) {
	// Open and read the products.json file
	file, err := os.Open("pkg/data/products.json")
	if err != nil {
		return nil, fmt.Errorf("failed to open products.json: %w", err)
	}
	defer file.Close()

	// Read file contents
	data, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read products.json: %w", err)
	}

	// Parse JSON data
	var jsonData any
	err = json.Unmarshal(data, &jsonData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	// Convert to []any slice
	products = []any{jsonData}
	return products, nil
}

func (c *adminUseCase) GetShopProfileImageById(ctx context.Context, shopId string) (string, error) {
	shopProfileImage, err := c.adminRepo.GetShopProfileImageById(ctx, shopId)
	if err != nil {
		return "", fmt.Errorf("failed to get shop profile image by id \nerror:%v", err.Error())
	}
	return shopProfileImage, nil
}

func (c *adminUseCase) UserLogout(ctx context.Context, adminId string) error {
	err := c.adminRepo.DeleteRefreshSessionByUserID(ctx, adminId)
	if err != nil {
		return fmt.Errorf("failed to logout user \nerror:%v", err.Error())
	}
	return nil
}

func (c *adminUseCase) GetShopSocialDetails(ctx context.Context, shopID string) ([]domain.ShopSocial, error) {
	shopSocialDetails, err := c.adminRepo.GetShopSocialDetails(ctx, shopID)
	if err != nil {
		return nil, fmt.Errorf("failed to get shop social details \nerror:%v", err.Error())
	}
	return shopSocialDetails, nil
}

func (c *adminUseCase) GetAdminByID(ctx context.Context, adminID string) (domain.Admin, error) {
	return c.adminRepo.GetAdminByID(ctx, adminID)
}

func (c *adminUseCase) GetDashboardStats(ctx context.Context) (domain.DashboardStats, error) {
	return c.adminRepo.GetDashboardStats(ctx)
}

// ── Sales Executive referral program (CRUD) ────────────────────────────────

func (c *adminUseCase) FindAdminByReferralCouponID(ctx context.Context, referralCouponID string) (domain.Admin, error) {
	return c.adminRepo.FindAdminByReferralCouponID(ctx, referralCouponID)
}

func (c *adminUseCase) FindAdminByEmail(ctx context.Context, email string) (domain.Admin, error) {
	return c.adminRepo.FindAdminByEmail(ctx, email)
}

func (c *adminUseCase) CreateShopReferral(ctx context.Context, callerID string, body domain.ShopReferral) (domain.ShopReferral, error) {
	caller, err := c.adminRepo.GetAdminByID(ctx, callerID)
	if err != nil || caller.ID == "" {
		return domain.ShopReferral{}, domain.NotFoundError("caller admin")
	}
	if caller.Role != domain.AdminRoleSuperAdmin {
		return domain.ShopReferral{}, domain.ForbiddenError("only super admins can create referral attachments")
	}
	if body.ReferralCouponID == "" || body.PlatformUserID == "" || body.ShopID == "" {
		return domain.ShopReferral{}, domain.ValidationError("body", "referral_coupon_id, platform_user_id and shop_id are required")
	}
	if body.Status == "" {
		body.Status = domain.ShopReferralStatusActive
	} else if !body.Status.IsValid() {
		return domain.ShopReferral{}, domain.ValidationError("status", "must be 'active' or 'inactive'")
	}

	existing, err := c.adminRepo.GetShopReferralByShopID(ctx, body.ShopID)
	if err != nil {
		return domain.ShopReferral{}, err
	}
	if existing.ID != "" {
		return domain.ShopReferral{}, domain.ConflictError("shop already attached", "this shop already has a referral attachment; delete it first to reassign")
	}

	return c.adminRepo.CreateShopReferral(ctx, body)
}

func (c *adminUseCase) ListShopReferrals(ctx context.Context, callerID string, pagination request.Pagination) ([]domain.ShopReferral, error) {
	caller, err := c.adminRepo.GetAdminByID(ctx, callerID)
	if err != nil || caller.ID == "" {
		return nil, domain.NotFoundError("caller admin")
	}
	if caller.Role != domain.AdminRoleSuperAdmin {
		return nil, domain.ForbiddenError("only super admins can list all referral attachments")
	}
	return c.adminRepo.ListShopReferrals(ctx, pagination)
}

func (c *adminUseCase) GetShopReferral(ctx context.Context, callerID, referralID string) (domain.ShopReferral, error) {
	caller, err := c.adminRepo.GetAdminByID(ctx, callerID)
	if err != nil || caller.ID == "" {
		return domain.ShopReferral{}, domain.NotFoundError("caller admin")
	}
	referral, err := c.adminRepo.GetShopReferralByID(ctx, referralID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.ShopReferral{}, domain.NotFoundError("referral")
		}
		return domain.ShopReferral{}, err
	}
	if caller.Role != domain.AdminRoleSuperAdmin && caller.ID != referral.PlatformUserID {
		return domain.ShopReferral{}, domain.ForbiddenError("you cannot view this referral")
	}
	return referral, nil
}

func (c *adminUseCase) UpdateShopReferral(ctx context.Context, callerID, referralID string, body domain.ShopReferral) error {
	caller, err := c.adminRepo.GetAdminByID(ctx, callerID)
	if err != nil || caller.ID == "" {
		return domain.NotFoundError("caller admin")
	}
	if caller.Role != domain.AdminRoleSuperAdmin {
		return domain.ForbiddenError("only super admins can update referral attachments")
	}

	updates := map[string]interface{}{}
	if body.Status != "" {
		if !body.Status.IsValid() {
			return domain.ValidationError("status", "must be 'active' or 'inactive'")
		}
		updates["status"] = body.Status
	}
	if body.PlatformUserID != "" {
		updates["platform_user_id"] = body.PlatformUserID
	}
	if body.ReferralCouponID != "" {
		updates["referral_coupon_id"] = body.ReferralCouponID
	}
	if len(updates) == 0 {
		return domain.ValidationError("body", "no updatable fields provided")
	}

	if err := c.adminRepo.UpdateShopReferral(ctx, referralID, updates); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.NotFoundError("referral")
		}
		return err
	}
	return nil
}

func (c *adminUseCase) DeleteShopReferral(ctx context.Context, callerID, referralID string) error {
	caller, err := c.adminRepo.GetAdminByID(ctx, callerID)
	if err != nil || caller.ID == "" {
		return domain.NotFoundError("caller admin")
	}
	if caller.Role != domain.AdminRoleSuperAdmin {
		return domain.ForbiddenError("only super admins can delete referral attachments")
	}
	if err := c.adminRepo.DeleteShopReferral(ctx, referralID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.NotFoundError("referral")
		}
		return err
	}
	return nil
}

func (c *adminUseCase) ListShopsForSalesExecutive(ctx context.Context, callerID, platformUserID string, pagination request.Pagination) ([]domain.ShopReferral, error) {
	caller, err := c.adminRepo.GetAdminByID(ctx, callerID)
	if err != nil || caller.ID == "" {
		return nil, domain.NotFoundError("caller admin")
	}
	if platformUserID == "" {
		platformUserID = callerID
	}
	if caller.Role != domain.AdminRoleSuperAdmin && caller.ID != platformUserID {
		return nil, domain.ForbiddenError("you can only view shops attached to your own account")
	}
	return c.adminRepo.ListShopReferralsByPlatformUser(ctx, platformUserID, pagination)
}
