package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	usecaseInterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

// CustomerNudgeHandler exposes the customer onboarding-nudge templates,
// schedule, stats and manual sweep to admin-portal.
type CustomerNudgeHandler struct {
	uc usecaseInterfaces.CustomerNudgeUseCase
}

func NewCustomerNudgeHandler(uc usecaseInterfaces.CustomerNudgeUseCase) *CustomerNudgeHandler {
	return &CustomerNudgeHandler{uc: uc}
}

func (h *CustomerNudgeHandler) GetTemplates(ctx *gin.Context) {
	templates, err := h.uc.GetTemplates(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch customer nudge templates", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Customer nudge templates fetched", templates)
}

func (h *CustomerNudgeHandler) UpdateTemplate(ctx *gin.Context) {
	var req request.UpdateOnboardingNudgeTemplate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}
	tmpl := domain.CustomerNudgeTemplate{Key: ctx.Param("key"), Title: req.Title, Body: req.Body, ImageURL: req.ImageURL, Route: req.Route}
	if err := h.uc.SaveTemplate(ctx.Request.Context(), tmpl); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to update template", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Template updated")
}

func (h *CustomerNudgeHandler) GetSettings(ctx *gin.Context) {
	settings, err := h.uc.GetSettings(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch customer nudge settings", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Customer nudge settings fetched", settings)
}

func (h *CustomerNudgeHandler) UpdateSettings(ctx *gin.Context) {
	var req request.UpdateOnboardingNudgeSettings
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}
	settings := domain.CustomerNudgeSettings{Enabled: req.Enabled, GapHours: req.GapHours, DurationDays: req.DurationDays}
	if err := h.uc.SaveSettings(ctx.Request.Context(), settings); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to update settings", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Settings updated")
}

func (h *CustomerNudgeHandler) GetStats(ctx *gin.Context) {
	stats, err := h.uc.GetStats(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch customer nudge stats", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Customer nudge stats fetched", stats)
}

func (h *CustomerNudgeHandler) RunSweep(ctx *gin.Context) {
	result, err := h.uc.RunSweep(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Sweep failed", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Sweep complete", result)
}
