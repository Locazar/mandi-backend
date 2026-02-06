package usecase

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/service/otp"
	"github.com/rohit221990/mandi-backend/pkg/service/token"
	service "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

type adminUseCase struct {
	adminRepo    interfaces.AdminRepository
	userRepo     interfaces.UserRepository
	authRepo     interfaces.AuthRepository
	optAuth      otp.OtpAuth
	tokenService token.TokenService
}

func NewAdminUseCase(repo interfaces.AdminRepository, userRepo interfaces.UserRepository, authRepo interfaces.AuthRepository, optAuth otp.OtpAuth, tokenService token.TokenService) service.AdminUseCase {

	return &adminUseCase{
		adminRepo:    repo,
		userRepo:     userRepo,
		authRepo:     authRepo,
		optAuth:      optAuth,
		tokenService: tokenService,
	}
}

func (c *adminUseCase) SignUp(ctx context.Context, signUpDetails domain.Admin) (string, error) {

	fmt.Printf("Admin SignUp called with mobile: %s\n", signUpDetails.Mobile)

	// Validate mobile number
	if signUpDetails.Mobile == "" || signUpDetails.Mobile == "null" {
		return "", fmt.Errorf("mobile number is required")
	}

	// Check if admin already exists by phone
	existAdmin, err := c.adminRepo.FindAdminByPhone(ctx, signUpDetails.Mobile)
	fmt.Printf("Existing admin check result: %+v, error: %v\n", existAdmin, err) // Debugging line
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to check admin details already exist")
	}

	// If admin already exists, return error
	if existAdmin.ID != 0 && existAdmin.VerifiedSeller {
		return "", errors.New("can't save admin - an admin already exists with this phone")
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

	errChan := make(chan error, 2)
	wait := sync.WaitGroup{}
	wait.Add(2)

	// Send OTP in goroutine
	go func() {
		defer wait.Done()
		fmt.Printf("Sending OTP to: %s\n", countryCode+signUpDetails.Mobile)
		_, err := c.optAuth.SentOtp(countryCode + signUpDetails.Mobile)
		if err != nil {
			errChan <- fmt.Errorf("failed to send otp \nerrors:%v", err.Error())
		}
	}()

	var adminID uint

	// Save admin in goroutine
	go func() {
		defer wait.Done()

		// Generate hashed password
		hashPass, err := utils.GenerateHashFromPassword(signUpDetails.Password)
		if err != nil {
			errChan <- utils.PrependMessageToError(err, "failed to hash the password")
			return
		}

		signUpDetails.Password = string(hashPass)
		err = c.adminRepo.SaveAdmin(ctx, signUpDetails)
		if err != nil {
			errChan <- utils.PrependMessageToError(err, "failed to save admin details")
			return
		}

		// Get the saved admin ID
		savedAdmin, err := c.adminRepo.FindAdminByPhone(ctx, signUpDetails.Mobile)
		if err != nil {
			errChan <- utils.PrependMessageToError(err, "failed to get saved admin ID")
			return
		}
		adminID = savedAdmin.ID
	}()

	wait.Wait()

	// Check for any errors from goroutines
	close(errChan)
	for err := range errChan {
		if err != nil {
			return "", err
		}
	}

	// Create OTP session
	otpID := uuid.NewString()
	otpSession := domain.OtpSession{
		OtpID:    otpID,
		AdminID:  adminID, // Using admin ID as user ID for OTP session
		Phone:    signUpDetails.Mobile,
		UserType: "Seller",
		ExpireAt: time.Now().Add(otpExpireDuration), // 2 minutes expire for otp
	}

	fmt.Printf("Creating OTP session: %+v\n", otpSession) // Debugging line
	err = c.authRepo.SaveOtpSession(ctx, otpSession)
	if err != nil {
		return "", utils.PrependMessageToError(err, "failed to save otp session")
	}

	return otpID, nil
}

func (c *adminUseCase) GetAdminWithShopVerificationByPhone(ctx context.Context, phone string) (domain.Admin, domain.ShopVerification, error) {
	return c.adminRepo.FindAdminWithShopVerificationByPhone(ctx, phone)
}

func (c *adminUseCase) AdminSignUpOtpVerify(ctx context.Context,
	otpVerifyDetails request.OTPVerify) (userID uint, err error) {

	fmt.Printf("Starting OTP verification for OTP ID: %s\n", otpVerifyDetails.OtpID)
	otpSession, err := c.authRepo.FindOtpSession(ctx, otpVerifyDetails.OtpID)
	fmt.Printf("otpSession: %+v\n", otpSession)
	fmt.Printf("err: %v\n", err)
	if err != nil {
		return 0, utils.PrependMessageToError(err, "failed to find otp session from database")
	}
	// fmt.Printf("otpSession.ExpireAt: %+v\n", otpSession.ExpireAt)
	// fmt.Printf("time.Since(otpSession.ExpireAt): %v\n", time.Since(otpSession.ExpireAt))
	// if time.Since(otpSession.ExpireAt) > 0 {
	// 	return 0, ErrOtpExpired
	// }

	valid, err := c.optAuth.VerifyOtp(countryCode+otpSession.Phone, otpVerifyDetails.Otp)
	fmt.Printf("valid: %v\n", valid)
	fmt.Printf("err: %v\n", err)
	if err != nil {
		return 0, utils.PrependMessageToError(err, "failed to verify otp")
	}
	if !valid {
		return 0, ErrInvalidOtp
	}

	fmt.Printf("otpSession.AdminID: %d\n", otpSession.AdminID)
	err = c.userRepo.UpdateAdminVerified(ctx, otpSession.AdminID)
	if err != nil {
		return 0, utils.PrependMessageToError(err, "failed to update user verified on database")
	}
	return otpSession.AdminID, nil

}
func (c *adminUseCase) GenerateAccessToken(ctx context.Context, tokenParams service.GenerateTokenParams) (string, error) {

	tokenReq := token.GenerateTokenRequest{
		UserID:   tokenParams.UserID,
		UsedFor:  tokenParams.UserType,
		ExpireAt: time.Now().Add(AccessTokenDuration),
	}

	tokenRes, err := c.tokenService.GenerateToken(tokenReq)

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
		RefreshToken: tokenRes.TokenString,
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

func (c *adminUseCase) VerifyShop(ctx context.Context, verify request.ShopVerification, adminId string) error {
	VerificationStatus := false
	if verify.Photo_Shop_Verification && verify.Business_Doc_Verification &&
		verify.Identity_Doc_Verification && verify.Address_Proof_Verification {
		VerificationStatus = true
	}
	err := c.adminRepo.VerifyShop(ctx, verify, adminId, VerificationStatus)

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

func (c *adminUseCase) GetAllAdvertisements(ctx context.Context, pagination request.Pagination) (ads []domain.Advertisement, err error) {
	ads, err = c.adminRepo.GetAllAdvertisements(ctx, pagination)
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

func (c *adminUseCase) CreateShop(ctx context.Context, shop domain.ShopDetails) (domain.ShopDetails, error) {
	createdShop, err := c.adminRepo.CreateShop(ctx, shop)
	if err != nil {
		return domain.ShopDetails{}, fmt.Errorf("failed to create shop \nerror:%v", err.Error())
	}
	return createdShop, nil
}

func (c *adminUseCase) GetAllShops(ctx context.Context, pagination request.Pagination) (shops []domain.ShopDetails, err error) {
	shops, err = c.adminRepo.GetAllShops(ctx, pagination)
	if err != nil {
		return nil, fmt.Errorf("failed to get all shops \nerror:%v", err.Error())
	}
	return shops, nil
}

func (c *adminUseCase) GetShopByID(ctx context.Context, shopID uint) (shop domain.ShopDetails, err error) {
	shop, err = c.adminRepo.GetShopByID(ctx, shopID)
	if err != nil {
		return domain.ShopDetails{}, fmt.Errorf("failed to get shop by id \nerror:%v", err.Error())
	}
	fmt.Printf("Retrieved shop: %+v\n", shop) // Debugging line
	return shop, nil
}
func (c *adminUseCase) UpdateShop(ctx context.Context, shop map[string]interface{}, shopId string) (map[string]interface{}, error) {
	updatedShop, err := c.adminRepo.UpdateShop(ctx, shop, shopId)
	if err != nil {
		return nil, fmt.Errorf("failed to update shop \nerror:%v", err.Error())
	}
	return updatedShop, nil
}

func (c *adminUseCase) GetShopByOwnerID(ctx context.Context, ownerID uint) (shop domain.ShopDetails, err error) {
	shop, err = c.adminRepo.GetShopByOwnerID(ctx, ownerID)
	if err != nil {
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

func (c *adminUseCase) SendNotificationToUser(ctx context.Context, userID uint, message string) error {
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

func (c *adminUseCase) UploadShopDocument(ctx context.Context, shopID uint, documentType string, documentValue string) error {

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
