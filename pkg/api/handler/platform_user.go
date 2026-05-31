package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	usecaseInterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type platformUserHandler struct {
	useCase      usecaseInterfaces.PlatformUserUseCase
	adminUseCase usecaseInterfaces.AdminUseCase
}

func NewPlatformUserHandler(useCase usecaseInterfaces.PlatformUserUseCase, adminUseCase usecaseInterfaces.AdminUseCase) interfaces.PlatformUserHandler {
	return &platformUserHandler{useCase: useCase, adminUseCase: adminUseCase}
}

func (h *platformUserHandler) callerID(ctx *gin.Context) string {
	return h.adminUseCase.DecodeTokenData(ctx.GetHeader("Authorization"))
}

func (h *platformUserHandler) handleAppErr(ctx *gin.Context, fallbackMsg string, err error) {
	var appErr *domain.AppError
	if errors.As(err, &appErr) {
		response.ErrorResponse(ctx, appErr.StatusCode, appErr.Message, appErr, nil)
		return
	}
	response.ErrorResponse(ctx, http.StatusInternalServerError, fallbackMsg, err, nil)
}

func (h *platformUserHandler) ListAdmins(ctx *gin.Context) {
	pagination := request.GetPagination(ctx)
	admins, err := h.useCase.ListAdmins(ctx, pagination)
	if err != nil {
		h.handleAppErr(ctx, "Failed to list platform users", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved platform users", admins)
}

func (h *platformUserHandler) CreateAdmin(ctx *gin.Context) {
	var body domain.Admin
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	otpID, err := h.useCase.CreateAdmin(ctx, body)
	if err != nil {
		h.handleAppErr(ctx, "Failed to create platform user", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusCreated, "Admin account created. They can log in via OTP.", map[string]string{"otp_id": otpID})
}

func (h *platformUserHandler) UpdateAdminRole(ctx *gin.Context) {
	callerID := h.callerID(ctx)
	targetID := ctx.Param("admin_id")
	var body struct {
		Role domain.AdminRole `json:"role" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	if err := h.useCase.UpdateAdminRole(ctx, callerID, targetID, body.Role); err != nil {
		h.handleAppErr(ctx, "Failed to update role", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Role updated successfully")
}

func (h *platformUserHandler) DeactivateAdmin(ctx *gin.Context) {
	callerID := h.callerID(ctx)
	targetID := ctx.Param("admin_id")
	if err := h.useCase.DeactivateAdmin(ctx, callerID, targetID); err != nil {
		h.handleAppErr(ctx, "Failed to deactivate platform user", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Platform user deactivated successfully")
}
