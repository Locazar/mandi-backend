package handler

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type WARelayHandler struct {
	uc          interfaces.WAChatUseCase
	verifyToken string // WA_WEBHOOK_VERIFY_TOKEN from config
}

func NewWARelayHandler(uc interfaces.WAChatUseCase, verifyToken string) *WARelayHandler {
	return &WARelayHandler{uc: uc, verifyToken: verifyToken}
}

// VerifyWebhook handles Meta's GET challenge for webhook subscription.
// GET /api/webhook/whatsapp
func (h *WARelayHandler) VerifyWebhook(ctx *gin.Context) {
	mode := ctx.Query("hub.mode")
	token := ctx.Query("hub.verify_token")
	challenge := ctx.Query("hub.challenge")

	if mode == "subscribe" && token == h.verifyToken {
		ctx.String(http.StatusOK, challenge)
		return
	}
	ctx.Status(http.StatusForbidden)
}

// ReceiveWebhook handles incoming messages forwarded by Meta.
// POST /api/webhook/whatsapp
func (h *WARelayHandler) ReceiveWebhook(ctx *gin.Context) {
	var payload domain.WAWebhookPayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		ctx.Status(http.StatusBadRequest)
		return
	}
	// Meta expects 200 quickly; process asynchronously.
	// Use context.Background() — the request context is cancelled once the
	// response is sent, which would kill every DB query in the goroutine.
	go func() {
		if err := h.uc.HandleIncoming(context.Background(), payload); err != nil {
			log.Printf("[wa_relay] webhook handling failed: %v", err)
		}
	}()
	ctx.Status(http.StatusOK)
}

// ListChats returns active WA relay chats for the admin.
// GET /api/admin/wa-chats?status=active&page=1
func (h *WARelayHandler) ListChats(ctx *gin.Context) {
	status := ctx.DefaultQuery("status", "")
	page := 1
	if p := ctx.Query("page"); p != "" {
		var n int
		if _, err := fmt.Sscanf(p, "%d", &n); err == nil && n > 0 {
			page = n
		}
	}

	chats, err := h.uc.ListChats(ctx.Request.Context(), status, page)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to list WA chats", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "WA chats", chats)
}

// Stats returns, per shop, how many customers reached the seller through the
// WhatsApp relay — the numbers to show a seller before asking for payment.
// GET /api/admin/wa-chats/stats?from=2026-11-01&to=2026-11-30 (IST dates,
// both inclusive; default: the last 30 days including today).
func (h *WARelayHandler) Stats(ctx *gin.Context) {
	ist := time.FixedZone("IST", 5*60*60+30*60)
	now := time.Now().In(ist)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, ist)
	from, to := today.AddDate(0, 0, -29), today

	if v := ctx.Query("from"); v != "" {
		d, err := time.ParseInLocation("2006-01-02", v, ist)
		if err != nil {
			response.ErrorResponse(ctx, http.StatusBadRequest, "from must be YYYY-MM-DD", err, nil)
			return
		}
		from = d
	}
	if v := ctx.Query("to"); v != "" {
		d, err := time.ParseInLocation("2006-01-02", v, ist)
		if err != nil {
			response.ErrorResponse(ctx, http.StatusBadRequest, "to must be YYYY-MM-DD", err, nil)
			return
		}
		to = d
	}
	if to.Before(from) {
		response.ErrorResponse(ctx, http.StatusBadRequest, "to must not be before from", nil, nil)
		return
	}

	stats, err := h.uc.ShopStats(ctx.Request.Context(), from, to.AddDate(0, 0, 1))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to load WhatsApp stats", err, nil)
		return
	}
	if stats == nil {
		stats = []domain.WAShopStat{}
	}
	response.SuccessResponse(ctx, http.StatusOK, "WhatsApp relay stats", gin.H{
		"from":  from.Format("2006-01-02"),
		"to":    to.Format("2006-01-02"),
		"shops": stats,
	})
}

// GetChat returns a single chat with its messages.
// GET /api/admin/wa-chats/:id
func (h *WARelayHandler) GetChat(ctx *gin.Context) {
	id := ctx.Param("id")
	chat, msgs, err := h.uc.GetChat(ctx.Request.Context(), id)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusNotFound, "Chat not found", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "WA chat", gin.H{
		"chat":     chat,
		"messages": msgs,
	})
}
