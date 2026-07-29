package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler"
	"github.com/rohit221990/mandi-backend/pkg/api/middleware"
)

// LanguageRoutes registers the read-only selectable-languages endpoint used by
// the seller-app onboarding language picker. Purely additive.
func LanguageRoutes(api *gin.RouterGroup, mw middleware.Middleware, h *handler.LanguageHandler) {
	api.GET("/languages", mw.AuthenticateUser(), h.GetLanguages)
}
