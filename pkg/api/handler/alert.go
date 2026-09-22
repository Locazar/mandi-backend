package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	usecaseinterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
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
// substituteShopLink recursively replaces the {{shop_link}} placeholder in
// every string value of an alert's decoded content with the given seller's
// own shop public URL (e.g. a "share" CTA's link, or the title/description).
// Best-effort: if the shop lookup fails, the placeholder is left as-is
// rather than failing the whole alert response.
func (h *AlertHandler) substituteShopLink(ctx context.Context, sellerID string, content interface{}) interface{} {
	shop, err := h.adminUseCase.GetShopByOwnerID(ctx, sellerID)
	if err != nil || shop.ID == "" {
		return content
	}
	link := utils.ShopPublicURL(shop.ID, shop.ShopName, shop.City)
	var replace func(v interface{}) interface{}
	replace = func(v interface{}) interface{} {
		switch val := v.(type) {
		case string:
			return strings.ReplaceAll(val, "{{shop_link}}", link)
		case map[string]interface{}:
			for k, sub := range val {
				val[k] = replace(sub)
			}
			return val
		case []interface{}:
			for i, sub := range val {
				val[i] = replace(sub)
			}
			return val
		default:
			return v
		}
	}
	return replace(content)
}

func (h *AlertHandler) GetSellerAlerts(ctx *gin.Context) {
	// Extract seller_id from authentication context
	tokenString := ctx.GetHeader("Authorization")
	sellerID := h.adminUseCase.DecodeTokenData(tokenString)
	if sellerID == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Unauthorized: seller_id not found", nil, nil)
		return
	}

	// Get alerts
	alerts, err := h.alertUseCase.GetSellerAlerts(ctx, sellerID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch alerts", err, nil)
		return
	}

	// Filter by type query param if provided
	if alertType := ctx.Query("type"); alertType != "" {
		filtered := alerts[:0]
		for _, a := range alerts {
			if string(a.Type) == alertType {
				filtered = append(filtered, a)
			}
		}
		alerts = filtered
	}

	// Convert to response format
	alertResponses := make([]map[string]interface{}, 0)
	for _, alert := range alerts {
		// alert.Content stores the template's content_schema as a raw JSON
		// string; decode it so the client receives a JSON object (matching
		// the AlertContent shape) instead of an escaped string.
		var content interface{}
		if alert.Content != "" {
			if err := json.Unmarshal([]byte(alert.Content), &content); err != nil {
				content = nil
			}
		}
		// {{shop_link}} is a template placeholder (same convention as onboarding
		// nudges) — admin authors one template, each seller gets their own shop's
		// public link substituted in, rather than one fixed link for everyone.
		// Only looked up when actually used: most alert content never references it.
		if strings.Contains(alert.Content, "{{shop_link}}") {
			content = h.substituteShopLink(ctx, sellerID, content)
		}

		alertResp := map[string]interface{}{
			"id":          alert.ID,
			"key":         alert.Key,
			"title":       alert.Title,
			"content":     content,
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
