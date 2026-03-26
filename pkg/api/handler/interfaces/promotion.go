package interfaces

import "github.com/gin-gonic/gin"

type PromotionHandler interface {
	GetAllPromotionCategories(ctx *gin.Context)
	GetPromotionCategoryByID(ctx *gin.Context)

	GetAllPromotionTypes(ctx *gin.Context)
	GetPromotionTypesByCategoryID(ctx *gin.Context)
	GetPromotionTypeByID(ctx *gin.Context)

	CreatePromotion(ctx *gin.Context)
	GetAllPromotions(ctx *gin.Context)
	GetPromotionByID(ctx *gin.Context)
	DeletePromotion(ctx *gin.Context)
}
