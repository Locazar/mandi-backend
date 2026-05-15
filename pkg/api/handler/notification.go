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
//	@Summary		Register FCM device token
//	@Security		BearerAuth
//	@ID				RegisterDeviceToken
//	@Tags			Notification
//	@Accept			json
//	@Produce		json
//	@Param			input	body	request.NotificationDeviceToken	true	"Device token payload"
//	@Router			/notifications/register-token [post]
//	@Success		200	{object}	response.Response{}	"Token registered"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		500	{object}	response.Response{}	"Internal server error"
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
//	@Summary		Unregister FCM device token
//	@Security		BearerAuth
//	@ID				UnregisterDeviceToken
//	@Tags			Notification
//	@Accept			json
//	@Produce		json
//	@Param			input	body	request.UnregisterDeviceToken	true	"Device token payload"
//	@Router			/notifications/unregister-token [delete]
//	@Success		200	{object}	response.Response{}	"Token unregistered"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		500	{object}	response.Response{}	"Internal server error"
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
//	@Summary		Send a direct FCM push notification
//	@Security		BearerAuth
//	@ID				SendPushNotification
//	@Tags			Notification
//	@Accept			json
//	@Produce		json
//	@Param			input	body	request.SendPushRequest	true	"Push notification payload"
//	@Router			/notifications/push [post]
//	@Success		200	{object}	response.Response{}	"Notification sent"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		500	{object}	response.Response{}	"Internal server error"
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
//	@Summary		Save a notification record
//	@Security		BearerAuth
//	@ID				SaveNotification
//	@Tags			Notification
//	@Accept			json
//	@Produce		json
//	@Param			input	body	request.Notification	true	"Notification payload"
//	@Router			/notifications [post]
//	@Success		201	{object}	response.Response{}	"Notification saved"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		500	{object}	response.Response{}	"Internal server error"
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
//	@Summary		Get notifications with filters and pagination
//	@Security		BearerAuth
//	@ID				GetNotificationsBy
//	@Tags			Notification
//	@Produce		json
//	@Param			filter		query	request.GetNotification	false	"Filters"
//	@Param			pagination	query	request.Pagination		false	"Pagination"
//	@Router			/notifications [get]
//	@Success		200	{object}	response.Response{}	"Notifications retrieved"
//	@Failure		400	{object}	response.Response{}	"Invalid query params"
//	@Failure		500	{object}	response.Response{}	"Internal server error"
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
//	@Summary		Mark a notification as read
//	@Security		BearerAuth
//	@ID				MarkNotificationAsRead
//	@Tags			Notification
//	@Param			notification_id	path	uint	true	"Notification ID"
//	@Produce		json
//	@Router			/notifications/{notification_id}/read [patch]
//	@Success		200	{object}	response.Response{}	"Marked as read"
//	@Failure		400	{object}	response.Response{}	"Invalid ID"
//	@Failure		500	{object}	response.Response{}	"Internal server error"
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
