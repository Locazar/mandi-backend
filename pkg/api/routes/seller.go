package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler"
	"github.com/rohit221990/mandi-backend/pkg/api/middleware"
)

// SellerRoutes registers seller-facing endpoints under /api/seller.
// All routes require a valid seller JWT (AuthenticateUser).
// This group is completely separate from the admin and user route groups.
func SellerRoutes(api *gin.RouterGroup, mw middleware.Middleware, jobHandler *handler.JobHandler) {
	seller := api.Group("/seller", mw.AuthenticateUser())
	{
		jobs := seller.Group("/jobs")
		{
			jobs.POST("", jobHandler.PostSellerJob())
			jobs.GET("", jobHandler.GetSellerJobs())
		}
	}
}
