package interfaces

import "github.com/gin-gonic/gin"

type SellerGuideHandler interface {
	// Public
	GetCategories(ctx *gin.Context)
	GetShopPhotoTips(ctx *gin.Context)
	// Admin — guide videos
	GetGuideVideos(ctx *gin.Context)
	UploadGuideVideo(ctx *gin.Context)
	ReplaceGuideVideo(ctx *gin.Context)
	DeleteGuideVideo(ctx *gin.Context)
	// Admin — training videos
	GetTrainingVideos(ctx *gin.Context)
	UploadTrainingVideo(ctx *gin.Context)
	ReplaceTrainingVideo(ctx *gin.Context)
	DeleteTrainingVideo(ctx *gin.Context)
}
