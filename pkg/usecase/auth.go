package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/service/otp"
	"github.com/rohit221990/mandi-backend/pkg/service/sms"
	"github.com/rohit221990/mandi-backend/pkg/service/token"
	service "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
	"gorm.io/gorm"
)

const (
	countryCode = "+91"
)

type authUseCase struct {
	authRepo interfaces.AuthRepository

	userRepo          interfaces.UserRepository
	adminRepo         interfaces.AdminRepository
	tokenService      token.TokenService
	optAuth           otp.OtpAuth
	otpService        *otp.MobileOTPService
	smsService        *sms.TwoFactorSMSService
	skipOTPValidation bool
}

func NewAuthUseCase(authRepo interfaces.AuthRepository, tokenService token.TokenService,
	userRepo interfaces.UserRepository, adminRepo interfaces.AdminRepository,
	optAuth otp.OtpAuth, otpService *otp.MobileOTPService, smsService *sms.TwoFactorSMSService,
	skipOTPValidation ...bool) service.AuthUseCase {

	skipValidation := false
	if len(skipOTPValidation) > 0 {
		skipValidation = skipOTPValidation[0]
	}

	return &authUseCase{
		userRepo:          userRepo,
		adminRepo:         adminRepo,
		tokenService:      tokenService,
		authRepo:          authRepo,
		optAuth:           optAuth,
		otpService:        otpService,
		smsService:        smsService,
		skipOTPValidation: skipValidation,
	}
}

const (
	// 30-day session, shared by both customer/user and seller/admin logins
	// (pkg/usecase/admin.go uses the same constants). Access and refresh
	// share the same duration today rather than the stricter
	// short-access/long-refresh split some APIs use, since none of the
	// client apps currently proactively rotate the access token mid-session
	// — they only call the renew-access-token endpoint after a 401.
	AccessTokenDuration  = time.Hour * 24 * 30
	RefreshTokenDuration = time.Hour * 24 * 30
)

func (c *authUseCase) UserLogin(ctx context.Context, loginDetails request.Login) (string, error) {

	var (
		user domain.User
		err  error
	)
	switch {
	case loginDetails.Email != "":
		user, err = c.userRepo.FindUserByEmail(ctx, loginDetails.Email)
	case loginDetails.Phone != "":
		user, err = c.userRepo.FindUserByPhoneNumber(ctx, loginDetails.Phone)
	default:
		return "", ErrEmptyLoginCredentials
	}

	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to find user from database")
	}

	if user.ID == "" {
		return "", ErrUserNotExist
	}

	// if !user.Verified {
	// 	return "", ErrUserNotVerified
	// }

	if user.BlockStatus {
		return "", ErrUserBlocked
	}

	err = utils.ComparePasswordWithHashedPassword(loginDetails.Password, user.Password)
	if err != nil {
		return "", ErrWrongPassword
	}

	return user.ID, nil
}

// createUserByPhone creates a minimal phone-only user account. Safe under
// concurrent requests for the same new phone number: if two requests race
// on the INSERT, the loser recovers by re-fetching the winner's row rather
// than erroring out.
func (c *authUseCase) createUserByPhone(ctx context.Context, phone string) (domain.User, error) {
	newUserID, err := c.userRepo.SaveUser(ctx, domain.User{Phone: phone})
	if err == nil {
		return domain.User{ID: newUserID, Phone: phone}, nil
	}

	if strings.Contains(err.Error(), "unique constraint") ||
		strings.Contains(err.Error(), "duplicate key") {
		recovered, findErr := c.userRepo.FindUserByPhoneNumber(ctx, phone)
		if findErr == nil && recovered.ID != "" {
			return recovered, nil
		}
	}

	return domain.User{}, fmt.Errorf("failed to create user account \nerror:%v", err.Error())
}

func (c *authUseCase) UserLoginOtpSend(ctx context.Context, loginDetails request.OTPLogin) (string, error) {

	var (
		user domain.User
		err  error
	)

	switch {

	case loginDetails.Email != "":
		user, err = c.userRepo.FindUserByEmail(ctx, loginDetails.Email)
		if err != nil {
			return "", fmt.Errorf("can't find the user \nerror:%v", err.Error())
		}
		if user.ID == "" {
			return "", ErrUserNotExist
		}
	case loginDetails.Phone != "":
		user, err = c.userRepo.FindUserByPhoneNumber(ctx, loginDetails.Phone)
		if err != nil {
			return "", fmt.Errorf("can't find the user \nerror:%v", err.Error())
		}
		if user.ID == "" {
			// No account yet for this phone number — sign-in doubles as
			// signup here, mirroring the seller app's
			// findOrCreateAdminByMobile pattern: any phone number that
			// requests an OTP gets a minimal account created on the fly.
			// These phone-only accounts authenticate via OTP, never a
			// password (SaveUser stores an empty password placeholder).
			user, err = c.createUserByPhone(ctx, loginDetails.Phone)
			if err != nil {
				return "", err
			}
		}
	default:
		return "", ErrEmptyLoginCredentials
	}

	if user.BlockStatus {
		return "", ErrUserBlocked
	}

	// Generate a 6-digit OTP and store only its hash (never plaintext)
	generatedOTP, err := c.otpService.GenerateOTP()
	if err != nil {
		return "", fmt.Errorf("failed to generate otp \nerror:%v", err.Error())
	}

	otpHash, err := c.otpService.HashOTP(generatedOTP)
	if err != nil {
		return "", fmt.Errorf("failed to hash otp \nerror:%v", err.Error())
	}

	otpID := uuid.NewString()
	otpSession := domain.OtpSession{
		OtpID:    otpID,
		OtpHash:  otpHash,
		UserID:   user.ID,
		Phone:    user.Phone,
		UserType: domain.UserTypeUser,
		ExpireAt: c.otpService.CalculateOTPExpiry(),
	}
	if err := c.authRepo.SaveOtpSession(ctx, otpSession); err != nil {
		return "", fmt.Errorf("failed to save otp session \nerror:%v", err.Error())
	}

	// Send the OTP via the 2factor.in SMS API (skipped when SKIP_OTP_VALIDATION=true)
	if !c.skipOTPValidation {
		if err := c.smsService.SendOTPSMS(user.Phone, generatedOTP); err != nil {
			return "", fmt.Errorf("failed to send otp \nerror:%v", err.Error())
		}
	} else {
		log.Printf("[SendOTP user] skipOTPValidation=true, not sending SMS, otp=%s phone=%s", generatedOTP, user.Phone)
	}

	return otpID, nil
}

func (c *authUseCase) LoginOtpVerify(ctx context.Context, otpVerifyDetails request.OTPVerify) (string, error) {

	otpSession, err := c.authRepo.FindOtpSession(ctx, otpVerifyDetails.OtpID)
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to find otp session from database")
	}

	// Reject expired OTP sessions
	if c.otpService.IsOTPExpired(otpSession.ExpireAt) {
		return "", ErrOtpExpired
	}

	// Verify the entered OTP against the stored hash (skip when SKIP_OTP_VALIDATION=true)
	log.Printf("[VerifyOTP user] skipOTPValidation=%v otp_id=%s", c.skipOTPValidation, otpVerifyDetails.OtpID)
	if !c.skipOTPValidation {
		if err := c.otpService.VerifyOTP(otpVerifyDetails.Otp, otpSession.OtpHash); err != nil {
			return "", ErrInvalidOtp
		}
	}

	// OTP verified — invalidate the session so the same OTP cannot be reused.
	if delErr := c.authRepo.DeleteOtpSession(ctx, otpSession.OtpID); delErr != nil {
		log.Printf("warning: failed to delete used otp session %s: %v", otpSession.OtpID, delErr)
	}

	return otpSession.UserID, nil
}

func (c *authUseCase) AdminLogin(ctx context.Context, loginDetails request.Login) (domain.Admin, error) {

	var (
		admin domain.Admin
		err   error
	)
	switch {
	case loginDetails.Email != "":
		admin, err = c.adminRepo.FindAdminByEmail(ctx, loginDetails.Email)
	case loginDetails.Phone != "":
		admin, _, err = c.adminRepo.FindAdminWithShopVerificationByPhone(ctx, loginDetails.Phone)
	default:
		return domain.Admin{}, ErrEmptyLoginCredentials
	}

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.Admin{}, ErrUserNotExist
		}
		return domain.Admin{}, utils.PrependMessageToError(err, "failed to find admin")
	}

	if admin.ID == "" {
		return domain.Admin{}, ErrUserNotExist
	}

	if admin.Password == "" {
		// Account has no password set (e.g. an OTP-only seller account) —
		// password login can never succeed for it, regardless of what was
		// typed, rather than comparing against an empty hash.
		return domain.Admin{}, ErrWrongPassword
	}
	if err := utils.ComparePasswordWithHashedPassword(loginDetails.Password, admin.Password); err != nil {
		return domain.Admin{}, ErrWrongPassword
	}

	return admin, nil
}

func (c *authUseCase) AdminSignUpOtpSend(ctx context.Context, phone string) (string, error) {

	if phone == "" {
		return "", ErrEmptyLoginCredentials
	}

	admin, err := c.adminRepo.FindAdminByPhone(ctx, phone)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", utils.PrependMessageToError(err, "failed to check admin phone")
	}

	// Send OTP via SMS
	_, err = c.optAuth.SentOtp(countryCode + phone)
	if err != nil {
		return "", fmt.Errorf("failed to send otp \nerrors:%v", err.Error())
	}

	otpID := uuid.NewString()
	otpSession := domain.OtpSession{
		OtpID:    otpID,
		UserID:   admin.ID,
		Phone:    phone,
		UserType: domain.UserType(token.Admin),
		AdminID:  admin.ID,
		ExpireAt: c.otpService.CalculateOTPExpiry(),
	}

	if err := c.authRepo.SaveOtpSession(ctx, otpSession); err != nil {
		return "", utils.PrependMessageToError(err, "failed to save otp session")
	}

	return otpID, nil
}

func (c *authUseCase) AdminSignUpOtpVerify(ctx context.Context, otpVerifyDetails request.OTPVerify) (string, error) {
	otpSession, err := c.authRepo.FindOtpSession(ctx, otpVerifyDetails.OtpID)
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to find otp session from database")
	}

	if otpSession.Phone == "" {
		return "", ErrInvalidOtp
	}

	// Reject expired OTP sessions.
	if c.otpService.IsOTPExpired(otpSession.ExpireAt) {
		return "", ErrOtpExpired
	}

	// Verify the code with the SMS provider that issued it. AdminSignUpOtpSend
	// delegates to optAuth.SentOtp and stores no local hash on the session, so
	// verification must go through optAuth.VerifyOtp — otpService.VerifyOTP has
	// no OtpHash to compare against here and would always fail.
	//
	// Without this check the endpoint logged in — or silently REGISTERED — a
	// seller account for any phone number given only a valid otp_id, with no
	// knowledge of the code that was texted.
	log.Printf("[VerifyOTP admin] skipOTPValidation=%v otp_id=%s", c.skipOTPValidation, otpVerifyDetails.OtpID)
	if !c.skipOTPValidation {
		valid, verifyErr := c.optAuth.VerifyOtp(countryCode+otpSession.Phone, otpVerifyDetails.Otp)
		if verifyErr != nil {
			return "", utils.PrependMessageToError(verifyErr, "failed to verify otp")
		}
		if !valid {
			return "", ErrInvalidOtp
		}
	}

	// The OTP is spent once verified, whatever the outcome below — a failed
	// registration must require a fresh code rather than leave this one live.
	if delErr := c.authRepo.DeleteOtpSession(ctx, otpSession.OtpID); delErr != nil {
		log.Printf("warning: failed to delete used otp session %s: %v", otpSession.OtpID, delErr)
	}

	admin, err := c.AdminLogin(ctx, request.Login{Phone: otpSession.Phone})
	if err == nil && admin.ID != "" {
		return admin.ID, nil
	}

	if err == nil || !errors.Is(err, ErrUserNotExist) {
		return "", utils.PrependMessageToError(err, "failed to verify admin login")
	}

	// Role is intentionally left unset: role is a platform-user-only concept
	// (assigned by an admin-portal super_admin via CreateAdmin) — a
	// self-registered seller/customer account must never carry one. See
	// SaveAdmin for why this used to silently become 'super_admin'.
	newAdmin := domain.Admin{
		Mobile:         otpSession.Phone,
		Status:         "active",
		VerifiedSeller: false,
	}

	savedAdmin, err := c.adminRepo.SaveAdmin(ctx, newAdmin)
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to register admin")
	}

	if savedAdmin.ID == "" {
		return "", ErrUserNotExist
	}

	return savedAdmin.ID, nil
}

func (c *authUseCase) GenerateAccessToken(ctx context.Context, tokenParams service.GenerateTokenParams) (string, error) {

	tokenReq := token.GenerateTokenRequest{
		UserID:   tokenParams.UserID,
		UsedFor:  tokenParams.UserType,
		ExpireAt: time.Now().Add(AccessTokenDuration),
	}

	tokenRes, err := c.tokenService.GenerateToken(tokenReq)

	return tokenRes.TokenString, err
}
func (c *authUseCase) GenerateRefreshToken(ctx context.Context, tokenParams service.GenerateTokenParams) (string, error) {

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
		TokenID:      tokenRes.TokenID,
		UserType:     string(tokenReq.UsedFor),
		RefreshToken: utils.HashRefreshToken(tokenRes.TokenString),
		ExpireAt:     expireAt.Format(time.RFC3339),
	})
	if err != nil {
		return "", err
	}
	log.Printf("successfully refresh token created and refresh session stored in database")
	return tokenRes.TokenString, nil
}

func (c *authUseCase) VerifyAndGetRefreshTokenSession(ctx context.Context, refreshToken string, usedFor token.UserType) (domain.RefreshSession, error) {

	verifyReq := token.VerifyTokenRequest{
		TokenString: refreshToken,
		UsedFor:     usedFor,
	}
	verifyRes, err := c.tokenService.VerifyToken(verifyReq)
	if err != nil {
		return domain.RefreshSession{}, utils.PrependMessageToError(ErrInvalidRefreshToken, err.Error())
	}

	refreshSession, err := c.authRepo.FindRefreshSessionByTokenID(ctx, verifyRes.TokenID, string(usedFor))
	if err != nil {
		return domain.RefreshSession{}, err
	}

	if refreshSession.TokenID == "" {
		return domain.RefreshSession{}, ErrRefreshSessionNotExist
	}

	if time.Since(refreshSession.ExpireAt) > 0 {
		return domain.RefreshSession{}, ErrRefreshSessionExpired
	}

	// if refreshSession.IsBlocked {
	// 	return domain.RefreshSession{}, ErrRefreshSessionBlocked
	// }

	return refreshSession, nil
}

func (c *authUseCase) UserSignUp(ctx context.Context, signUpDetails domain.User) (string, error) {

	existUser, err := c.userRepo.FindUserByUserNameEmailOrPhoneNotID(ctx, signUpDetails)
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to check user details already exist")
	}

	// if user credentials already exist and  verified then return it as errors
	if existUser.ID != "" && existUser.PhoneVerified {
		err = utils.CompareUserExistingDetails(existUser, signUpDetails)
		err = utils.AppendMessageToError(ErrUserAlreadyExit, err.Error())
		return "", err
	}

	userID := existUser.ID

	if userID == "" { // if user not exist then save user on database
		hashPass, err := utils.GenerateHashFromPassword(signUpDetails.Password)
		if err != nil {
			return "", utils.PrependMessageToError(err, "failed to hash the password")
		}

		signUpDetails.Password = string(hashPass)
		userID, err = c.userRepo.SaveUser(ctx, signUpDetails)

		if err != nil {
			return "", utils.PrependMessageToError(err, "failed to save user details")
		}
	}

	// Generate a 6-digit OTP and store only its hash (never plaintext), exactly
	// as LoginOtpSend does. SingUpOtpVerify compares the submitted code against
	// otpSession.OtpHash, so a session saved without one can never verify —
	// bcrypt rejects every candidate against an empty hash.
	generatedOTP, err := c.otpService.GenerateOTP()
	if err != nil {
		return "", fmt.Errorf("failed to generate otp \nerror:%v", err.Error())
	}

	otpHash, err := c.otpService.HashOTP(generatedOTP)
	if err != nil {
		return "", fmt.Errorf("failed to hash otp \nerror:%v", err.Error())
	}

	otpID := uuid.NewString()
	otpSession := domain.OtpSession{
		OtpID:    otpID,
		OtpHash:  otpHash,
		UserID:   userID,
		Phone:    signUpDetails.Phone,
		UserType: domain.UserTypeUser,
		ExpireAt: c.otpService.CalculateOTPExpiry(),
	}
	if err := c.authRepo.SaveOtpSession(ctx, otpSession); err != nil {
		return "", fmt.Errorf("failed to save otp session \nerror:%v", err.Error())
	}

	// Send the OTP via the 2factor.in SMS API (skipped when SKIP_OTP_VALIDATION=true)
	if !c.skipOTPValidation {
		if err := c.smsService.SendOTPSMS(signUpDetails.Phone, generatedOTP); err != nil {
			return "", fmt.Errorf("failed to send otp \nerror:%v", err.Error())
		}
	} else {
		log.Printf("[SendOTP signup] skipOTPValidation=true, not sending SMS, otp=%s phone=%s", generatedOTP, signUpDetails.Phone)
	}

	return otpID, nil
}

func (c *authUseCase) SingUpOtpVerify(ctx context.Context,
	otpVerifyDetails request.OTPVerify) (userID string, err error) {

	otpSession, err := c.authRepo.FindOtpSession(ctx, otpVerifyDetails.OtpID)
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to find otp session from database")
	}

	// Reject expired OTP sessions.
	if c.otpService.IsOTPExpired(otpSession.ExpireAt) {
		return "", ErrOtpExpired
	}

	// Verify the entered OTP against the stored hash (skip when
	// SKIP_OTP_VALIDATION=true). Without this the endpoint accepted ANY value
	// as a valid OTP, so possession of an otp_id was enough to obtain a token.
	// Mirrors LoginOtpVerify, which is the reference implementation.
	log.Printf("[VerifyOTP signup] skipOTPValidation=%v otp_id=%s", c.skipOTPValidation, otpVerifyDetails.OtpID)
	if !c.skipOTPValidation {
		if err := c.otpService.VerifyOTP(otpVerifyDetails.Otp, otpSession.OtpHash); err != nil {
			return "", ErrInvalidOtp
		}
	}

	err = c.userRepo.UpdateVerified(ctx, otpSession.UserID)
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to update user verified on database")
	}

	// OTP verified — invalidate the session so the same OTP cannot be reused.
	if delErr := c.authRepo.DeleteOtpSession(ctx, otpSession.OtpID); delErr != nil {
		log.Printf("warning: failed to delete used otp session %s: %v", otpSession.OtpID, delErr)
	}

	return otpSession.UserID, nil
}

// google login
func (c *authUseCase) GoogleLogin(ctx context.Context, user domain.User) (userID string, err error) {

	existUser, err := c.userRepo.FindUserByEmail(ctx, user.Email)
	if err != nil {
		return userID, fmt.Errorf("failed to get user details with given email \nerror:%v", err.Error())
	}

	if existUser.ID != "" {
		return existUser.ID, nil
	}

	userID, err = c.userRepo.SaveUser(ctx, user)
	if err != nil {
		return userID, fmt.Errorf("failed to save user details \nerror:%v", err.Error())
	}

	return userID, nil
}

func (c *authUseCase) UserLoginOtpSendEmail(ctx context.Context, emailDetails request.OTPLoginEmail) (string, error) {

	user, err := c.userRepo.FindUserByEmail(ctx, emailDetails.Email)
	if err != nil {
		return "", fmt.Errorf("can't find the user \nerror:%v", err.Error())
	}

	if user.ID == "" {
		return "", ErrUserNotExist
	}

	if user.BlockStatus {
		return "", ErrUserBlocked
	}

	otpCode, err := c.optAuth.SentOtpEmail(user.Email)
	if err != nil {
		return "", fmt.Errorf("failed to send otp \nerrors:%v", err.Error())
	}

	otpID := uuid.NewString()

	otpSession := domain.OtpSession{
		OtpID:    otpCode,
		UserID:   user.ID,
		ExpireAt: c.otpService.CalculateOTPExpiry(),
	}
	err = c.authRepo.SaveOtpSession(ctx, otpSession)
	if err != nil {
		return "", fmt.Errorf("failed to save otp session \nerror:%v", err.Error())
	}

	return otpID, nil
}

func (c *authUseCase) LoginOtpVerifyEmail(ctx context.Context, otpVerifyDetails request.OTPVerify) (string, error) {

	otpSession, err := c.authRepo.FindOtpSessionEmail(ctx, otpVerifyDetails.OtpID)
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to find otp session from database")
	}

	// Reject expired OTP sessions.
	if c.otpService.IsOTPExpired(otpSession.ExpireAt) {
		return "", ErrOtpExpired
	}

	valid, err := c.optAuth.VerifyOtpEmail(otpSession.Email, otpVerifyDetails.Otp)
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to verify otp")
	}
	if !valid {
		return "", ErrInvalidOtp
	}

	return otpSession.UserID, nil
}

func (c *authUseCase) UserLogout(ctx context.Context, adminID string, userType string) error {
	err := c.userRepo.DeleteRefreshSessionByUserID(ctx, adminID, userType)
	return err
}
