package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler"
	handlerInterface "github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/middleware"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// QRCodeRoutes registers the dynamic QR redirect feature:
//   - a PUBLIC redirect (GET /r/:code) that scanners hit — no auth, this is the
//     link the QR encodes;
//   - admin CRUD + PNG under /api/admin/qr-codes, gated by the marketing RBAC.
//
// Purely additive — the root group is only used for the single /r/:code route.
func QRCodeRoutes(root gin.IRouter, api *gin.RouterGroup, mw middleware.Middleware, adminHandler handlerInterface.AdminHandler, h *handler.QRCodeHandler) {
	// Public: the scannable link. Kept short (/r/:code) so the QR is low-density.
	root.GET("/r/:code", h.Redirect)

	// Admin CRUD: /api/admin/qr-codes
	admin := api.Group("/admin/qr-codes",
		mw.AuthenticateAdmin(),
		adminHandler.RequirePermission(domain.PermCanManageMarketing),
	)
	{
		admin.POST("", h.CreateQRCode)
		admin.GET("", h.ListQRCodes)
		admin.GET("/:id", h.GetQRCode)
		admin.PUT("/:id", h.UpdateQRCode)
		admin.DELETE("/:id", h.DeleteQRCode)
		admin.GET("/:id/image.png", h.GetQRImage)
	}
}
