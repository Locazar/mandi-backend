package handler

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	handlerInterface "github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	usecaseIface "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

type subscriptionPaymentHandler struct {
	subPaymentUseCase usecaseIface.SubscriptionPaymentUseCase
}

func NewSubscriptionPaymentHandler(uc usecaseIface.SubscriptionPaymentUseCase) handlerInterface.SubscriptionPaymentHandler {
	return &subscriptionPaymentHandler{subPaymentUseCase: uc}
}

// CreateSubscriptionOrder godoc
//
//	@Summary		Create subscription order (User)
//	@Security		BearerAuth
//	@Description	Creates a Razorpay order for a subscription plan
//	@Tags			User Subscription
//	@Id				CreateSubscriptionOrder
//	@Param			input	body	request.CreateSubscriptionOrderRequest	true	"Plan ID"
//	@Router			/subscriptions/create-order [post]
//	@Success		200	{object}	response.Response{}	"Successfully created subscription order"
//	@Failure		400	{object}	response.Response{}	"Invalid input"
//	@Failure		500	{object}	response.Response{}	"Failed to create subscription order"
func (h *subscriptionPaymentHandler) CreateSubscriptionOrder(ctx *gin.Context) {
	var body request.CreateSubscriptionOrderRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	userID := utils.GetUserIdFromContext(ctx)

	result, err := h.subPaymentUseCase.CreateSubscriptionOrder(ctx, userID, body)
	if err != nil {
		errResponse(ctx, "Failed to create subscription order", err)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully created subscription order", result)
}

// VerifySubscriptionPayment godoc
//
//	@Summary		Verify subscription payment (User)
//	@Security		BearerAuth
//	@Description	Verifies Razorpay payment for a subscription
//	@Tags			User Subscription
//	@Id				VerifySubscriptionPayment
//	@Param			input	body	request.VerifySubscriptionPaymentRequest	true	"Razorpay verification data"
//	@Router			/subscriptions/verify-payment [post]
//	@Success		200	{object}	response.Response{}	"Successfully verified subscription payment"
//	@Failure		402	{object}	response.Response{}	"Payment not approved"
//	@Failure		500	{object}	response.Response{}	"Failed to verify payment"
func (h *subscriptionPaymentHandler) VerifySubscriptionPayment(ctx *gin.Context) {
	var body request.VerifySubscriptionPaymentRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	userID := utils.GetUserIdFromContext(ctx)

	result, err := h.subPaymentUseCase.VerifySubscriptionPayment(ctx, userID, body)
	if err != nil {
		errResponse(ctx, "Failed to verify subscription payment", err)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully verified subscription payment", result)
}

// HandlePaymentFailure godoc
//
//	@Summary		Log payment failure (User)
//	@Security		BearerAuth
//	@Description	Logs a subscription payment failure for observability
//	@Tags			User Subscription
//	@Id				HandlePaymentFailure
//	@Param			input	body	request.PaymentFailureRequest	true	"Failure details"
//	@Router			/subscriptions/payment-failed [post]
//	@Success		200	{object}	response.Response{}	"Payment failure logged"
func (h *subscriptionPaymentHandler) HandlePaymentFailure(ctx *gin.Context) {
	var body request.PaymentFailureRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	userID := utils.GetUserIdFromContext(ctx)

	if err := h.subPaymentUseCase.HandlePaymentFailure(ctx, userID, body); err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to log payment failure", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Payment failure logged")
}

// HandleWebhook godoc
//
//	@Summary		Razorpay webhook (Public)
//	@Description	Receives Razorpay webhook events — no auth required
//	@Tags			Webhook
//	@Id				HandleWebhook
//	@Router			/webhook/razorpay [post]
//	@Success		200	{object}	response.Response{}	"Webhook processed"
//	@Failure		400	{object}	response.Response{}	"Invalid webhook signature"
func (h *subscriptionPaymentHandler) HandleWebhook(ctx *gin.Context) {
	signature := ctx.GetHeader("X-Razorpay-Signature")

	rawBody, err := io.ReadAll(ctx.Request.Body)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to read request body", err, nil)
		return
	}

	err = h.subPaymentUseCase.HandleWebhook(ctx, signature, rawBody)
	if err != nil {
		errResponse(ctx, "Webhook processing failed", err)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Webhook processed")
}
