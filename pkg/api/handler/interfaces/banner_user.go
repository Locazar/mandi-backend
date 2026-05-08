package interfaces

import "github.com/gin-gonic/gin"

type BannerUserHandler interface {
	GetFilteredBanners(ctx *gin.Context)
}
