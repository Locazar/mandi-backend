package interfaces

import "github.com/gin-gonic/gin"

type SellerGuideHandler interface {
	GetCategories(ctx *gin.Context)
}
