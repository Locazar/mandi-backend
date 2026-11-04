package handler

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type ShopProductEnquiryHandler struct {
	uc interfaces.ShopProductEnquiryUseCase
}

func NewShopProductEnquiryHandler(uc interfaces.ShopProductEnquiryUseCase) *ShopProductEnquiryHandler {
	return &ShopProductEnquiryHandler{uc: uc}
}

// Submit handles a customer's enquiry from the shop profile page.
// POST /api/public/shop-product-enquiries  (no auth required)
func (h *ShopProductEnquiryHandler) Submit(ctx *gin.Context) {
	var req struct {
		ShopID        string `json:"shop_id" binding:"required"`
		ShopName      string `json:"shop_name"`
		EnquiryType   string `json:"enquiry_type" binding:"required"`
		Message       string `json:"message" binding:"required"`
		CustomerPhone string `json:"customer_phone"`
		CustomerCity  string `json:"customer_city"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}

	t := strings.TrimSpace(req.EnquiryType)
	if t != domain.EnquiryTypeProductAvailability && t != domain.EnquiryTypeCustomerQuery {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid enquiry_type", nil, nil)
		return
	}

	e := domain.ShopProductEnquiry{
		ShopID:        strings.TrimSpace(req.ShopID),
		ShopName:      strings.TrimSpace(req.ShopName),
		EnquiryType:   t,
		Message:       strings.TrimSpace(req.Message),
		CustomerPhone: strings.TrimSpace(req.CustomerPhone),
		CustomerCity:  strings.TrimSpace(req.CustomerCity),
	}

	if err := h.uc.Submit(ctx.Request.Context(), e); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to save enquiry", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Enquiry submitted. We'll be in touch.")
}

// List returns enquiries for the admin portal.
// GET /api/admin/shop-product-enquiries?shop_id=&actioned=false&type=
func (h *ShopProductEnquiryHandler) List(ctx *gin.Context) {
	shopID := ctx.Query("shop_id")
	enquiryType := ctx.Query("type")

	var actioned *bool
	if v := ctx.Query("actioned"); v != "" {
		b := v == "true"
		actioned = &b
	}

	items, err := h.uc.List(ctx.Request.Context(), shopID, actioned, enquiryType)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch enquiries", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Enquiries fetched", items)
}

// MarkActioned marks an enquiry as handled by admin.
// PATCH /api/admin/shop-product-enquiries/:id/action
func (h *ShopProductEnquiryHandler) MarkActioned(ctx *gin.Context) {
	id := ctx.Param("id")
	var body struct {
		Note string `json:"note"`
	}
	_ = ctx.ShouldBindJSON(&body)

	if err := h.uc.MarkActioned(ctx.Request.Context(), id, body.Note); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update enquiry", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Enquiry marked as actioned")
}
