package handler

import (
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	notificationSvc "github.com/rohit221990/mandi-backend/pkg/service/notification"
)

// fcmTokenPattern matches the characters FCM registration tokens are made of.
var fcmTokenPattern = regexp.MustCompile(`^[A-Za-z0-9:_\-]{100,400}$`)

// WebPushHandler enrols website visitors' browsers in the customer broadcast
// topic so admin-portal's "All Users" push reaches them.
type WebPushHandler struct {
	fcm *notificationSvc.FCMPushService
}

func NewWebPushHandler() *WebPushHandler {
	return &WebPushHandler{fcm: notificationSvc.NewFCMPushService()}
}

type webSubscribeRequest struct {
	Token string `json:"token" binding:"required"`
}

// Subscribe godoc
//
//	@Summary	Subscribe a website visitor's FCM token to the all_users topic
//	@Tags		Notification
//	@Router		/public/push/web-subscribe [post]
//	@Success	200	{object}	response.Response{}
func (h *WebPushHandler) Subscribe(ctx *gin.Context) {
	var req webSubscribeRequest
	if err := ctx.ShouldBindJSON(&req); err != nil || !fcmTokenPattern.MatchString(req.Token) {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid token", nil, nil)
		return
	}
	if err := h.fcm.SubscribeToTopic(ctx.Request.Context(), "all_users", []string{req.Token}); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to subscribe", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Subscribed")
}
