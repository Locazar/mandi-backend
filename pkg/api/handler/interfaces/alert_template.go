package interfaces

import "github.com/gin-gonic/gin"

type AlertTemplateHandler interface {
	// Admin endpoints
	ListTemplates(ctx *gin.Context)
	CreateTemplate(ctx *gin.Context)
	UpdateTemplate(ctx *gin.Context)
	DeleteTemplate(ctx *gin.Context)
	ToggleTemplate(ctx *gin.Context)
	SeedDefaults(ctx *gin.Context)

	// Seller-facing endpoints
	GetTemplateForSeller(ctx *gin.Context)
	GetFlows(ctx *gin.Context)
	GetFlow(ctx *gin.Context)
	CompleteStep(ctx *gin.Context)
}
