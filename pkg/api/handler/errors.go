package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
)

// sentinelStatus maps known usecase sentinel errors to HTTP status codes.
// Add new entries here as new sentinel errors are introduced in pkg/usecase/errors.go.
var sentinelStatus = map[error]int{
	// auth
	usecase.ErrEmptyLoginCredentials: http.StatusBadRequest,
	usecase.ErrUserNotExist:          http.StatusUnauthorized,
	usecase.ErrUserNotVerified:       http.StatusForbidden,
	usecase.ErrUserBlocked:           http.StatusForbidden,
	usecase.ErrWrongPassword:         http.StatusUnauthorized,
	// otp
	usecase.ErrOtpExpired: http.StatusUnauthorized,
	usecase.ErrInvalidOtp: http.StatusBadRequest,
	// refresh token
	usecase.ErrInvalidRefreshToken:    http.StatusUnauthorized,
	usecase.ErrRefreshSessionNotExist: http.StatusUnauthorized,
	usecase.ErrRefreshSessionExpired:  http.StatusUnauthorized,
	usecase.ErrRefreshSessionBlocked:  http.StatusForbidden,
	// signup
	usecase.ErrUserAlreadyExit: http.StatusConflict,
	// cart
	usecase.ErrProductItemOutOfStock:     http.StatusConflict,
	usecase.ErrCartItemAlreadyExist:      http.StatusConflict,
	usecase.ErrCartItemNotExit:           http.StatusBadRequest,
	usecase.ErrEmptyCart:                 http.StatusBadRequest,
	usecase.ErrRequireMinimumCartItemQty: http.StatusBadRequest,
	usecase.ErrInvalidCartItemUpdateQty:  http.StatusBadRequest,
	// admin
	usecase.ErrSameBlockStatus: http.StatusBadRequest,
	// category / variation / product
	usecase.ErrCategoryAlreadyExist:        http.StatusConflict,
	usecase.ErrVariationAlreadyExist:       http.StatusConflict,
	usecase.ErrVariationOptionAlreadyExist: http.StatusConflict,
	usecase.ErrProductAlreadyExist:         http.StatusConflict,
	usecase.ErrProductItemAlreadyExist:     http.StatusConflict,
	usecase.ErrNotEnoughVariations:         http.StatusBadRequest,
	// offer
	usecase.ErrOfferNameAlreadyExist:     http.StatusConflict,
	usecase.ErrInvalidOfferEndDate:       http.StatusBadRequest,
	usecase.ErrOfferAlreadyEnded:         http.StatusBadRequest,
	usecase.ErrCategoryOfferAlreadyExist: http.StatusConflict,
	usecase.ErrProductOfferAlreadyExist:  http.StatusConflict,
	// order
	usecase.ErrOutOfStockOnCart: http.StatusConflict,
	// wishlist
	usecase.ErrExistWishListProductItem: http.StatusConflict,
	// payment
	usecase.ErrBlockedPayment:          http.StatusForbidden,
	usecase.ErrPaymentAmountReachedMax: http.StatusBadRequest,
	usecase.ErrPaymentNotApproved:      http.StatusPaymentRequired,
	// subscription
	usecase.ErrSubscriptionPlanNotFound:    http.StatusNotFound,
	usecase.ErrSubscriptionPlanInactive:    http.StatusBadRequest,
	usecase.ErrActiveSubscriptionExists:    http.StatusConflict,
	usecase.ErrSubscriptionOrderNotFound:   http.StatusNotFound,
	usecase.ErrSubscriptionOrderNotOwned:   http.StatusForbidden,
	usecase.ErrSubscriptionOrderNotCreated: http.StatusBadRequest,
	usecase.ErrSubscriptionOrderExpired:    http.StatusBadRequest,
	usecase.ErrPaymentAmountMismatch:       http.StatusBadRequest,
	usecase.ErrPaymentCurrencyMismatch:     http.StatusBadRequest,
	usecase.ErrPaymentOrderMismatch:        http.StatusBadRequest,
	usecase.ErrInvalidWebhookSignature:     http.StatusBadRequest,
	usecase.ErrTrialAlreadyUsed:            http.StatusConflict,
	usecase.ErrTrialPlanNotFound:           http.StatusInternalServerError,
	// brand
	usecase.ErrBrandAlreadyExist: http.StatusConflict,
	// alert / shop
	usecase.ErrShopNotFound: http.StatusNotFound,
}

// errResponse writes a structured error response for usecase errors.
// Known sentinel errors are mapped to their correct HTTP status.
// Unknown errors default to 500 Internal Server Error.
func errResponse(ctx *gin.Context, message string, err error) {
	status := http.StatusInternalServerError
	for sentinel, code := range sentinelStatus {
		if errors.Is(err, sentinel) {
			status = code
			break
		}
	}
	response.ErrorResponse(ctx, status, message, err, nil)
}
