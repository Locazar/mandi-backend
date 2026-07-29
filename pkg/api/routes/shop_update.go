package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler"
	handlerInterface "github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/middleware"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// ShopUpdateRoutes registers the per-shop "What's New" advertisement endpoints:
// a customer read (user-authed) and admin CRUD (admin-authed, marketing RBAC).
// Purely additive — mounted under the existing /api group in server.go.
func ShopUpdateRoutes(api *gin.RouterGroup, mw middleware.Middleware, adminHandler handlerInterface.AdminHandler, h *handler.ShopUpdateHandler) {
	// Customer read: GET /api/shop-updates?lat=&long=&radius=&pincode=
	api.GET("/shop-updates", mw.AuthenticateUser(), h.GetShopUpdatesForUser)

	// Admin CRUD: /api/admin/shop-updates (+ product items)
	admin := api.Group("/admin/shop-updates",
		mw.AuthenticateAdmin(),
		adminHandler.RequirePermission(domain.PermCanManageMarketing),
	)
	{
		admin.POST("/", h.CreateShopUpdate)
		admin.GET("/", h.ListShopUpdates)
		admin.GET("/:id", h.GetShopUpdate)
		admin.PUT("/:id", h.UpdateShopUpdate)
		admin.DELETE("/:id", h.DeleteShopUpdate)

		admin.POST("/:id/products", h.AddProduct)
		admin.PUT("/:id/products/:productId", h.UpdateProduct)
		admin.DELETE("/:id/products/:productId", h.DeleteProduct)
	}
}
