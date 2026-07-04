package interfaces

import "github.com/gin-gonic/gin"

type PlatformUserHandler interface {
	ListAdmins(ctx *gin.Context)
	CreateAdmin(ctx *gin.Context)
	UpdateAdminRole(ctx *gin.Context)
	UpdateAdmin(ctx *gin.Context)
	DeactivateAdmin(ctx *gin.Context)
	DeleteAdmin(ctx *gin.Context)
}
