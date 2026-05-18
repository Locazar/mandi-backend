package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	usecaseinterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

// AlertHandler handles alert-related HTTP requests
type AlertHandler struct {
	alertUseCase usecaseinterfaces.AlertUseCase
	adminUseCase usecaseinterfaces.AdminUseCase
}

// NewAlertHandler creates a new alert handler
func NewAlertHandler(alertUseCase usecaseinterfaces.AlertUseCase, adminUseCase usecaseinterfaces.AdminUseCase) *AlertHandler {
	return &AlertHandler{
		alertUseCase: alertUseCase,
		adminUseCase: adminUseCase,
	}
}

// GetSellerAlerts handles GET /api/v1/seller/alerts
// @Summary Get seller alerts
// @Description Fetch all dynamic alerts for the logged-in seller
// @Tags Alerts
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/seller/alerts [get]
func (h *AlertHandler) GetSellerAlerts(ctx *gin.Context) {
	// Extract seller_id from authentication context
	tokenString := ctx.GetHeader("Authorization")
	adminId := h.adminUseCase.DecodeTokenData(tokenString)
	ownerID, err := strconv.ParseUint(adminId, 10, 64)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid owner ID", err, nil)
		return
	}
	if ownerID == 0 {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Unauthorized: seller_id not found", nil, nil)
		return
	}

	// Get alerts
	alerts, err := h.alertUseCase.GetSellerAlerts(ctx, strconv.FormatUint(ownerID, 10))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch alerts", err, nil)
		return
	}

	// Convert to response format
	alertResponses := make([]map[string]interface{}, 0)
	for _, alert := range alerts {
		alertResp := map[string]interface{}{
			"id":          alert.ID,
			"key":         alert.Key,
			"title":       alert.Title,
			"content":     alert.Content,
			"description": alert.Description,
			"type":        alert.Type,
			"priority":    alert.Priority,
			"is_active":   alert.IsActive,
			"frequency":   alert.Frequency,
			"valid_from":  alert.ValidFrom,
			"valid_until": alert.ValidUntil,
			"metadata":    alert.Metadata,
			"actions": func() []map[string]interface{} {
				actions := make([]map[string]interface{}, 0, len(alert.Actions))
				for _, action := range alert.Actions {
					actions = append(actions, map[string]interface{}{
						"label":       action.Label,
						"action_type": action.ActionType,
						"action_url":  action.ActionURL,
						"payload":     action.Payload,
					})
				}
				return actions
			}(),
		}
		alertResponses = append(alertResponses, alertResp)
	}

	response.SuccessResponse(ctx, http.StatusOK, "Alerts fetched successfully", alertResponses)
}

// DismissAlert handles POST /api/v1/seller/alerts/:key/dismiss
// @Summary Dismiss an alert
// @Description Mark an alert as dismissed to prevent repeated display
// @Tags Alerts
// @Security BearerAuth
// @Param key path string true "Alert key"
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/v1/seller/alerts/{key}/dismiss [post]
func (h *AlertHandler) DismissAlert(ctx *gin.Context) {
	alertKey := ctx.Param("key")
	if alertKey == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Alert key is required", nil, nil)
		return
	}

	sellerID, exists := ctx.Get("seller_id")
	if !exists {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Unauthorized: seller_id not found", nil, nil)
		return
	}

	sellerIDStr, ok := sellerID.(string)
	if !ok {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid seller_id format", nil, nil)
		return
	}

	err := h.alertUseCase.DismissAlert(ctx, sellerIDStr, alertKey)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to dismiss alert", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Alert dismissed successfully", nil)
}
