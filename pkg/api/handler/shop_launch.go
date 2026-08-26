package handler

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
)

// PreviewShopLaunchAudience godoc
//
//	@Summary		Count the customers a shop-launch announcement would reach
//	@Description	Resolves how many nearby customers and devices are within radius_km of the shop's pinned location. Sends nothing.
//	@Security		BearerAuth
//	@ID				PreviewShopLaunchAudience
//	@Tags			Admin Notification
//	@Produce		json
//	@Param			shop_id		query		string	true	"Shop ID"
//	@Param			radius_km	query		number	true	"Radius in km (1-10)"
//	@Router			/admin/notifications/shop-launch/audience [get]
//	@Success		200			{object}	response.Response{}	"Audience resolved"
//	@Failure		400			{object}	response.Response{}	"Invalid input"
//	@Failure		404			{object}	response.Response{}	"Shop not found"
//	@Failure		500			{object}	response.Response{}	"Internal server error"
func (h *NotificationHandler) PreviewShopLaunchAudience(ctx *gin.Context) {
	shopID := ctx.Query("shop_id")
	if shopID == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "shop_id is required", nil, nil)
		return
	}
	radiusKm, err := strconv.ParseFloat(ctx.Query("radius_km"), 64)
	if err != nil || radiusKm < 1 || radiusKm > 10 {
		response.ErrorResponse(ctx, http.StatusBadRequest, "radius_km must be a number between 1 and 10", err, nil)
		return
	}

	result, err := h.notificationUsecase.PreviewShopLaunchAudience(ctx.Request.Context(), shopID, radiusKm)
	if err != nil {
		writeShopLaunchError(ctx, err, "Failed to resolve announcement audience")
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully resolved announcement audience", result)
}

// SendShopLaunchAnnouncement godoc
//
//	@Summary		Announce a newly opened shop to nearby customers
//	@Description	Pushes a "shop is now open" notification, with the shop's image, to every customer whose saved address is within radius_km (1-10) of the shop's pinned location, and records an in-app copy for each.
//	@Security		BearerAuth
//	@ID				SendShopLaunchAnnouncement
//	@Tags			Admin Notification
//	@Accept			json
//	@Produce		json
//	@Param			input	body		request.ShopLaunchAnnouncement	true	"Announcement"
//	@Router			/admin/notifications/shop-launch [post]
//	@Success		200		{object}	response.Response{}	"Announcement sent"
//	@Failure		400		{object}	response.Response{}	"Invalid input"
//	@Failure		404		{object}	response.Response{}	"Shop not found"
//	@Failure		500		{object}	response.Response{}	"Internal server error"
func (h *NotificationHandler) SendShopLaunchAnnouncement(ctx *gin.Context) {
	var req request.ShopLaunchAnnouncement
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}

	// Attribution for the in-app notification rows. Set by RequirePermission,
	// which every route in this group passes through.
	adminID := ctx.GetString("adminId")

	// Background context: a wide radius means thousands of tokens across many
	// FCM round trips, and the admin's browser navigating away must not cancel
	// a send that is already partly delivered.
	result, err := h.notificationUsecase.SendShopLaunchAnnouncement(context.Background(), adminID, req)
	if err != nil {
		writeShopLaunchError(ctx, err, "Failed to send shop launch announcement")
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully sent shop launch announcement", result)
}

// SendRadiusAnnouncement godoc
//
//	@Summary		Push a notification to every customer within a radius of a point
//	@Description	Raw-coordinate version of the shop-launch announcement. Targets customers whose saved address falls within radius_km of (latitude, longitude), and records an in-app copy for each.
//	@Security		BearerAuth
//	@ID				SendRadiusAnnouncement
//	@Tags			Admin Notification
//	@Accept			json
//	@Produce		json
//	@Param			input	body		request.RadiusAnnouncement	true	"Announcement"
//	@Router			/admin/notifications/radius [post]
//	@Success		200		{object}	response.Response{}	"Announcement sent"
//	@Failure		400		{object}	response.Response{}	"Invalid input"
//	@Failure		500		{object}	response.Response{}	"Internal server error"
func (h *NotificationHandler) SendRadiusAnnouncement(ctx *gin.Context) {
	var req request.RadiusAnnouncement
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}

	adminID := ctx.GetString("adminId")

	// Background context, for the same reason as the shop-launch send: a wide
	// radius outlives the admin's request.
	result, err := h.notificationUsecase.SendRadiusAnnouncement(context.Background(), adminID, req)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to send radius announcement", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully sent radius announcement", result)
}

// writeShopLaunchError maps the two "the admin needs to fix something" cases to
// 4xx so the portal can show an actionable message, and everything else to 500.
func writeShopLaunchError(ctx *gin.Context, err error, fallback string) {
	switch {
	case errors.Is(err, usecase.ErrAnnouncementShopNotFound):
		response.ErrorResponse(ctx, http.StatusNotFound, "Shop not found", err, nil)
	case errors.Is(err, usecase.ErrShopHasNoLocation):
		response.ErrorResponse(ctx, http.StatusBadRequest,
			"This shop has no pinned location yet, so nearby customers cannot be found. Set the shop's address on the map first.", err, nil)
	default:
		response.ErrorResponse(ctx, http.StatusInternalServerError, fallback, err, nil)
	}
}
