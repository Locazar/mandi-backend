package handler

import (
	"fmt"
	"net/http"

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
	go func() {
		if err := h.uc.HandleIncoming(ctx.Request.Context(), payload); err != nil {
			_ = err // already logged inside HandleIncoming
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
