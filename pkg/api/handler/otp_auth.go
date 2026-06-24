// Create a handle for OTP send for admin signup using mobileAuthUseCase

package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"

	// mobileAuthUseCase defines the interface for mobile authentication use case
	"github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

// Handler struct that holds the mobileAuthUseCase
type Handler struct {
	mobileAuthUseCase interfaces.MobileAuthUseCase
}

// NewHandler creates a new Handler with the given mobileAuthUseCase
func NewHandler(mobileAuthUseCase interfaces.MobileAuthUseCase) *Handler {
	return &Handler{mobileAuthUseCase: mobileAuthUseCase}
}

// SellerSignUpOtpSend handles the OTP send request for seller signup
// @Summary		Send OTP for Seller Signup
// @Description	API for seller to send OTP for signup
// @Tags			Seller Authentication
// @Accept			json
// @Produce		json
// @Param			request	body		request.OTPLogin	true	"Phone number for OTP send"
// @Success		200		{object}	response.SuccessResponse
// @Failure		400		{object}	response.ErrorResponse
// @Failure		500		{object}	response.ErrorResponse
// @Router			/seller/signup/otp/send [post]
func (h *Handler) SellerSignUpOtpSend(ctx *gin.Context) {
	var body request.OTPLogin
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	if body.Phone == "" {
		err := errors.New("mobile number is required")
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	otpResp, err := h.mobileAuthUseCase.SendOTP(ctx, body.Phone, ctx.ClientIP(), ctx.Request.UserAgent())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to send OTP", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "OTP sent successfully", otpResp)
}

// SellerSignUpOtpVerify handles the OTP verification request for seller signup
// @Summary		Verify OTP for Seller Signup
// @Description	API for seller to verify OTP for signup
// @Tags			Seller Authentication
// @Accept			json
// @Produce		json
// @Param			request	body		request.VerifyOTPRequest	true	"Phone number and OTP to verify"
// @Success		200		{object}	response.SuccessResponse
// @Failure		400		{object}	response.ErrorResponse
// @Failure		500		{object}	response.ErrorResponse
// @Router			/seller/signup/otp/verify [post]
func (h *Handler) SellerSignUpOtpVerify(ctx *gin.Context) {
	var body request.VerifyOTPRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, body)
		return
	}

	verifyResp, err := h.mobileAuthUseCase.VerifyOTP(ctx, &body, ctx.ClientIP(), ctx.Request.UserAgent())
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to verify OTP", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "OTP verified successfully", verifyResp)
}