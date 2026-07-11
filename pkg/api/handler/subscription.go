package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	handlerInterface "github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	usecaseIface "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

type subscriptionHandler struct {
	subscriptionUseCase usecaseIface.SubscriptionUseCase
}

func NewSubscriptionHandler(uc usecaseIface.SubscriptionUseCase) handlerInterface.SubscriptionHandler {
	return &subscriptionHandler{subscriptionUseCase: uc}
}

// GetSubscriptionStatus godoc
//
//	@Summary		Get subscription status (User)
//	@Security		BearerAuth
//	@Description	Returns current subscription/trial status for the authenticated user
//	@Tags			User Subscription
//	@Id				GetSubscriptionStatus
//	@Router			/subscriptions/status [get]
//	@Success		200	{object}	response.Response{}	"Subscription status"
//	@Failure		500	{object}	response.Response{}	"Failed to get subscription status"
func (h *subscriptionHandler) GetSubscriptionStatus(ctx *gin.Context) {
	userID := utils.GetUserIdFromContext(ctx)

	result, err := h.subscriptionUseCase.GetSubscriptionStatus(ctx, userID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get subscription status", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Subscription status", result)
}

// StartTrial godoc
//
//	@Summary		Start free trial (User)
//	@Security		BearerAuth
//	@Description	Activates a 90-day free trial for the authenticated user
//	@Tags			User Subscription
//	@Id				StartTrial
//	@Router			/subscriptions/start-trial [post]
//	@Success		200	{object}	response.Response{}	"Free trial activated"
//	@Failure		409	{object}	response.Response{}	"Trial already used or active subscription exists"
//	@Failure		500	{object}	response.Response{}	"Failed to start trial"
func (h *subscriptionHandler) StartTrial(ctx *gin.Context) {
	userID := utils.GetUserIdFromContext(ctx)

	result, err := h.subscriptionUseCase.StartTrial(ctx, userID)
	if err != nil {
		errResponse(ctx, "Failed to start trial", err)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Free trial activated", result)
}

// GetPaidPlans godoc
//
//	@Summary		Get paid subscription plans (User)
//	@Security		BearerAuth
//	@Description	Returns available paid subscription plans (excludes free trial)
//	@Tags			User Subscription
//	@Id				GetPaidPlans
//	@Router			/subscriptions/plans [get]
//	@Success		200	{object}	response.Response{}	"Subscription plans"
//	@Failure		500	{object}	response.Response{}	"Failed to get plans"
func (h *subscriptionHandler) GetPaidPlans(ctx *gin.Context) {
	result, err := h.subscriptionUseCase.GetPaidPlans(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get subscription plans", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Subscription plans", result)
}

// GetBillingHistory godoc
//
//	@Summary		Get billing history (User)
//	@Security		BearerAuth
//	@Description	Returns the authenticated user's paid subscription payments (most recent first) with per-order GST-inclusive split
//	@Tags			User Subscription
//	@Id				GetBillingHistory
//	@Router			/subscriptions/billing-history [get]
//	@Success		200	{object}	response.Response{}	"Billing history"
//	@Failure		500	{object}	response.Response{}	"Failed to get billing history"
func (h *subscriptionHandler) GetBillingHistory(ctx *gin.Context) {
	userID := utils.GetUserIdFromContext(ctx)

	result, err := h.subscriptionUseCase.GetBillingHistory(ctx, userID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get billing history", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved billing history", result)
}
