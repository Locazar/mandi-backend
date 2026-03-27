package handler

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	usecaseInterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

// NotificationHandler handles all notification-related HTTP endpoints.
type NotificationHandler struct {
	notificationUsecase usecaseInterfaces.NotificationUseCase
}

// NewNotificationHandler creates a new NotificationHandler.
func NewNotificationHandler(notificationUsecase usecaseInterfaces.NotificationUseCase) *NotificationHandler {
	return &NotificationHandler{notificationUsecase: notificationUsecase}
}

// RegisterDeviceToken godoc
//
// @SummaryRegister FCM device token
// @SecurityBearerAuth
// @IDRegisterDeviceToken
// @TagsNotification
// @Acceptjson
// @Producejson
// @Paraminputbodyrequest.NotificationDeviceTokentrue"Device token payload"
// @Router/notifications/register-token [post]
// @Success200{object}response.Response{}"Token registered"
// @Failure400{object}response.Response{}"Invalid input"
// @Failure500{object}response.Response{}"Internal server error"
func (h *NotificationHandler) RegisterDeviceToken(ctx *gin.Context) {
	var req request.NotificationDeviceToken
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}
	if err := h.notificationUsecase.RegisterDeviceToken(ctx.Request.Context(), req); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to register device token", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Device token registered successfully")
}

// UnregisterDeviceToken godoc
//
// @SummaryUnregister FCM device token
// @SecurityBearerAuth
// @IDUnregisterDeviceToken
// @TagsNotification
// @Acceptjson
// @Producejson
// @Paraminputbodyrequest.UnregisterDeviceTokentrue"Token to remove"
// @Router/notifications/unregister-token [delete]
// @Success200{object}response.Response{}"Token removed"
// @Failure400{object}response.Response{}"Invalid input"
// @Failure500{object}response.Response{}"Internal server error"
func (h *NotificationHandler) UnregisterDeviceToken(ctx *gin.Context) {
	var req request.UnregisterDeviceToken
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}
	if err := h.notificationUsecase.UnregisterDeviceToken(ctx.Request.Context(), req); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to unregister device token", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Device token unregistered successfully")
}

// SendPushNotification godoc
//
// @SummarySend a direct FCM push notification
// @SecurityBearerAuth
// @IDSendPushNotification
// @TagsNotification
// @Acceptjson
// @Producejson
// @Paraminputbodyrequest.SendPushRequesttrue"Push notification payload"
// @Router/notifications/push [post]
// @Success200{object}response.Response{}"Notification sent"
// @Failure400{object}response.Response{}"Invalid input"
// @Failure500{object}response.Response{}"Internal server error"
func (h *NotificationHandler) SendPushNotification(ctx *gin.Context) {
	var req request.SendPushRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}
	// Use Background context so notifications aren't cancelled if client disconnects.
	// This allows FCM to complete delivery even after the HTTP response is sent.
	if err := h.notificationUsecase.SendPushNotification(context.Background(), req); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to send push notification", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Push notification sent successfully")
}

// SaveNotification godoc
//
// @SummarySave a notification record
// @SecurityBearerAuth
// @IDSaveNotification
// @TagsNotification
// @Acceptjson
// @Producejson
// @Paraminputbodyrequest.Notificationtrue"Notification payload"
// @Router/notifications [post]
// @Success201{object}response.Response{}"Notification saved"
// @Failure400{object}response.Response{}"Invalid input"
// @Failure500{object}response.Response{}"Internal server error"
func (h *NotificationHandler) SaveNotification(ctx *gin.Context) {
	var req request.Notification
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}
	if err := h.notificationUsecase.SaveNotification(ctx.Request.Context(), req); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to save notification", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusCreated, "Notification saved successfully")
}

// GetNotificationsBy godoc
//
// @SummaryGet notifications with filters and pagination
// @SecurityBearerAuth
// @IDGetNotificationsBy
// @TagsNotification
// @Producejson
// @Paramfilterqueryrequest.GetNotificationfalse"Filters"
// @Parampaginationqueryrequest.Paginationfalse"Pagination"
// @Router/notifications [get]
// @Success200{object}response.Response{}"Notifications retrieved"
// @Failure400{object}response.Response{}"Invalid query params"
// @Failure500{object}response.Response{}"Internal server error"
func (h *NotificationHandler) GetNotificationsBy(ctx *gin.Context) {
	var filter request.GetNotification
	if err := ctx.ShouldBindQuery(&filter); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid query parameters", err, nil)
		return
	}
	if filter.UserID == 0 {
		if uid := utils.GetUserIdFromContext(ctx); uid != 0 {
			filter.UserID = uid
		}
	}
	pagination := request.GetPagination(ctx)
	notifications, err := h.notificationUsecase.GetNotificationsBy(ctx.Request.Context(), filter, pagination)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch notifications", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Notifications retrieved", notifications)
}

// MarkNotificationAsRead godoc
//
// @SummaryMark a notification as read
// @SecurityBearerAuth
// @IDMarkNotificationAsRead
// @TagsNotification
// @Paramnotification_idpathuinttrue"Notification ID"
// @Producejson
// @Router/notifications/{notification_id}/read [patch]
// @Success200{object}response.Response{}"Marked as read"
// @Failure400{object}response.Response{}"Invalid ID"
// @Failure500{object}response.Response{}"Internal server error"
func (h *NotificationHandler) MarkNotificationAsRead(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("notification_id"), 10, 32)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid notification ID", err, nil)
		return
	}
	if err := h.notificationUsecase.MarkNotificationAsRead(ctx.Request.Context(), uint(id)); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to mark notification as read", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Notification marked as read")
}

// GenerateFCMToken is a backward-compatible alias for RegisterDeviceToken.
func (h *NotificationHandler) GenerateFCMToken(ctx *gin.Context) {
	h.RegisterDeviceToken(ctx)
}
