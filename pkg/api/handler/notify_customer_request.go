package handler

import (
	"context"
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type NotifyCustomerRequestHandler struct {
	uc             interfaces.NotifyCustomerRequestUseCase
	notificationUC interfaces.NotificationUseCase
}

func NewNotifyCustomerRequestHandler(uc interfaces.NotifyCustomerRequestUseCase, notificationUC interfaces.NotificationUseCase) *NotifyCustomerRequestHandler {
	return &NotifyCustomerRequestHandler{uc: uc, notificationUC: notificationUC}
}

// Submit is called by the seller app when the seller taps "Notify Customer".
// Records the request to DB (once per enquiry) and always sends an FCM push to the customer.
// POST /api/admin/notify-customer-requests
func (h *NotifyCustomerRequestHandler) Submit(ctx *gin.Context) {
	var req struct {
		EnquiryID    string  `json:"enquiry_id" binding:"required"`
		ProductID    string  `json:"product_id"`
		ProductName  string  `json:"product_name"`
		ImageURL     string  `json:"image_url"`
		DealAmount   float64 `json:"deal_amount"`
		CustomerID   string  `json:"customer_id"`
		CustomerName string  `json:"customer_name"`
		SellerID     string  `json:"seller_id"`
		ShopID       string  `json:"shop_id"`
		Status       string  `json:"status"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}

	notifyReq := domain.NotifyCustomerRequest{
		EnquiryID:    req.EnquiryID,
		ProductID:    req.ProductID,
		ProductName:  req.ProductName,
		ImageURL:     req.ImageURL,
		DealAmount:   req.DealAmount,
		CustomerID:   req.CustomerID,
		CustomerName: req.CustomerName,
		SellerID:     req.SellerID,
		ShopID:       req.ShopID,
		Status:       req.Status,
	}

	// Record to DB. A duplicate (same enquiry) is treated as OK — the admin
	// record was already created on the first tap; subsequent taps still fire
	// the push to the customer.
	submitErr := h.uc.Submit(ctx.Request.Context(), notifyReq)
	if submitErr != nil && !strings.Contains(submitErr.Error(), "already submitted") {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to submit notification request", submitErr, nil)
		return
	}

	// Always push an FCM notification to the customer regardless of whether
	// this is the first or a repeated tap.
	if req.CustomerID != "" {
		body := "The seller has sent you an update about your deal."
		if req.ProductName != "" {
			switch req.Status {
			case "accepted":
				body = "Great news! Your deal for " + req.ProductName + " has been accepted. Visit the shop to collect it."
			case "rejected":
				body = "Update on your deal for " + req.ProductName + ". Please check your enquiry for details."
			default:
				body = "You have a new update on your deal for " + req.ProductName + "."
			}
		}
		_, pushErr := h.notificationUC.SendPushNotification(context.Background(), request.SendPushRequest{
			OwnerID:   req.CustomerID,
			OwnerType: "user",
			Title:     "Deal Update 🔔",
			Body:      body,
			EventType: "seller_notification",
			Data: map[string]string{
				"enquiry_id": req.EnquiryID,
				"action_url": "/inquiry?id=" + req.EnquiryID,
				"route":      "/inquiry?id=" + req.EnquiryID,
			},
		})
		if pushErr != nil {
			log.Printf("WARN [NotifyCustomerRequest] push failed for customer %s enquiry %s: %v", req.CustomerID, req.EnquiryID, pushErr)
		}
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Request sent. Customer will be notified.")
}

// List returns all notify-customer-requests for the admin portal.
// GET /api/admin/notify-customer-requests?actioned=false
func (h *NotifyCustomerRequestHandler) List(ctx *gin.Context) {
	var actioned *bool
	if v := ctx.Query("actioned"); v != "" {
		b := v == "true"
		actioned = &b
	}

	requests, err := h.uc.List(ctx.Request.Context(), actioned)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch notification requests", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Notification requests fetched", requests)
}

// MarkActioned marks a request as handled by admin.
// PATCH /api/admin/notify-customer-requests/:id/action
func (h *NotifyCustomerRequestHandler) MarkActioned(ctx *gin.Context) {
	id := ctx.Param("id")
	var body struct {
		Note string `json:"note"`
	}
	_ = ctx.ShouldBindJSON(&body)

	if err := h.uc.MarkActioned(ctx.Request.Context(), id, body.Note); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update request", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Request marked as actioned")
}
