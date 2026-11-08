package handler

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type LeadRequestHandler struct {
	uc interfaces.LeadRequestUseCase
}

func NewLeadRequestHandler(uc interfaces.LeadRequestUseCase) *LeadRequestHandler {
	return &LeadRequestHandler{uc: uc}
}

// Submit handles POST /api/public/lead-requests (no auth)
func (h *LeadRequestHandler) Submit(ctx *gin.Context) {
	var req request.SubmitLeadRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}

	source := ctx.GetHeader("Referer")

	lead, err := h.uc.Submit(ctx.Request.Context(), req.Phone, req.Requirement, req.City, source)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to submit request", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Request submitted successfully", lead)
}

// List handles GET /api/admin/lead-requests (admin auth)
func (h *LeadRequestHandler) List(ctx *gin.Context) {
	city := ctx.Query("city")
	status := ctx.Query("status")
	limit := 50
	offset := 0

	if l := ctx.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := ctx.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	leads, total, err := h.uc.List(ctx.Request.Context(), city, status, limit, offset)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch lead requests", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Lead requests fetched", gin.H{
		"leads":  leads,
		"total":  total,
		"limit":  limit,
		"offset": offset,
	})
}

// UpdateStatus handles PATCH /api/admin/lead-requests/:id/status (admin auth)
func (h *LeadRequestHandler) UpdateStatus(ctx *gin.Context) {
	idStr := ctx.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid lead request ID", err, nil)
		return
	}

	var req request.UpdateLeadStatusRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}

	if err := h.uc.UpdateStatus(ctx.Request.Context(), uint(id), req.Status, req.AdminNote); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to update status", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Status updated")
}
