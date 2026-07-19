package routes

import (
	"github.com/gin-gonic/gin"
	handlerInterface "github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
)

// SellerGuideRoutes registers public seller guide endpoints under /api.
func SellerGuideRoutes(api *gin.RouterGroup, sellerGuideHandler handlerInterface.SellerGuideHandler) {
	sellerGuide := api.Group("/seller-guide")
	{
		sellerGuide.GET("/categories", sellerGuideHandler.GetCategories)
		sellerGuide.GET("/shop-photo-tips", sellerGuideHandler.GetShopPhotoTips)
		sellerGuide.GET("/guide-videos", sellerGuideHandler.GetPublicGuideVideos)
		sellerGuide.GET("/training-videos", sellerGuideHandler.GetPublicTrainingVideos)
		sellerGuide.GET("/product-upload-guide-video", sellerGuideHandler.GetPublicProductUploadGuideVideo)
	}
}
