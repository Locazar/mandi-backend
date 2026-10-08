package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	notificationSvc "github.com/rohit221990/mandi-backend/pkg/service/notification"
)

// AppUpdateConfigHandler exposes the seller/customer app force-update
// config (config_seller/app, config_customer/app in Firestore) to
// admin-portal, replacing what was previously a Firebase-console-only edit.
type AppUpdateConfigHandler struct{}

func NewAppUpdateConfigHandler() *AppUpdateConfigHandler {
	return &AppUpdateConfigHandler{}
}

func parseAppUpdatePlatform(ctx *gin.Context) (domain.AppUpdatePlatform, bool) {
	platform := domain.AppUpdatePlatform(ctx.Param("platform"))
	if !platform.IsValid() {
		response.ErrorResponse(ctx, http.StatusBadRequest, "platform must be 'seller' or 'customer'", nil, nil)
		return "", false
	}
	return platform, true
}

// GetConfig godoc
//
//	@Summary		Get the seller/customer app force-update config
//	@Security		BearerAuth
//	@Tags			Notification
//	@Param			platform	path	string	true	"seller or customer"
//	@Router			/admin/app-update-config/{platform} [get]
//	@Success		200	{object}	response.Response{}
func (h *AppUpdateConfigHandler) GetConfig(ctx *gin.Context) {
	platform, ok := parseAppUpdatePlatform(ctx)
	if !ok {
		return
	}
	cfg, err := notificationSvc.GetAppUpdateConfig(ctx.Request.Context(), platform)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch app update config", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "App update config fetched", cfg)
}

// UpdateConfig godoc
//
//	@Summary		Edit the seller/customer app force-update config
//	@Security		BearerAuth
//	@Tags			Notification
//	@Param			platform	path	string								true	"seller or customer"
//	@Param			input		body	request.UpdateAppUpdateConfig	true	"Config fields"
//	@Router			/admin/app-update-config/{platform} [put]
//	@Success		200	{object}	response.Response{}
func (h *AppUpdateConfigHandler) UpdateConfig(ctx *gin.Context) {
	platform, ok := parseAppUpdatePlatform(ctx)
	if !ok {
		return
	}
	var req request.UpdateAppUpdateConfig
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}
	cfg := domain.AppUpdateConfig{
		MinimumAppVersion: req.MinimumAppVersion,
		LatestVersion:     req.LatestVersion,
		ForceUpdate:       req.ForceUpdate,
		PlayStoreURL:      req.PlayStoreURL,
		AppStoreURL:       req.AppStoreURL,
	}
	if err := notificationSvc.SaveAppUpdateConfig(ctx.Request.Context(), platform, cfg); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to save app update config", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "App update config saved")
}
