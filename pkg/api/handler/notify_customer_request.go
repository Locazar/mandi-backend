package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type NotifyCustomerRequestHandler struct {
	uc interfaces.NotifyCustomerRequestUseCase
}

func NewNotifyCustomerRequestHandler(uc interfaces.NotifyCustomerRequestUseCase) *NotifyCustomerRequestHandler {
	return &NotifyCustomerRequestHandler{uc: uc}
}

// Submit is called by the seller app when the seller taps "Notify Customer".
// POST /api/admin/notify-customer-requests
func (h *NotifyCustomerRequestHandler) Submit(ctx *gin.Context) {
	var req struct {
		EnquiryID    string  `json:"enquiry_id" binding:"required"`
		ProductID    string  `json:"product_id"`
		ProductName  string  `json:"product_name"`
		ImageURL     string  `json:"image_url"`
		DealAmount   float64 `json:"deal_amount"`
		CustomerID   string  `json:"customer_id" binding:"required"`
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

	if err := h.uc.Submit(ctx.Request.Context(), notifyReq); err != nil {
		if strings.Contains(err.Error(), "already submitted") {
			response.ErrorResponse(ctx, http.StatusConflict, err.Error(), err, nil)
			return
		}
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to submit notification request", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Request sent to Locazar. Customer will be notified.")
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
