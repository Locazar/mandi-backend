package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	usecaseInterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

// SubscriptionExpiredNudgeHandler exposes the already-expired-subscription
// nudge template, settings, and manual sweep trigger to admin-portal.
type SubscriptionExpiredNudgeHandler struct {
	uc usecaseInterfaces.SubscriptionExpiredNudgeUseCase
}

func NewSubscriptionExpiredNudgeHandler(uc usecaseInterfaces.SubscriptionExpiredNudgeUseCase) *SubscriptionExpiredNudgeHandler {
	return &SubscriptionExpiredNudgeHandler{uc: uc}
}

// GetSettings godoc
//
//	@Summary	Get subscription-expired-nudge schedule settings
//	@Security	BearerAuth
//	@Tags		Notification
//	@Router		/admin/subscription-expired-nudges/settings [get]
//	@Success	200	{object}	response.Response{}
func (h *SubscriptionExpiredNudgeHandler) GetSettings(ctx *gin.Context) {
	settings, err := h.uc.GetSettings(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch settings", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Settings fetched", settings)
}

// UpdateSettings godoc
//
//	@Summary	Edit subscription-expired-nudge schedule settings
//	@Security	BearerAuth
//	@Tags		Notification
//	@Param		input	body	request.UpdateSubscriptionExpiredNudgeSettings	true	"Settings"
//	@Router		/admin/subscription-expired-nudges/settings [put]
//	@Success	200	{object}	response.Response{}
func (h *SubscriptionExpiredNudgeHandler) UpdateSettings(ctx *gin.Context) {
	var req request.UpdateSubscriptionExpiredNudgeSettings
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}
	settings := domain.SubscriptionExpiredNudgeSettings{
		Enabled:       req.Enabled,
		FireHourIST:   req.FireHourIST,
		FireMinuteIST: req.FireMinuteIST,
	}
	if err := h.uc.SaveSettings(ctx.Request.Context(), settings); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to update settings", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Settings updated")
}

// GetTemplate godoc
//
//	@Summary	Get subscription-expired-nudge template
//	@Security	BearerAuth
//	@Tags		Notification
//	@Router		/admin/subscription-expired-nudges/template [get]
//	@Success	200	{object}	response.Response{}
func (h *SubscriptionExpiredNudgeHandler) GetTemplate(ctx *gin.Context) {
	tmpl, err := h.uc.GetTemplate(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch template", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Template fetched", tmpl)
}

// UpdateTemplate godoc
//
//	@Summary	Edit subscription-expired-nudge template
//	@Security	BearerAuth
//	@Tags		Notification
//	@Param		input	body	request.UpdateSubscriptionExpiredNudgeTemplate	true	"Template fields"
//	@Router		/admin/subscription-expired-nudges/template [put]
//	@Success	200	{object}	response.Response{}
func (h *SubscriptionExpiredNudgeHandler) UpdateTemplate(ctx *gin.Context) {
	var req request.UpdateSubscriptionExpiredNudgeTemplate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}
	tmpl := domain.SubscriptionExpiredNudgeTemplate{
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
//	@Summary		Manually trigger one subscription-expired-nudge sweep
//	@Description	Same sweep the daily in-process ticker runs. Synchronous. Idempotent.
//	@Security		BearerAuth
//	@Tags			Notification
//	@Router			/admin/subscription-expired-nudges/run-sweep [post]
//	@Success		200	{object}	response.Response{}
func (h *SubscriptionExpiredNudgeHandler) RunSweep(ctx *gin.Context) {
	result, err := h.uc.RunSweep(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Sweep failed", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Sweep complete", result)
}
