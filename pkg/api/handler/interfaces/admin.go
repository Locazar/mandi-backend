package interfaces

import "github.com/gin-gonic/gin"

type AdminHandler interface {
	GetAllUsers(ctx *gin.Context)
	BlockUser(ctx *gin.Context)

	AdminSignUp(ctx *gin.Context)
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
	SendNotificationToUsersInRadius(ctx *gin.Context)
}
