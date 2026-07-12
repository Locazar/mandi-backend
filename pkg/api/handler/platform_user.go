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
	callerID := h.callerID(ctx)
	pagination := request.GetPagination(ctx)
	roleFilter := ctx.Query("role")
	admins, err := h.useCase.ListAdmins(ctx, callerID, roleFilter, pagination)
	if err != nil {
		h.handleAppErr(ctx, "Failed to list platform users", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved platform users", admins)
}

func (h *platformUserHandler) CreateAdmin(ctx *gin.Context) {
	callerID := h.callerID(ctx)
	var body domain.Admin
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	// Reject any create request that supplies the record's own primary key.
	if body.ID != "" {
		appErr := domain.ValidationError("id", "must not be provided when creating a new admin")
		response.ErrorResponse(ctx, appErr.StatusCode, appErr.Message, appErr, nil)
		return
	}

	otpID, err := h.useCase.CreateAdmin(ctx, callerID, body)
	if err != nil {
		h.handleAppErr(ctx, "Failed to create platform user", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusCreated, "Platform user created. They can log in with their email or phone number and password.", map[string]string{"otp_id": otpID})
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

func (h *platformUserHandler) UpdateAdmin(ctx *gin.Context) {
	callerID := h.callerID(ctx)
	targetID := ctx.Param("admin_id")
	var body domain.Admin
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	if err := h.useCase.UpdateAdmin(ctx, callerID, targetID, body); err != nil {
		h.handleAppErr(ctx, "Failed to update platform user", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Platform user updated successfully")
}

func (h *platformUserHandler) DeleteAdmin(ctx *gin.Context) {
	callerID := h.callerID(ctx)
	targetID := ctx.Param("admin_id")
	if err := h.useCase.DeleteAdmin(ctx, callerID, targetID); err != nil {
		h.handleAppErr(ctx, "Failed to delete platform user", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Platform user deleted successfully")
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
