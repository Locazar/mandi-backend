package routes

import (
	"github.com/gin-gonic/gin"
	handlerInterface "github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
)

func UserBannerRoutes(api *gin.RouterGroup, bannerUserHandler handlerInterface.BannerUserHandler) {
	banners := api.Group("/banners")
	{
		banners.POST("/filtered", bannerUserHandler.GetFilteredBanners)
	}
}
