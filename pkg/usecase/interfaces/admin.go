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
	VerifyShop(ctx context.Context, verify request.ShopVerification) error
	GetVerificationStatus(ctx context.Context, adminId string) (domain.Admin, domain.ShopVerification, error)
	UpdateSellerProductLimit(ctx context.Context, adminID string, limit int) error

	// SubmitShopForReview is the seller's "submit for verification" action —
	// moves the shop to under_review. ApproveShop/RejectShop are the admin's
	// explicit go-live decision, and reindex/notify as a side effect.
	SubmitShopForReview(ctx context.Context, verify request.ShopVerification) error
	ApproveShop(ctx context.Context, shopID string) error
	RejectShop(ctx context.Context, shopID, remark string) error

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

	// Global app config (single structured config used by client apps)
	GetGlobalConfig(ctx context.Context) (domain.GlobalAppConfig, error)
	UpdateGlobalConfig(ctx context.Context, cfg domain.GlobalAppConfig) (domain.GlobalAppConfig, error)
	GetOnboardingWizardCopy(ctx context.Context) (domain.OnboardingWizardCopy, error)
	UpdateOnboardingWizardCopy(ctx context.Context, copy domain.OnboardingWizardCopy) (domain.OnboardingWizardCopy, error)

	// Help center (contact settings + FAQs)
	GetHelpSettings(ctx context.Context) (domain.HelpSettings, error)
	UpdateHelpSettings(ctx context.Context, settings domain.HelpSettings) (domain.HelpSettings, error)
	ListHelpFAQs(ctx context.Context) ([]domain.HelpFAQ, error)
	CreateHelpFAQ(ctx context.Context, faq domain.HelpFAQ) (domain.HelpFAQ, error)
	UpdateHelpFAQ(ctx context.Context, faq domain.HelpFAQ) (domain.HelpFAQ, error)
	DeleteHelpFAQ(ctx context.Context, faqID string) error
	GetHelpCenter(ctx context.Context) (domain.HelpCenter, error)

	// Subscription plans (admin panel)
	ListSubscriptionPlans(ctx context.Context) ([]domain.SubscriptionPlan, error)
	CreateSubscriptionPlan(ctx context.Context, plan domain.SubscriptionPlan) (domain.SubscriptionPlan, error)
	UpdateSubscriptionPlan(ctx context.Context, plan domain.SubscriptionPlan) (domain.SubscriptionPlan, error)
	DeleteSubscriptionPlan(ctx context.Context, planID string) error

	CreateShop(ctx context.Context, shop domain.ShopDetails) (domain.ShopDetails, error)
	GetAllShops(ctx context.Context, pagination request.Pagination) (shops []domain.ShopDetails, err error)
	SearchShops(ctx context.Context, filter request.ShopSearch) ([]domain.ShopDetails, error)
	GetShopConflicts(ctx context.Context, radiusMeters float64) ([]domain.ShopConflict, error)
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

	// Shop phone number change — OTP-gated so a new number can never be
	// saved without proving ownership of it first.
	SendShopPhoneChangeOtp(ctx context.Context, callerID, newPhone string) (otpID string, err error)
	VerifyShopPhoneChangeOtp(ctx context.Context, callerID, otpID, otp string) (phone string, err error)

	// Roles & permissions (custom roles + per-role sidebar/feature access).
	CreateRole(ctx context.Context, callerID string, role domain.Role) (domain.Role, error)
	ListRoles(ctx context.Context, callerID string) ([]domain.Role, error)
	GetRole(ctx context.Context, callerID, roleID string) (domain.Role, error)
	UpdateRole(ctx context.Context, callerID, roleID string, body domain.Role) (domain.Role, error)
	DeleteRole(ctx context.Context, callerID, roleID string) error
	GetRoleByName(ctx context.Context, name string) (domain.Role, error)
	// GetPermissionsForRole resolves the effective permission set for a
	// role name — super_admin always gets every key (see domain.Role doc);
	// every other role reads its stored role_permissions rows.
	GetPermissionsForRole(ctx context.Context, roleName string) ([]domain.RolePermissionGrant, error)

	// Subscription GST configuration (admin-managed rate applied to
	// subscription checkout — see subscriptionPaymentUseCase.CreateSubscriptionOrder).
	GetSubscriptionGSTConfig(ctx context.Context) (domain.SubscriptionGSTConfig, error)
	UpdateSubscriptionGSTConfig(ctx context.Context, cfg domain.SubscriptionGSTConfig) (domain.SubscriptionGSTConfig, error)

	// Used by platform-user creation to enforce email uniqueness (SignUp
	// itself only checks phone uniqueness).
	FindAdminByEmail(ctx context.Context, email string) (domain.Admin, error)

	// Sales Executive referral program (CRUD + attach lookup).
	FindAdminByReferralCouponID(ctx context.Context, referralCouponID string) (domain.Admin, error)
	CreateShopReferral(ctx context.Context, callerID string, body domain.ShopReferral) (domain.ShopReferral, error)
	ListShopReferrals(ctx context.Context, callerID string, pagination request.Pagination) ([]domain.ShopReferral, error)
	GetShopReferral(ctx context.Context, callerID, referralID string) (domain.ShopReferral, error)
	UpdateShopReferral(ctx context.Context, callerID, referralID string, body domain.ShopReferral) error
	DeleteShopReferral(ctx context.Context, callerID, referralID string) error
	// ListShopsForSalesExecutive returns the shops attached to the given
	// platform user. A Sales Executive may pass their own ID; only a super
	// admin may pass someone else's.
	ListShopsForSalesExecutive(ctx context.Context, callerID, platformUserID string, pagination request.Pagination) ([]domain.ShopReferral, error)
	// CountShopsAttachedToPlatformUser is used to block deleting a platform
	// user who is still an active Sales Executive for one or more shops.
	CountShopsAttachedToPlatformUser(ctx context.Context, platformUserID string) (int64, error)
	// RevokeAdminSessions deletes any active refresh session for adminID —
	// called on account deletion so an already-issued token can't keep
	// working against a deleted account until it naturally expires.
	RevokeAdminSessions(ctx context.Context, adminID string) error

	// Category requests: a seller asking for a department/category that
	// isn't in the existing list.
	CreateCategoryRequest(ctx context.Context, callerID string, req domain.CategoryRequest) (domain.CategoryRequest, error)
	ListCategoryRequests(ctx context.Context, callerID, status string, pagination request.Pagination) ([]domain.CategoryRequest, error)
	ListMyCategoryRequests(ctx context.Context, callerID string, pagination request.Pagination) ([]domain.CategoryRequest, error)
	UpdateCategoryRequestStatus(ctx context.Context, callerID, requestID string, status domain.CategoryRequestStatus, adminResponse string) error
}

// GetCategory(ctx context.Context) (helper.Category, any)
// 	SetCategory(ctx context.Context, body helper.Category)
