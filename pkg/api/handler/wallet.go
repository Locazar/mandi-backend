package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

// GetUserWallet godoc
//
//	@Summary		Get user wallet  (User)
//	@Security		BearerAuth
//	@Description	API for user to get user wallet
//	@Id				GetUserWallet
//	@Tags			User Profile
//	@Router			/account/wallet [get]
//	@Success		200	{object}	response.Response{}	"Successfully retrieve user wallet"
//	@Failure		500	{object}	response.Response{}	"Failed to retrieve user wallet"
func (c *OrderHandler) GetUserWallet(ctx *gin.Context) {

	userID := utils.GetUserIdFromContext(ctx)

	wallet, err := c.orderUseCase.FindUserWallet(ctx, userID)

	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve user wallet", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieve user wallet", wallet)
}

// GetUserWalletTransactions godoc
//
//	@Summary		Get user wallet  (User)
//	@Security		BearerAuth
//	@Description	API for user to get user wallet transaction
//	@Id				GetUserWalletTransactions
//	@Tags			User Profile
//	@Router			/account/wallet/transactions [get]
//	@Success		200	{object}	response.Response{}	"Successfully retrieved user wallet transactions"
//	@Success		204	{object}	response.Response{}	"No wallet transaction for user"
//	@Failure		500	{object}	response.Response{}	"Failed to retrieve user wallet transactions"
func (c *OrderHandler) GetUserWalletTransactions(ctx *gin.Context) {

	userID := utils.GetUserIdFromContext(ctx)
	pagination := request.GetPagination(ctx)

	transactions, err := c.orderUseCase.FindUserWalletTransactions(ctx, userID, pagination)

	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve user wallet transactions", err, nil)
		return
	}

	if len(transactions) == 0 {
		response.SuccessResponse(ctx, http.StatusNoContent, "No user wallet transaction found", []interface{}{})
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved user wallet transactions", transactions)
}
