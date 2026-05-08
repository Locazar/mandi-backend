package interfaces

import "github.com/gin-gonic/gin"

// UIHandler defines methods for UI-related endpoints
type UIHandler interface {
	SellerUIEndpoint(ctx *gin.Context)
}
