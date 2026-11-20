package handler

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/service/token"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
)

type ShopActionLogHandler struct {
	svc          *usecase.ShopActionLogService
	tokenService token.TokenService
}

func NewShopActionLogHandler(svc *usecase.ShopActionLogService, ts token.TokenService) *ShopActionLogHandler {
	return &ShopActionLogHandler{svc: svc, tokenService: ts}
}

// Record accepts a shop interaction event from the customer web app.
// POST /api/public/shop/action-log — no auth required, always returns 204.
func (h *ShopActionLogHandler) Record(ctx *gin.Context) {
	var req struct {
		Action        string   `json:"action"`
		ShopID        string   `json:"shop_id"`
		ShopName      string   `json:"shop_name"`
		CustomerPhone string   `json:"customer_phone"`
		CustomerLat   *float64 `json:"customer_lat"`
		CustomerLon   *float64 `json:"customer_lon"`
		LocationLabel string   `json:"location_label"`
		SessionID     string   `json:"session_id"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil || req.Action == "" || req.ShopID == "" {
		ctx.Status(http.StatusNoContent)
		return
	}

	validActions := map[string]bool{
		domain.ShopActionCall:       true,
		domain.ShopActionWhatsApp:   true,
		domain.ShopActionDirections: true,
		domain.ShopActionCopyNumber: true,
	}
	if !validActions[req.Action] {
		ctx.Status(http.StatusNoContent)
		return
	}

	// Optionally decode customer ID from Bearer token — never block on failure.
	customerID := ""
	if authHeader := ctx.GetHeader("Authorization"); authHeader != "" {
		if tokenStr := strings.TrimPrefix(authHeader, "Bearer "); tokenStr != authHeader {
			customerID = h.tokenService.DecodeTokenData(tokenStr)
		}
	}

	entry := domain.ShopActionLog{
		Action:        req.Action,
		ShopID:        strings.TrimSpace(req.ShopID),
		ShopName:      strings.TrimSpace(req.ShopName),
		CustomerID:    customerID,
		CustomerPhone: strings.TrimSpace(req.CustomerPhone),
		CustomerLat:   req.CustomerLat,
		CustomerLon:   req.CustomerLon,
		LocationLabel: strings.TrimSpace(req.LocationLabel),
		SessionID:     strings.TrimSpace(req.SessionID),
	}

	_ = h.svc.Record(ctx.Request.Context(), entry)
	ctx.Status(http.StatusNoContent)
}

// List returns shop action logs for the admin portal.
// GET /api/admin/shop/action-logs?action=&shop_id=&seller_id=&from=&to=&limit=50&offset=0
func (h *ShopActionLogHandler) List(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(ctx.DefaultQuery("offset", "0"))

	filter := usecase.ShopActionLogFilter{
		Action:   ctx.Query("action"),
		ShopID:   ctx.Query("shop_id"),
		SellerID: ctx.Query("seller_id"),
		DateFrom: ctx.Query("from"),
		DateTo:   ctx.Query("to"),
		Limit:    limit,
		Offset:   offset,
	}

	logs, total, err := h.svc.List(ctx.Request.Context(), filter)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch action logs", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Shop action logs fetched", gin.H{
		"items":  logs,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}
