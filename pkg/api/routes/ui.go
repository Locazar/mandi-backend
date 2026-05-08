package routes

import (
	"github.com/gin-gonic/gin"
	handlerInterface "github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/middleware"
)

// UIRoutes registers UI-related routes
func UIRoutes(api *gin.RouterGroup, middleware middleware.Middleware, uiHandler handlerInterface.UIHandler) {

	seller := api.Group("/ui")
	{
		seller.POST("", middleware.AuthenticateAdmin(), uiHandler.SellerUIEndpoint)
		seller.GET("/", middleware.AuthenticateAdmin(), uiHandler.SellerUIEndpoint) // Optional: Allow GET for testing in browser
	}

}
