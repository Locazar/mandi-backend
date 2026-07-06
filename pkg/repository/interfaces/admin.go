package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type AdminRepository interface {
	FindAdminByEmail(ctx context.Context, email string) (domain.Admin, error)
	FindAdminByPhone(ctx context.Context, userName string) (domain.Admin, error)
	FindAdminWithShopVerificationByPhone(ctx context.Context, phone string) (domain.Admin, domain.ShopVerification, error)
	GetAdminByID(ctx context.Context, adminID string) (domain.Admin, error)
	SaveAdmin(ctx context.Context, admin domain.Admin) (domain.Admin, error)

	FindAllUser(ctx context.Context, pagination request.Pagination) (users []response.User, err error)

	CreateFullSalesReport(ctc context.Context, reqData request.SalesReport) (salesReport []response.SalesReport, err error)

	//stock side
	FindStockBySKU(ctx context.Context, sku string) (stock response.Stock, err error)
	VerifyShop(ctx context.Context, shopVerification request.ShopVerification, adminId string, verificationStatus bool) error
	// Advertisement Management
	CreateAdvertisement(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error)
	GetAllAdvertisements(ctx context.Context, pagination request.Pagination, filter domain.AdvertisementFilter) (ads []domain.Advertisement, err error)
	GetAdvertisementByID(ctx context.Context, advertisementID string) (domain.Advertisement, error)
	GetActiveAdvertisements(ctx context.Context) ([]domain.Advertisement, error)
	GetActiveAdvertisementsFiltered(ctx context.Context, filter domain.AdvertisementFilter) ([]domain.Advertisement, error)
	UpdateAdvertisement(ctx context.Context, ad domain.Advertisement) (domain.Advertisement, error)
	DeleteAdvertisement(ctx context.Context, advertisementID string) error

	// Advertisement Requests (seller-raised)
	CreateAdvertisementRequest(ctx context.Context, req domain.AdvertisementRequest) (domain.AdvertisementRequest, error)
	GetAllAdvertisementRequests(ctx context.Context, pagination request.Pagination, adminID string) ([]domain.AdvertisementRequest, error)
	GetAdvertisementRequestByID(ctx context.Context, requestID string) (domain.AdvertisementRequest, error)
	UpdateAdvertisementRequest(ctx context.Context, req domain.AdvertisementRequest) (domain.AdvertisementRequest, error)
	DeleteAdvertisementRequest(ctx context.Context, requestID string) error
	SetAdvertisementRequestPaymentOrder(ctx context.Context, requestID, orderID string) error
	GetAdvertisementRequestByPaymentOrderID(ctx context.Context, orderID string) (domain.AdvertisementRequest, error)
	MarkAdvertisementRequestPaid(ctx context.Context, requestID, paymentID string) error
	MarkAdvertisementRequestPaymentFailed(ctx context.Context, requestID string) error

	// Advertisement pricing configuration (admin-managed)
	ListAdvertisementPlans(ctx context.Context, activeOnly bool) ([]domain.AdvertisementPlanConfig, error)
	GetAdvertisementPlanByKey(ctx context.Context, planKey string) (domain.AdvertisementPlanConfig, error)
	CreateAdvertisementPlan(ctx context.Context, plan domain.AdvertisementPlanConfig) (domain.AdvertisementPlanConfig, error)
	UpdateAdvertisementPlan(ctx context.Context, plan domain.AdvertisementPlanConfig) (domain.AdvertisementPlanConfig, error)
	DeleteAdvertisementPlan(ctx context.Context, planID string) error
	GetAdvertisementPricingConfig(ctx context.Context) (domain.AdvertisementPricingConfig, error)
	UpdateAdvertisementPricingConfig(ctx context.Context, cfg domain.AdvertisementPricingConfig) (domain.AdvertisementPricingConfig, error)

	// Feature flags
	ListFeatureFlags(ctx context.Context) ([]domain.FeatureFlag, error)
	CreateFeatureFlag(ctx context.Context, flag domain.FeatureFlag) (domain.FeatureFlag, error)
	UpdateFeatureFlag(ctx context.Context, flag domain.FeatureFlag) (domain.FeatureFlag, error)
	DeleteFeatureFlag(ctx context.Context, flagID string) error

	// Subscription plans (admin panel)
	ListSubscriptionPlans(ctx context.Context) ([]domain.SubscriptionPlan, error)
	CreateSubscriptionPlan(ctx context.Context, plan domain.SubscriptionPlan) (domain.SubscriptionPlan, error)
	UpdateSubscriptionPlan(ctx context.Context, plan domain.SubscriptionPlan) (domain.SubscriptionPlan, error)
	DeleteSubscriptionPlan(ctx context.Context, planID string) error

	//Shop Details
	CreateShop(ctx context.Context, shop domain.ShopDetails) (domain.ShopDetails, error)
	GetAllShops(ctx context.Context, pagination request.Pagination) (shops []domain.ShopDetails, err error)
	SearchShops(ctx context.Context, filter request.ShopSearch) ([]domain.ShopDetails, error)
	GetShopByID(ctx context.Context, shopID string) (shop domain.ShopDetails, err error)
	UpdateShop(ctx context.Context, shop map[string]interface{}, shopId string) (map[string]interface{}, error)
	GetShopByOwnerID(ctx context.Context, ownerID string) (shop domain.ShopDetails, err error)

	SendNotificationToUsersInRadius(ctx context.Context, requestData request.NotificationRadiusRequest) error
	SendNotificationToUser(ctx context.Context, userID string, message string) error
	UploadAdminProfileImage(ctx context.Context, adminID string, imagePath string, shopId string) (string, error)
	UploadShopDocument(ctx context.Context, shopID string, documentType string, documentValue string) error
	UploadAddress(ctx context.Context, adminId string, address request.AddressRequest) error
	UploadAdminDocumentOtpSend(ctx context.Context, adminID string, documentType string, documentValue string) error

	GetVerificationStatus(ctx context.Context, adminId string) (domain.Admin, domain.ShopVerification, error)
	GetShopProfileImageById(ctx context.Context, shopId string) (string, error)
	DeleteRefreshSessionByUserID(ctx context.Context, adminId string) error
	GetShopSocialDetails(ctx context.Context, shopID string) ([]domain.ShopSocial, error)
	GetDashboardStats(ctx context.Context) (domain.DashboardStats, error)
}
