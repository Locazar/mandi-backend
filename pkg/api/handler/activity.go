package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/service/presence"
)

// GetAppActivity reports who is still using the apps and who uninstalled.
//
//	@Summary	Seller/customer app activity and install base
//	@Security	BearerAuth
//	@Tags		Admin
//	@Router		/admin/activity [get]
//	@Success	200	{object}	response.Response{}
func (c *adminHandler) GetAppActivity(ctx *gin.Context) {
	stats, err := presence.GetStats(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to load app activity", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "App activity", stats)
}
