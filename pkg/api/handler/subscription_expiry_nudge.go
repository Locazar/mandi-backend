package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	usecaseInterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

// SubscriptionExpiryNudgeHandler exposes the subscription-expiry-nudge
// template, schedule settings, and manual sweep trigger to admin-portal.
type SubscriptionExpiryNudgeHandler struct {
	uc usecaseInterfaces.SubscriptionExpiryNudgeUseCase
}

func NewSubscriptionExpiryNudgeHandler(uc usecaseInterfaces.SubscriptionExpiryNudgeUseCase) *SubscriptionExpiryNudgeHandler {
	return &SubscriptionExpiryNudgeHandler{uc: uc}
}

// GetSettings godoc
//
//	@Summary	Get subscription-expiry-nudge schedule settings
//	@Security	BearerAuth
//	@Tags		Notification
//	@Router		/admin/subscription-expiry-nudges/settings [get]
//	@Success	200	{object}	response.Response{}
func (h *SubscriptionExpiryNudgeHandler) GetSettings(ctx *gin.Context) {
	settings, err := h.uc.GetSettings(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch settings", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Settings fetched", settings)
}

// UpdateSettings godoc
//
//	@Summary	Edit subscription-expiry-nudge schedule settings
//	@Security	BearerAuth
//	@Tags		Notification
//	@Param		input	body	request.UpdateSubscriptionExpiryNudgeSettings	true	"Settings"
//	@Router		/admin/subscription-expiry-nudges/settings [put]
//	@Success	200	{object}	response.Response{}
func (h *SubscriptionExpiryNudgeHandler) UpdateSettings(ctx *gin.Context) {
	var req request.UpdateSubscriptionExpiryNudgeSettings
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}
	settings := domain.SubscriptionExpiryNudgeSettings{
		Enabled:          req.Enabled,
		DaysBeforeExpiry: req.DaysBeforeExpiry,
	}
	if err := h.uc.SaveSettings(ctx.Request.Context(), settings); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to update settings", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Settings updated")
}

// GetTemplate godoc
//
//	@Summary	Get subscription-expiry-nudge template
//	@Security	BearerAuth
//	@Tags		Notification
//	@Router		/admin/subscription-expiry-nudges/template [get]
//	@Success	200	{object}	response.Response{}
func (h *SubscriptionExpiryNudgeHandler) GetTemplate(ctx *gin.Context) {
	tmpl, err := h.uc.GetTemplate(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch template", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Template fetched", tmpl)
}

// UpdateTemplate godoc
//
//	@Summary	Edit subscription-expiry-nudge template
//	@Security	BearerAuth
//	@Tags		Notification
//	@Param		input	body	request.UpdateSubscriptionExpiryNudgeTemplate	true	"Template fields"
//	@Router		/admin/subscription-expiry-nudges/template [put]
//	@Success	200	{object}	response.Response{}
func (h *SubscriptionExpiryNudgeHandler) UpdateTemplate(ctx *gin.Context) {
	var req request.UpdateSubscriptionExpiryNudgeTemplate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}
	tmpl := domain.SubscriptionExpiryNudgeTemplate{
		Title:    req.Title,
		Body:     req.Body,
		ImageURL: req.ImageURL,
		Route:    req.Route,
	}
	if err := h.uc.SaveTemplate(ctx.Request.Context(), tmpl); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to update template", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Template updated")
}

// RunSweep godoc
//
//	@Summary		Manually trigger one subscription-expiry-nudge sweep
//	@Description	Same sweep the daily in-process ticker runs. Synchronous. Idempotent.
//	@Security		BearerAuth
//	@Tags			Notification
//	@Router			/admin/subscription-expiry-nudges/run-sweep [post]
//	@Success		200	{object}	response.Response{}
func (h *SubscriptionExpiryNudgeHandler) RunSweep(ctx *gin.Context) {
	result, err := h.uc.RunSweep(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Sweep failed", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Sweep complete", result)
}
