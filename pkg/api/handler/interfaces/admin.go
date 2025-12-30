package interfaces

import "github.com/gin-gonic/gin"

type AdminHandler interface {
	GetAllUsers(ctx *gin.Context)
	BlockUser(ctx *gin.Context)
	AdminSignUpVerify(ctx *gin.Context)

	AdminSignUp(ctx *gin.Context)
	GetAdminWithShopVerificationByPhone(ctx *gin.Context)
	GetFullSalesReport(ctx *gin.Context)

	// Advertisement Management
	CreateAdvertisement(ctx *gin.Context)
	GetAllAdvertisements(ctx *gin.Context)
	UpdateAdvertisement(ctx *gin.Context)
	DeleteAdvertisement(ctx *gin.Context)

	//Shop Details
	CreateShop(ctx *gin.Context)
	GetAllShops(ctx *gin.Context)
	GetShopByID(ctx *gin.Context)
	UpdateShop(ctx *gin.Context)
	GetShopByOwnerID(ctx *gin.Context)
	VerifyShop(ctx *gin.Context)
	GetVerificationStatus(ctx *gin.Context)
	SendNotificationToUsersInRadius(ctx *gin.Context)

	UploadAdminProfileImage(ctx *gin.Context)
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
	// Logout
	UserLogout(ctx *gin.Context)
}
