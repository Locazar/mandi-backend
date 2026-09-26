package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler"
	handlerInterface "github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/middleware"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// RewardRoutes registers the in-store rewards program. Customer routes accept
// only customer accounts — the usecase rejects ids not found in users, because
// AuthenticateUser also admits admin tokens.
func RewardRoutes(api *gin.RouterGroup, mw middleware.Middleware, adminHandler handlerInterface.AdminHandler, h *handler.RewardHandler) {
	customer := api.Group("/rewards", mw.AuthenticateUser())
	{
		customer.GET("/shops/:shop_id/eligibility", h.GetEligibility)
		customer.POST("/purchases", h.SubmitPurchase)
		customer.GET("/purchases", h.ListMyPurchases)
	}

	seller := api.Group("/admin/rewards", mw.AuthenticateAdmin())
	{
		seller.GET("/settings", h.GetSellerSettings)
		seller.PUT("/settings", h.UpdateSellerSettings)
		seller.GET("/purchases", h.ListSellerPurchases)
		seller.POST("/purchases/:purchase_id/claim", h.ClaimPurchase)
		seller.POST("/purchases/:purchase_id/reject", h.RejectPurchase)
		seller.GET("/account", h.GetSellerAccount)
		seller.GET("/account/ledger", h.ListSellerLedger)
	}

	program := api.Group("/admin/rewards/program", mw.AuthenticateAdmin(),
		adminHandler.RequirePermission(domain.PermCanManageSettings))
	{
		program.GET("/config", h.GetProgramConfig)
		program.PUT("/config", h.UpdateProgramConfig)
		program.GET("/purchases", h.ListAllPurchases)
		program.POST("/accounts/:account_id/adjust", h.AdjustAccount)
	}
}
