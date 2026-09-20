package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	usecaseInterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

// OnboardingNudgeHandler exposes the 4 editable onboarding-reminder
// templates and the schedule settings to admin-portal.
type OnboardingNudgeHandler struct {
	uc usecaseInterfaces.OnboardingNudgeUseCase
}

func NewOnboardingNudgeHandler(uc usecaseInterfaces.OnboardingNudgeUseCase) *OnboardingNudgeHandler {
	return &OnboardingNudgeHandler{uc: uc}
}

// GetTemplates godoc
//
//	@Summary		List the 4 onboarding-nudge templates
//	@Security		BearerAuth
//	@Tags			Notification
//	@Router			/admin/onboarding-nudges/templates [get]
//	@Success		200	{object}	response.Response{}
func (h *OnboardingNudgeHandler) GetTemplates(ctx *gin.Context) {
	templates, err := h.uc.GetTemplates(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch onboarding nudge templates", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Onboarding nudge templates fetched", templates)
}

// UpdateTemplate godoc
//
//	@Summary		Edit one onboarding-nudge template's title/body/image
//	@Security		BearerAuth
//	@Tags			Notification
//	@Param			key		path	string									true	"Template key"
//	@Param			input	body	request.UpdateOnboardingNudgeTemplate	true	"Template fields"
//	@Router			/admin/onboarding-nudges/templates/{key} [put]
//	@Success		200	{object}	response.Response{}
func (h *OnboardingNudgeHandler) UpdateTemplate(ctx *gin.Context) {
	key := ctx.Param("key")
	var req request.UpdateOnboardingNudgeTemplate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}
	tmpl := domain.OnboardingNudgeTemplate{Key: key, Title: req.Title, Body: req.Body, ImageURL: req.ImageURL, Route: req.Route}
	if err := h.uc.SaveTemplate(ctx.Request.Context(), tmpl); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to update template", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Template updated")
}

// GetSettings godoc
//
//	@Summary		Get the onboarding-nudge schedule settings
//	@Security		BearerAuth
//	@Tags			Notification
//	@Router			/admin/onboarding-nudges/settings [get]
//	@Success		200	{object}	response.Response{}
func (h *OnboardingNudgeHandler) GetSettings(ctx *gin.Context) {
	settings, err := h.uc.GetSettings(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch onboarding nudge settings", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Onboarding nudge settings fetched", settings)
}

// UpdateSettings godoc
//
//	@Summary		Edit the onboarding-nudge schedule (gap hours, duration, on/off)
//	@Security		BearerAuth
//	@Tags			Notification
//	@Param			input	body	request.UpdateOnboardingNudgeSettings	true	"Settings"
//	@Router			/admin/onboarding-nudges/settings [put]
//	@Success		200	{object}	response.Response{}
func (h *OnboardingNudgeHandler) UpdateSettings(ctx *gin.Context) {
	var req request.UpdateOnboardingNudgeSettings
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}
	settings := domain.OnboardingNudgeSettings{Enabled: req.Enabled, GapHours: req.GapHours, DurationDays: req.DurationDays}
	if err := h.uc.SaveSettings(ctx.Request.Context(), settings); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to update settings", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Settings updated")
}

// RunSweep godoc
//
//	@Summary		Manually run one onboarding-nudge sweep right now
//	@Description	Same sweep cmd/onboarding-nudges runs on its Cloud Scheduler cadence —
//	@Description	exposed here so an admin can trigger an out-of-band run (e.g. right after
//	@Description	editing a template, or before the scheduled job is set up) without shelling
//	@Description	into anything. Synchronous: blocks until the sweep finishes.
//	@Security		BearerAuth
//	@Tags			Notification
//	@Router			/admin/onboarding-nudges/run-sweep [post]
//	@Success		200	{object}	response.Response{}
func (h *OnboardingNudgeHandler) RunSweep(ctx *gin.Context) {
	result, err := h.uc.RunSweep(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Sweep failed", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Sweep complete", result)
}

// Backfill godoc
//
//	@Summary		Anchor every currently-active shop into the onboarding-nudge sequence
//	@Description	For shops that went live before this feature existed and so never got an
//	@Description	anchor from the approve action. Idempotent — already-anchored shops are
//	@Description	skipped. All anchored shops become due for their first nudge together on
//	@Description	the next sweep tick (a synchronized burst, not a trickle) — call this
//	@Description	deliberately, not routinely.
//	@Security		BearerAuth
//	@Tags			Notification
//	@Router			/admin/onboarding-nudges/backfill [post]
//	@Success		200	{object}	response.Response{}
func (h *OnboardingNudgeHandler) Backfill(ctx *gin.Context) {
	anchored, err := h.uc.BackfillActiveShops(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Backfill failed", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Backfill complete", gin.H{"anchored": anchored})
}

// GetStats godoc
//
//	@Summary		Onboarding-nudge delivery totals (all-time, today, per slot, last 7 days)
//	@Security		BearerAuth
//	@Tags			Notification
//	@Router			/admin/onboarding-nudges/stats [get]
//	@Success		200	{object}	response.Response{}
func (h *OnboardingNudgeHandler) GetStats(ctx *gin.Context) {
	stats, err := h.uc.GetStats(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch onboarding nudge stats", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Onboarding nudge stats fetched", stats)
}

// GetShopCounts godoc
//
//	@Summary		Per-shop onboarding-nudge delivery counts, by template type
//	@Security		BearerAuth
//	@Tags			Notification
//	@Param			shop_ids	query	string	true	"Comma-separated shop ids (max 500)"
//	@Router			/admin/onboarding-nudges/shop-counts [get]
//	@Success		200	{object}	response.Response{}
func (h *OnboardingNudgeHandler) GetShopCounts(ctx *gin.Context) {
	var shopIDs []string
	for _, id := range strings.Split(ctx.Query("shop_ids"), ",") {
		if id = strings.TrimSpace(id); id != "" {
			shopIDs = append(shopIDs, id)
		}
	}
	if len(shopIDs) > 500 {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Too many shop ids (max 500)", nil, nil)
		return
	}
	counts, err := h.uc.GetShopCounts(ctx.Request.Context(), shopIDs)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch shop nudge counts", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Shop nudge counts fetched", counts)
}
