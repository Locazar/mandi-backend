package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	usecaseInterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

// VerificationNotificationHandler exposes the editable copy for the pushes
// VerifyShop/ApproveShop/RejectShop send to the seller.
type VerificationNotificationHandler struct {
	uc usecaseInterfaces.AdminUseCase
}

func NewVerificationNotificationHandler(uc usecaseInterfaces.AdminUseCase) *VerificationNotificationHandler {
	return &VerificationNotificationHandler{uc: uc}
}

// GetTemplates godoc
//
//	@Summary		List the shop document-verification notification templates
//	@Security		BearerAuth
//	@Tags			Notification
//	@Router			/admin/verification-notifications/templates [get]
//	@Success		200	{object}	response.Response{}
func (h *VerificationNotificationHandler) GetTemplates(ctx *gin.Context) {
	templates, err := h.uc.GetVerificationNotificationTemplates(ctx.Request.Context())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to fetch verification notification templates", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Verification notification templates fetched", templates)
}

// UpdateTemplate godoc
//
//	@Summary		Edit one verification-notification template's title/body/image/route
//	@Security		BearerAuth
//	@Tags			Notification
//	@Param			key		path	string											true	"Template key"
//	@Param			input	body	request.UpdateVerificationNotificationTemplate	true	"Template fields"
//	@Router			/admin/verification-notifications/templates/{key} [put]
//	@Success		200	{object}	response.Response{}
func (h *VerificationNotificationHandler) UpdateTemplate(ctx *gin.Context) {
	key := ctx.Param("key")
	var req request.UpdateVerificationNotificationTemplate
	if err := ctx.ShouldBindJSON(&req); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Validation failed", err, nil)
		return
	}
	tmpl := domain.VerificationNotificationTemplate{Key: key, Title: req.Title, Body: req.Body, ImageURL: req.ImageURL, Route: req.Route}
	if err := h.uc.SaveVerificationNotificationTemplate(ctx.Request.Context(), tmpl); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to update template", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Template updated")
}
