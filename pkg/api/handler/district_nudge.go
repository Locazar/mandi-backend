package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	usecaseInterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

// DistrictNudgeHandler exposes the district-growth-nudge template, schedule
// settings, and manual sweep trigger to admin-portal.
type DistrictNudgeHandler struct {
	uc usecaseInterfaces.DistrictNudgeUseCase
}

func NewDistrictNudgeHandler(uc usecaseInterfaces.DistrictNudgeUseCase) *DistrictNudgeHandler {
	return &DistrictNudgeHandler{uc: uc}
}

// GetTemplate godoc
//
//	@Summary		Get the district-growth-nudge template
//	@Security		BearerAuth
//	@Tags			Notification
//	@Router			/admin/district-nudges/template [get]
//	@Success		200	{object}	response.Response{}
func (h *DistrictNudgeHandler) GetTemplate(ctx *gin.Context) {
	tmpl, err := h.uc.GetTemplate(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch district nudge template", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "District nudge template fetched", tmpl)
}

// UpdateTemplate godoc
//
//	@Summary		Edit the district-growth-nudge template
//	@Security		BearerAuth
//	@Tags			Notification
//	@Param			input	body	request.UpdateDistrictNudgeTemplate	true	"Template fields"
//	@Router			/admin/district-nudges/template [put]
//	@Success		200	{object}	response.Response{}
func (h *DistrictNudgeHandler) UpdateTemplate(ctx *gin.Context) {
	var req request.UpdateDistrictNudgeTemplate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}
	tmpl := domain.DistrictNudgeTemplate{Title: req.Title, Body: req.Body, ImageURL: req.ImageURL, Route: req.Route}
	if err := h.uc.SaveTemplate(ctx.Request.Context(), tmpl); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to update template", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Template updated")
}

// GetSettings godoc
//
//	@Summary		Get the district-growth-nudge schedule settings
//	@Security		BearerAuth
//	@Tags			Notification
//	@Router			/admin/district-nudges/settings [get]
//	@Success		200	{object}	response.Response{}
func (h *DistrictNudgeHandler) GetSettings(ctx *gin.Context) {
	settings, err := h.uc.GetSettings(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch district nudge settings", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "District nudge settings fetched", settings)
}

// UpdateSettings godoc
//
//	@Summary		Edit the district-growth-nudge schedule (state, radius, on/off)
//	@Security		BearerAuth
//	@Tags			Notification
//	@Param			input	body	request.UpdateDistrictNudgeSettings	true	"Settings"
//	@Router			/admin/district-nudges/settings [put]
//	@Success		200	{object}	response.Response{}
func (h *DistrictNudgeHandler) UpdateSettings(ctx *gin.Context) {
	var req request.UpdateDistrictNudgeSettings
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}
	settings := domain.DistrictNudgeSettings{Enabled: req.Enabled, State: req.State, RadiusKm: req.RadiusKm}
	if err := h.uc.SaveSettings(ctx.Request.Context(), settings); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to update settings", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Settings updated")
}

// RunSweep godoc
//
//	@Summary		Manually run one district-growth-nudge sweep right now
//	@Description	Same sweep the daily in-process ticker runs. Synchronous: blocks
//	@Description	until the sweep finishes. Idempotent — the sent ledger means this
//	@Description	is always safe to run in addition to the daily automated sweep.
//	@Security		BearerAuth
//	@Tags			Notification
//	@Router			/admin/district-nudges/run-sweep [post]
//	@Success		200	{object}	response.Response{}
func (h *DistrictNudgeHandler) RunSweep(ctx *gin.Context) {
	result, err := h.uc.RunSweep(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Sweep failed", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Sweep complete", result)
}
