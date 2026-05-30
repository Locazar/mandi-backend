package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	usecaseinterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

// AlertTemplateHandler handles alert template HTTP requests
type AlertTemplateHandler struct {
	templateUC usecaseinterfaces.AlertTemplateUseCase
	adminUC    usecaseinterfaces.AdminUseCase
}

// NewAlertTemplateHandler creates a new alert template handler
func NewAlertTemplateHandler(uc usecaseinterfaces.AlertTemplateUseCase, adminUC usecaseinterfaces.AdminUseCase) *AlertTemplateHandler {
	return &AlertTemplateHandler{
		templateUC: uc,
		adminUC:    adminUC,
	}
}

// extractSellerID extracts and validates seller ID from the Authorization header.
// Returns 0 and writes a 401 response if the token is missing, invalid, or unparseable.
func (h *AlertTemplateHandler) extractSellerID(ctx *gin.Context) string {
	tokenString := ctx.GetHeader("Authorization")
	if tokenString == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Unauthorized: missing token", nil, nil)
		return ""
	}
	adminId := h.adminUC.DecodeTokenData(tokenString)
	if adminId == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Unauthorized: invalid token", nil, nil)
		return ""
	}
	// adminId is already the string typed-prefix ID from the token
	if adminId == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Unauthorized: invalid seller ID in token", nil, nil)
		return ""
	}
	return adminId
}

// ListTemplates handles GET /api/admin/alert-templates
// @Summary List all alert templates (Admin)
// @Description Returns all alert templates
// @Tags Admin AlertTemplates
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{}
// @Failure 500 {object} response.Response{}
// @Router /admin/alert-templates [get]
func (h *AlertTemplateHandler) ListTemplates(ctx *gin.Context) {
	templates, err := h.templateUC.ListTemplates(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to list templates", err, nil)
		return
	}
	if templates == nil {
		templates = []map[string]interface{}{}
	}
	response.SuccessResponse(ctx, http.StatusOK, "Templates fetched successfully", templates)
}

// CreateTemplate handles POST /api/admin/alert-templates
// @Summary Create an alert template (Admin)
// @Description Creates a new alert template
// @Tags Admin AlertTemplates
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param inputs body map[string]interface{} true "Template data"
// @Success 201 {object} response.Response{}
// @Failure 400 {object} response.Response{}
// @Failure 500 {object} response.Response{}
// @Router /admin/alert-templates [post]
func (h *AlertTemplateHandler) CreateTemplate(ctx *gin.Context) {
	var body map[string]interface{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	key, ok := body["key"].(string)
	if !ok || key == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Field 'key' is required and must be a non-empty string", nil, nil)
		return
	}

	title, ok := body["title"].(string)
	if !ok || title == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Field 'title' is required and must be a non-empty string", nil, nil)
		return
	}

	if err := h.templateUC.CreateTemplate(ctx, body); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create template", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Template created successfully")
}

// UpdateTemplate handles PUT /api/admin/alert-templates/:key
// @Summary Update an alert template (Admin)
// @Description Updates an existing alert template by key
// @Tags Admin AlertTemplates
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param key path string true "Template key"
// @Param inputs body map[string]interface{} true "Template data"
// @Success 200 {object} response.Response{}
// @Failure 400 {object} response.Response{}
// @Failure 500 {object} response.Response{}
// @Router /admin/alert-templates/{key} [put]
func (h *AlertTemplateHandler) UpdateTemplate(ctx *gin.Context) {
	key := ctx.Param("key")
	if key == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Template key is required", nil, nil)
		return
	}

	var body map[string]interface{}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	if err := h.templateUC.UpdateTemplate(ctx, key, body); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update template", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Template updated successfully")
}

// DeleteTemplate handles DELETE /api/admin/alert-templates/:key
// @Summary Delete an alert template (Admin)
// @Description Deletes an alert template by key
// @Tags Admin AlertTemplates
// @Security BearerAuth
// @Produce json
// @Param key path string true "Template key"
// @Success 200 {object} response.Response{}
// @Failure 400 {object} response.Response{}
// @Failure 500 {object} response.Response{}
// @Router /admin/alert-templates/{key} [delete]
func (h *AlertTemplateHandler) DeleteTemplate(ctx *gin.Context) {
	key := ctx.Param("key")
	if key == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Template key is required", nil, nil)
		return
	}

	if err := h.templateUC.DeleteTemplate(ctx, key); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete template", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Template deleted")
}

// ToggleTemplate handles PATCH /api/admin/alert-templates/:key
// @Summary Toggle an alert template (Admin)
// @Description Enables or disables an alert template
// @Tags Admin AlertTemplates
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param key path string true "Template key"
// @Param inputs body map[string]interface{} true "enabled flag"
// @Success 200 {object} response.Response{}
// @Failure 400 {object} response.Response{}
// @Failure 500 {object} response.Response{}
// @Router /admin/alert-templates/{key} [patch]
func (h *AlertTemplateHandler) ToggleTemplate(ctx *gin.Context) {
	key := ctx.Param("key")
	if key == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Template key is required", nil, nil)
		return
	}

	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	if err := h.templateUC.ToggleTemplate(ctx, key, body.Enabled); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to toggle template", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Template toggled successfully")
}

// SeedDefaults handles POST /api/admin/alert-templates/seed
// @Summary Seed default templates (Admin)
// @Description Seeds the database with default alert templates
// @Tags Admin AlertTemplates
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{}
// @Failure 500 {object} response.Response{}
// @Router /admin/alert-templates/seed [post]
func (h *AlertTemplateHandler) SeedDefaults(ctx *gin.Context) {
	if err := h.templateUC.SeedDefaults(ctx); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to seed templates", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Templates seeded")
}

// GetFlows handles GET /api/alerts/flows
// @Summary Get onboarding flows for seller
// @Description Returns all active onboarding_step templates
// @Tags Seller AlertTemplates
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.Response{}
// @Failure 400 {object} response.Response{}
// @Failure 401 {object} response.Response{}
// @Failure 500 {object} response.Response{}
// @Router /alerts/flows [get]
func (h *AlertTemplateHandler) GetFlows(ctx *gin.Context) {
	sellerID := h.extractSellerID(ctx)
	if sellerID == "" {
		return
	}

	templates, err := h.templateUC.ListTemplates(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch flows", err, nil)
		return
	}

	// Filter to only onboarding_step templates
	flows := make([]map[string]interface{}, 0)
	for _, t := range templates {
		if templateType, ok := t["template_type"].(string); ok && templateType == "onboarding_step" {
			flows = append(flows, t)
		}
	}

	response.SuccessResponse(ctx, http.StatusOK, "Flows fetched successfully", flows)
}

// GetFlow handles GET /api/alerts/flows/:key
// @Summary Get a specific onboarding flow for a seller
// @Description Returns a specific flow and the seller's progress through it
// @Tags Seller AlertTemplates
// @Security BearerAuth
// @Produce json
// @Param key path string true "Flow key"
// @Success 200 {object} response.Response{}
// @Failure 400 {object} response.Response{}
// @Failure 401 {object} response.Response{}
// @Failure 500 {object} response.Response{}
// @Router /alerts/flows/{key} [get]
func (h *AlertTemplateHandler) GetFlow(ctx *gin.Context) {
	key := ctx.Param("key")
	if key == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Flow key is required", nil, nil)
		return
	}

	sellerID := h.extractSellerID(ctx)
	if sellerID == "" {
		return
	}

	flow, err := h.templateUC.GetFlow(ctx, sellerID, key)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.ErrorResponse(ctx, http.StatusNotFound, "Template not found", err, nil)
			return
		}
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch flow", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Flow fetched successfully", flow)
}

// GetTemplateForSeller handles GET /api/alerts/templates/:key
// @Summary Get a specific template for a seller
// @Description Returns a specific alert template with seller context
// @Tags Seller AlertTemplates
// @Security BearerAuth
// @Produce json
// @Param key path string true "Template key"
// @Success 200 {object} response.Response{}
// @Failure 400 {object} response.Response{}
// @Failure 401 {object} response.Response{}
// @Failure 500 {object} response.Response{}
// @Router /alerts/templates/{key} [get]
func (h *AlertTemplateHandler) GetTemplateForSeller(ctx *gin.Context) {
	key := ctx.Param("key")
	if key == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Template key is required", nil, nil)
		return
	}

	sellerID := h.extractSellerID(ctx)
	if sellerID == "" {
		return
	}

	template, err := h.templateUC.GetTemplateForSeller(ctx, sellerID, key)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			response.ErrorResponse(ctx, http.StatusNotFound, "Template not found", err, nil)
			return
		}
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch template", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Template fetched successfully", template)
}

// CompleteStep handles POST /api/alerts/flows/:key/step/:step_number/complete
// @Summary Complete a step in an onboarding flow
// @Description Marks a specific step in a flow as completed for a seller
// @Tags Seller AlertTemplates
// @Security BearerAuth
// @Produce json
// @Param key path string true "Flow key"
// @Param step_number path int true "Step number"
// @Success 200 {object} response.Response{}
// @Failure 400 {object} response.Response{}
// @Failure 401 {object} response.Response{}
// @Failure 500 {object} response.Response{}
// @Router /alerts/flows/{key}/step/{step_number}/complete [post]
func (h *AlertTemplateHandler) CompleteStep(ctx *gin.Context) {
	flowKey := ctx.Param("key")
	if flowKey == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Flow key is required", nil, nil)
		return
	}

	stepNumberStr := ctx.Param("step_number")
	stepNumber, err := strconv.Atoi(stepNumberStr)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid step number", err, nil)
		return
	}
	if stepNumber <= 0 {
		response.ErrorResponse(ctx, http.StatusBadRequest, "step_number must be >= 1", nil, nil)
		return
	}

	sellerID := h.extractSellerID(ctx)
	if sellerID == "" {
		return
	}

	if err := h.templateUC.CompleteStep(ctx, sellerID, flowKey, stepNumber); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to complete step", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Step completed successfully")
}
