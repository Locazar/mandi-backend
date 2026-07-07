package interfaces

import (
	"context"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type AdminUseCase interface {
	SignUp(ctx context.Context, admin domain.Admin) (string, error)
	SignupOtpSend(ctx context.Context, details domain.Admin) (string, error)
	AdminSignUpOtpVerify(ctx context.Context, otpVerifyDetails request.OTPVerify) (userID string, shop domain.ShopDetails, err error)
	GetAdminWithShopVerificationByPhone(ctx context.Context, phone string) (domain.Admin, domain.ShopVerification, error)
	GenerateAccessToken(ctx context.Context, tokenParams GenerateTokenParams) (tokenString string, err error)
	GenerateRefreshToken(ctx context.Context, tokenParams GenerateTokenParams) (tokenString string, err error)

	FindAllUser(ctx context.Context, pagination request.Pagination) (users []response.User, err error)
	BlockOrUnBlockUser(ctx context.Context, blockDetails request.BlockUser) error

	GetFullSalesReport(ctx context.Context, requestData request.SalesReport) (salesReport []response.SalesReport, err error)
	VerifyShop(ctx context.Context, verify request.ShopVerification, adminId string) error
	GetVerificationStatus(ctx context.Context, adminId string) (domain.Admin, domain.ShopVerification, error)

	CreateAdvertisement(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error)
	GetAllAdvertisements(ctx context.Context, pagination request.Pagination, filter domain.AdvertisementFilter) (ads []domain.Advertisement, err error)
	GetAdvertisementByID(ctx context.Context, advertisementID string) (domain.Advertisement, error)
	GetActiveAdvertisements(ctx context.Context) ([]domain.Advertisement, error)
	GetActiveAdvertisementsFiltered(ctx context.Context, filter domain.AdvertisementFilter) ([]domain.Advertisement, error)
	UpdateAdvertisement(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error)
	DeleteAdvertisement(ctx context.Context, advertisementID string) error

	// Advertisement Requests (seller-raised)
	GetAdvertisementPricePlans(ctx context.Context, startDate, endDate time.Time) ([]domain.AdvertisementPricePlan, error)
	CreateAdvertisementRequest(ctx context.Context, req domain.AdvertisementRequest) (domain.AdvertisementRequest, error)
	GetAllAdvertisementRequests(ctx context.Context, pagination request.Pagination, adminID string) ([]domain.AdvertisementRequest, error)
	GetAdvertisementRequestByID(ctx context.Context, requestID string) (domain.AdvertisementRequest, error)
	UpdateAdvertisementRequest(ctx context.Context, req domain.AdvertisementRequest) (domain.AdvertisementRequest, error)
	DeleteAdvertisementRequest(ctx context.Context, requestID string) error
	GetAdvertisementRequestInvoice(ctx context.Context, requestID string) (domain.AdvertisementInvoice, error)
	CreateAdvertisementPaymentOrder(ctx context.Context, adminID, requestID string) (orderID string, keyID string, amountMinor int64, err error)
	VerifyAdvertisementPayment(ctx context.Context, adminID, orderID, paymentID, signature string) error
	AdvertisementPaymentFailed(ctx context.Context, adminID, orderID string) error

	// Advertisement pricing management (admin panel)
	ListAdvertisementPlanConfigs(ctx context.Context) ([]domain.AdvertisementPlanConfig, error)
	CreateAdvertisementPlanConfig(ctx context.Context, plan domain.AdvertisementPlanConfig) (domain.AdvertisementPlanConfig, error)
	UpdateAdvertisementPlanConfig(ctx context.Context, plan domain.AdvertisementPlanConfig) (domain.AdvertisementPlanConfig, error)
	DeleteAdvertisementPlanConfig(ctx context.Context, planID string) error
	GetAdvertisementPricingConfig(ctx context.Context) (domain.AdvertisementPricingConfig, error)
	UpdateAdvertisementPricingConfig(ctx context.Context, cfg domain.AdvertisementPricingConfig) (domain.AdvertisementPricingConfig, error)

	// Feature flags
	GetFeatureFlagsObject(ctx context.Context) (map[string]bool, error)
	ListFeatureFlags(ctx context.Context) ([]domain.FeatureFlag, error)
	CreateFeatureFlag(ctx context.Context, flag domain.FeatureFlag) (domain.FeatureFlag, error)
	UpdateFeatureFlag(ctx context.Context, flag domain.FeatureFlag) (domain.FeatureFlag, error)
	DeleteFeatureFlag(ctx context.Context, flagID string) error

	// App configs
	ListAppConfigs(ctx context.Context) ([]domain.AppConfig, error)
	GetAppConfigByKey(ctx context.Context, configKey string) (domain.AppConfig, error)
	CreateAppConfig(ctx context.Context, cfg domain.AppConfig) (domain.AppConfig, error)
	UpdateAppConfig(ctx context.Context, cfg domain.AppConfig) (domain.AppConfig, error)
	DeleteAppConfig(ctx context.Context, configID string) error

	// Subscription plans (admin panel)
	ListSubscriptionPlans(ctx context.Context) ([]domain.SubscriptionPlan, error)
	CreateSubscriptionPlan(ctx context.Context, plan domain.SubscriptionPlan) (domain.SubscriptionPlan, error)
	UpdateSubscriptionPlan(ctx context.Context, plan domain.SubscriptionPlan) (domain.SubscriptionPlan, error)
	DeleteSubscriptionPlan(ctx context.Context, planID string) error

	CreateShop(ctx context.Context, shop domain.ShopDetails) (domain.ShopDetails, error)
	GetAllShops(ctx context.Context, pagination request.Pagination) (shops []domain.ShopDetails, err error)
	SearchShops(ctx context.Context, filter request.ShopSearch) ([]domain.ShopDetails, error)
	GetAdminByID(ctx context.Context, adminID string) (domain.Admin, error)
	GetShopByID(ctx context.Context, shopID string) (shop domain.ShopDetails, err error)
	UpdateShop(ctx context.Context, shop map[string]interface{}, shopId string) (map[string]interface{}, error)
	GetShopByOwnerID(ctx context.Context, ownerID string) (shop domain.ShopDetails, err error)
	SendNotificationToUsersInRadius(ctx context.Context, requestData request.NotificationRadiusRequest) error
	SendNotificationToUser(ctx context.Context, userID string, message string) error
	UploadAdminProfileImage(ctx context.Context, adminID string, imagePath string, shopId string) (string, error)
	DecodeTokenData(tokenString string) string
	UploadShopDocument(ctx context.Context, shopID string, documentType string, documentValue string) error
	UploadAddress(ctx context.Context, adminId string, address request.AddressRequest) error
	VerifyShopDocument(ctx context.Context, otp string) error
	UploadAdminDocumentOtpSend(ctx context.Context, adminId string, documentType string, documentValue string) error
	UploadAdminDocumentOtpVerify(ctx context.Context, otp string, documentType string, documentValue string) error
	GetAllProductDetails(ctx context.Context) (products []any, err error)
	GetShopProfileImageById(ctx context.Context, shopId string) (string, error)
	UserLogout(ctx context.Context, adminId string) error
	GetShopSocialDetails(ctx context.Context, shopID string) ([]domain.ShopSocial, error)
	GetDashboardStats(ctx context.Context) (domain.DashboardStats, error)
}

// GetCategory(ctx context.Context) (helper.Category, any)
// 	SetCategory(ctx context.Context, body helper.Category)
