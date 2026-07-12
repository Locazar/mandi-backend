package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/service/pincode"
)

// PincodeHandler exposes a read-only Indian PIN code lookup (city/district/
// state), used by the admin-portal to auto-fill shop address verification.
// It has no dependencies beyond its own HTTP client, so it's constructed
// directly in server.go rather than routed through the wire DI graph.
type PincodeHandler struct {
	client *pincode.Client
}

// NewPincodeHandler builds a PincodeHandler.
func NewPincodeHandler() *PincodeHandler {
	return &PincodeHandler{client: pincode.NewClient()}
}

// GetPincodeDetails godoc
// @Summary		Look up city/district/state for an Indian PIN code
// @Tags			pincode
// @Param			pincode	path	string	true	"6-digit Indian PIN code"
// @Success		200	{object}	pincode.Details
// @Failure		400	{object}	response.Response
// @Failure		404	{object}	response.Response
// @Failure		502	{object}	response.Response
// @Router			/pincode/{pincode} [get]
func (h *PincodeHandler) GetPincodeDetails(ctx *gin.Context) {
	pin := ctx.Param("pincode")

	details, err := h.client.Lookup(ctx.Request.Context(), pin)
	if err != nil {
		if appErr, ok := err.(*domain.AppError); ok {
			response.ErrorResponse(ctx, appErr.StatusCode, appErr.Message, appErr, nil)
			return
		}
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch pincode details", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Pincode details fetched successfully", details)
}
