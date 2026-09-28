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
// shopLinkFor looks up the given seller's own shop public URL, for
// substituting the {{shop_link}} template placeholder. Best-effort: returns
// "" if the seller has no shop yet or the lookup fails, so callers can leave
// the placeholder as-is rather than failing the whole alert response.
func (h *AlertHandler) shopLinkFor(ctx context.Context, sellerID string) string {
	shop, err := h.adminUseCase.GetShopByOwnerID(ctx, sellerID)
	if err != nil || shop.ID == "" {
		return ""
	}
	return utils.ShopPublicURL(shop.ID, shop.ShopName, shop.City)
}

// substitutePlaceholder recursively replaces {{shop_link}} in every string
// value of a decoded alert content tree (e.g. a "share" CTA's link, or a
// nested title/description) with the given link.
func substitutePlaceholder(v interface{}, link string) interface{} {
	switch val := v.(type) {
	case string:
		return strings.ReplaceAll(val, "{{shop_link}}", link)
	case map[string]interface{}:
		for k, sub := range val {
			val[k] = substitutePlaceholder(sub, link)
		}
		return val
	case []interface{}:
		for i, sub := range val {
			val[i] = substitutePlaceholder(sub, link)
		}
		return val
	default:
		return v
	}
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
		// Checked across content AND the template's own top-level title/description
		// (separate DB columns from content_schema, and easy to miss — an admin
		// typing {{shop_link}} into the plain "Description" field on the template
		// form, rather than into the content-schema editor, previously never got
		// substituted at all).
		title := alert.Title
		description := alert.Description
		if strings.Contains(alert.Content, "{{shop_link}}") ||
			strings.Contains(title, "{{shop_link}}") ||
			strings.Contains(description, "{{shop_link}}") {
			if link := h.shopLinkFor(ctx, sellerID); link != "" {
				content = substitutePlaceholder(content, link)
				title = strings.ReplaceAll(title, "{{shop_link}}", link)
				description = strings.ReplaceAll(description, "{{shop_link}}", link)
			}
		}

		alertResp := map[string]interface{}{
			"id":          alert.ID,
			"key":         alert.Key,
			"title":       title,
			"content":     content,
			"description": description,
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

	// Resolve the seller the same way GetSellerAlerts does. This used to read
	// ctx.Get("seller_id"), which AuthenticateUser never sets (it sets "userId"),
	// so every dismissal returned 401 and no dismissal was ever recorded.
	sellerID := h.sellerIDFrom(ctx)
	if sellerID == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Unauthorized: seller_id not found", nil, nil)
		return
	}

	err := h.alertUseCase.DismissAlert(ctx, sellerID, alertKey)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to dismiss alert", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Alert dismissed successfully", nil)
}

// MarkAlertShown records that the seller was actually shown this alert.
//
// This is what makes a template's `frequency` mean anything: ShouldShowAlert
// compares "now" against the last `shown` log entry, and GetLastAlertActionTimes
// only counts rows whose action is "shown". Without this endpoint nothing could
// ever write such a row, so every alert re-appeared on every app open regardless
// of whether it was configured once / daily / weekly.
//
// Idempotency is deliberately not enforced: a duplicate log row for the same day
// is harmless, because the frequency check only reads the most recent timestamp.
func (h *AlertHandler) MarkAlertShown(ctx *gin.Context) {
	alertKey := ctx.Param("key")
	if alertKey == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Alert key is required", nil, nil)
		return
	}

	sellerID := h.sellerIDFrom(ctx)
	if sellerID == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Unauthorized: seller_id not found", nil, nil)
		return
	}

	if err := h.alertUseCase.LogAlertView(ctx, sellerID, alertKey); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to record alert view", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Alert view recorded", nil)
}

// sellerIDFrom resolves the calling seller from the bearer token, which is the
// one approach that works across this group (the auth middleware stores
// "userId", not "seller_id").
func (h *AlertHandler) sellerIDFrom(ctx *gin.Context) string {
	return h.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
}
