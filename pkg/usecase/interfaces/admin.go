package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type AdminUseCase interface {
	SignUp(ctx context.Context, admin domain.Admin) (string, error)
	AdminSignUpOtpVerify(ctx context.Context, otpVerifyDetails request.OTPVerify) (userID uint, err error)
	GetAdminWithShopVerificationByPhone(ctx context.Context, phone string) (domain.Admin, domain.ShopVerification, error)
	GenerateAccessToken(ctx context.Context, tokenParams GenerateTokenParams) (tokenString string, err error)
	GenerateRefreshToken(ctx context.Context, tokenParams GenerateTokenParams) (tokenString string, err error)

	FindAllUser(ctx context.Context, pagination request.Pagination) (users []response.User, err error)
	BlockOrUnBlockUser(ctx context.Context, blockDetails request.BlockUser) error

	GetFullSalesReport(ctx context.Context, requestData request.SalesReport) (salesReport []response.SalesReport, err error)
	VerifyShop(ctx context.Context, verify request.ShopVerification, adminId string) error
	GetVerificationStatus(ctx context.Context, adminId string) (domain.Admin, domain.ShopVerification, error)

	CreateAdvertisement(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error)
	GetAllAdvertisements(ctx context.Context, pagination request.Pagination) (ads []domain.Advertisement, err error)
	UpdateAdvertisement(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error)
	DeleteAdvertisement(ctx context.Context, advertisementID string) error
	CreateShop(ctx context.Context, shop domain.ShopDetails) (domain.ShopDetails, error)
	GetAllShops(ctx context.Context, pagination request.Pagination) (shops []domain.ShopDetails, err error)
	GetAdminByID(ctx context.Context, adminID uint) (domain.Admin, error)
	GetShopByID(ctx context.Context, shopID uint) (shop domain.ShopDetails, err error)
	UpdateShop(ctx context.Context, shop map[string]interface{}, shopId string) (map[string]interface{}, error)
	GetShopByOwnerID(ctx context.Context, ownerID uint) (shop domain.ShopDetails, err error)
	SendNotificationToUsersInRadius(ctx context.Context, requestData request.NotificationRadiusRequest) error
	SendNotificationToUser(ctx context.Context, userID uint, message string) error
	UploadAdminProfileImage(ctx context.Context, adminID string, imagePath string, shopId string) (string, error)
	DecodeTokenData(tokenString string) string
	UploadShopDocument(ctx context.Context, shopID uint, documentType string, documentValue string) error
	UploadAddress(ctx context.Context, adminId string, address request.AddressRequest) error
	VerifyShopDocument(ctx context.Context, otp string) error
	UploadAdminDocumentOtpSend(ctx context.Context, adminId string, documentType string, documentValue string) error
	UploadAdminDocumentOtpVerify(ctx context.Context, otp string, documentType string, documentValue string) error
	GetAllProductDetails(ctx context.Context) (products []any, err error)
	GetShopProfileImageById(ctx context.Context, shopId string) (string, error)
	UserLogout(ctx context.Context, adminId string) error
	GetShopSocialDetails(ctx context.Context, shopID uint) ([]domain.ShopSocial, error)
}

// GetCategory(ctx context.Context) (helper.Category, any)
// 	SetCategory(ctx context.Context, body helper.Category)
