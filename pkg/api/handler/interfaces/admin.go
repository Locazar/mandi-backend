package interfaces

import (
	"github.com/rohit221990/mandi-backend/pkg/domain"

	"github.com/gin-gonic/gin"
)

type AdminHandler interface {
	// RequirePermission is route-group middleware — see the implementation
	// doc comment in pkg/api/handler/admin.go for what it enforces.
	RequirePermission(key domain.PermissionKey) gin.HandlerFunc

	GetAllUsers(ctx *gin.Context)
	BlockUser(ctx *gin.Context)
	AdminSignUpVerify(ctx *gin.Context)

	AdminSignUp(ctx *gin.Context)
	SignupOtpSend(ctx *gin.Context)
	GetAdminWithShopVerificationByPhone(ctx *gin.Context)
	GetFullSalesReport(ctx *gin.Context)

	// Advertisement Management
	CreateAdvertisement(ctx *gin.Context)
	GetAllAdvertisements(ctx *gin.Context)
	GetActiveAdvertisements(ctx *gin.Context)
	GetActiveAdvertisementsFiltered(ctx *gin.Context)
	UpdateAdvertisement(ctx *gin.Context)
	DeleteAdvertisement(ctx *gin.Context)

	// Notification image library (welcome banners/frames used by the push composer)
	UploadNotificationImage(ctx *gin.Context)
	ListNotificationImages(ctx *gin.Context)

	// Global app config (single structured config used by client apps)
	GetGlobalConfig(ctx *gin.Context)
	UpdateGlobalConfig(ctx *gin.Context)
	GetOnboardingCopy(ctx *gin.Context)
	UpdateOnboardingCopy(ctx *gin.Context)

	// Advertisement Requests (seller-raised)
	GetAdvertisementPricePlans(ctx *gin.Context)
	CreateAdvertisementRequest(ctx *gin.Context)
	GetAllAdvertisementRequests(ctx *gin.Context)
	GetAdvertisementRequestByID(ctx *gin.Context)
	UpdateAdvertisementRequest(ctx *gin.Context)
	DeleteAdvertisementRequest(ctx *gin.Context)
	GetAdvertisementRequestInvoice(ctx *gin.Context)
	CreateAdvertisementPaymentOrder(ctx *gin.Context)
	VerifyAdvertisementPayment(ctx *gin.Context)
	AdvertisementPaymentFailed(ctx *gin.Context)

	// Advertisement pricing management (admin panel)
	GetAdvertisementPricing(ctx *gin.Context)
	CreateAdvertisementPlan(ctx *gin.Context)
	UpdateAdvertisementPlan(ctx *gin.Context)
	DeleteAdvertisementPlan(ctx *gin.Context)
	UpdateAdvertisementPricingConfig(ctx *gin.Context)

	// Feature flags
	GetFeatureFlagsObject(ctx *gin.Context)
	ListFeatureFlags(ctx *gin.Context)
	CreateFeatureFlag(ctx *gin.Context)
	UpdateFeatureFlag(ctx *gin.Context)
	DeleteFeatureFlag(ctx *gin.Context)

	// App configs
	ListAppConfigs(ctx *gin.Context)
	GetAppConfigByKey(ctx *gin.Context)
	CreateAppConfig(ctx *gin.Context)
	UpdateAppConfig(ctx *gin.Context)
	DeleteAppConfig(ctx *gin.Context)

	// Help center (contact settings + FAQs)
	GetHelpSettings(ctx *gin.Context)
	UpdateHelpSettings(ctx *gin.Context)
	ListHelpFAQs(ctx *gin.Context)
	CreateHelpFAQ(ctx *gin.Context)
	UpdateHelpFAQ(ctx *gin.Context)
	DeleteHelpFAQ(ctx *gin.Context)
	GetHelpCenter(ctx *gin.Context)

	// Subscription plans (admin panel)
	ListSubscriptionPlans(ctx *gin.Context)
	CreateSubscriptionPlan(ctx *gin.Context)
	UpdateSubscriptionPlan(ctx *gin.Context)
	DeleteSubscriptionPlan(ctx *gin.Context)

	//Shop Details
	CreateShop(ctx *gin.Context)
	GetAllShops(ctx *gin.Context)
	SearchShops(ctx *gin.Context)
	GetShopConflicts(ctx *gin.Context)
	GetShopByID(ctx *gin.Context)
	UpdateShop(ctx *gin.Context)
	UpdateSellerProductLimit(ctx *gin.Context)
	GetShopByOwnerID(ctx *gin.Context)
	VerifyShop(ctx *gin.Context)
	SubmitShopForReview(ctx *gin.Context)
	ApproveShop(ctx *gin.Context)
	RejectShop(ctx *gin.Context)
	GetVerificationStatus(ctx *gin.Context)
	SendNotificationToUsersInRadius(ctx *gin.Context)
	GetShopSocialDetails(ctx *gin.Context)

	UploadAdminProfileImage(ctx *gin.Context)
	UploadBusinessDocument(ctx *gin.Context)
	AddAdminProfile(ctx *gin.Context)
	GetAdminProfile(ctx *gin.Context)
	UpdateAdminProfile(ctx *gin.Context)

	//Documents
	UploadShopDocument(ctx *gin.Context)
	VerifyShopDocument(ctx *gin.Context)

	//Address
	UploadAddress(ctx *gin.Context)

	// Identity Document
	AdminDocumentOtpSend(ctx *gin.Context)
	AdminDocumentOtpVerify(ctx *gin.Context)

	// Product details
	GetAllProductDetails(ctx *gin.Context)
	UploadShopById(ctx *gin.Context)
	GetShopProfileImageById(ctx *gin.Context)
	// Shop Time
	SetShopTime(ctx *gin.Context)
	GetShopTime(ctx *gin.Context)
	// Logout
	UserLogout(ctx *gin.Context)

	// Dashboard
	GetDashboardStats(ctx *gin.Context)

	// Subscription GST configuration
	GetSubscriptionGSTConfig(ctx *gin.Context)
	UpdateSubscriptionGSTConfig(ctx *gin.Context)

	// Shop phone number change (OTP-gated)
	SendShopPhoneChangeOtp(ctx *gin.Context)
	VerifyShopPhoneChangeOtp(ctx *gin.Context)

	// Roles & permissions
	GetMyPermissions(ctx *gin.Context)
	CreateRole(ctx *gin.Context)
	ListRoles(ctx *gin.Context)
	GetRole(ctx *gin.Context)
	UpdateRole(ctx *gin.Context)
	DeleteRole(ctx *gin.Context)

	// Category requests (seller asks for a missing department/category)
	CreateCategoryRequest(ctx *gin.Context)
	ListCategoryRequests(ctx *gin.Context)
	ListMyCategoryRequests(ctx *gin.Context)
	UpdateCategoryRequestStatus(ctx *gin.Context)

	// Sales Executive referral program
	CreateShopReferral(ctx *gin.Context)
	ListShopReferrals(ctx *gin.Context)
	GetShopReferral(ctx *gin.Context)
	UpdateShopReferral(ctx *gin.Context)
	DeleteShopReferral(ctx *gin.Context)
	ListShopsForSalesExecutive(ctx *gin.Context)
}
