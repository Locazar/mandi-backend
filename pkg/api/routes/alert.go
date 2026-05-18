package routes

import (
	"github.com/rohit221990/mandi-backend/pkg/api/handler"
	"github.com/gin-gonic/gin"
)

// SetupAlertRoutes registers alert-related routes
func SetupAlertRoutes(router *gin.RouterGroup, alertHandler *handler.AlertHandler) {
	alerts := router.Group("/alerts")
	{
		// GET /api/v1/seller/alerts - Get all alerts for seller
		alerts.GET("", alertHandler.GetSellerAlerts)

		// POST /api/v1/seller/alerts/:key/dismiss - Dismiss an alert
		alerts.POST("/:key/dismiss", alertHandler.DismissAlert)
	}
}
