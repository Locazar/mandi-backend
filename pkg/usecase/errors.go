// Package usecase — this file defines sentinel errors for within-usecase business logic branching.
//
// These are plain error values, NOT domain.AppError — they carry no HTTP status code.
//
// Usage rule:
//   - Return these from usecases when a specific named failure condition occurs.
//   - In handlers, convert them to domain.AppError before calling response.ErrorResponse.
//   - Never return these directly from a handler.
//
// Conversion example in a handler:
//
//	result, err := h.useCase.GetUser(ctx, id)
//	if errors.Is(err, ErrUserNotExist) {
//	    response.ErrorResponse(ctx, http.StatusNotFound, "user not found", domain.NotFoundError("user not found"), nil)
//	    return
//	}
//
// Adding a new sentinel: append to this file, then add a conversion case in the relevant handler.
// Use domain.AppError directly (not a sentinel) if the handler needs to control the HTTP status
// without intermediate branching in the usecase.
package usecase

import "errors"

var (
	// login
	ErrEmptyLoginCredentials = errors.New("all login credentials are empty")
	ErrUserNotExist          = errors.New("user not exist with given login credentials")
	ErrUserNotVerified       = errors.New("user not verified")
	ErrUserBlocked           = errors.New("user blocked by admin")
	ErrWrongPassword         = errors.New("password doesn't match")
	// otp
	ErrOtpExpired = errors.New("otp session expired")
	ErrInvalidOtp = errors.New("invalid otp")

	// refresh token
	ErrInvalidRefreshToken    = errors.New("invalid refresh token")
	ErrRefreshSessionNotExist = errors.New("there is no refresh token session for this token")
	ErrRefreshSessionExpired  = errors.New("refresh token expired in session")
	ErrRefreshSessionBlocked  = errors.New("refresh token blocked in session")

	// signup
	ErrUserAlreadyExit = errors.New("user already exist")

	// cart
	ErrProductItemOutOfStock = errors.New("product is now out of stock")
	ErrCartItemAlreadyExist  = errors.New("product_item already exist on the cart")
	ErrCartItemNotExit       = errors.New("product_item not exist on cart")
	ErrEmptyCart             = errors.New("user cart is empty")

	ErrRequireMinimumCartItemQty = errors.New("update cart item qty can not less than 1")
	ErrInvalidCartItemUpdateQty  = errors.New("update cart item qty reached max limit")

	// admin
	ErrSameBlockStatus = errors.New("user block status already in given status")

	//category
	ErrCategoryAlreadyExist = errors.New("category already exist")

	// variation
	ErrVariationAlreadyExist       = errors.New("variation already exist")
	ErrVariationOptionAlreadyExist = errors.New("variation already exist")

	// product
	ErrProductAlreadyExist = errors.New("product already exist with this name")

	// product item
	ErrProductItemAlreadyExist = errors.New("product item already exist with this configuration")
	ErrNotEnoughVariations     = errors.New("not enough variation options for this product select one variation option from each variation")
	ErrProductLimitExceeded    = errors.New("product upload limit reached for this shop")

	// offer
	ErrOfferNameAlreadyExist = errors.New("offer already exist this name")
	ErrInvalidOfferEndDate   = errors.New("invalid offer end date")
	ErrOfferAlreadyEnded     = errors.New("offer already ended")

	ErrCategoryOfferAlreadyExist = errors.New("an offer already exist for this category")
	ErrProductOfferAlreadyExist  = errors.New("an offer already exist for this product")

	// order
	ErrOutOfStockOnCart = errors.New("cart is not valid for order out of stock is in cart")

	// wish list
	ErrExistWishListProductItem = errors.New("product item already exist on wish list")

	//  payment
	ErrBlockedPayment          = errors.New("selected payment is blocked by admin")
	ErrPaymentAmountReachedMax = errors.New("order total price reached payment method maximum amount")
	ErrPaymentNotApproved      = errors.New("payment not approved")

	// subscription
	ErrSubscriptionPlanNotFound    = errors.New("subscription plan not found")
	ErrSubscriptionPlanInactive    = errors.New("subscription plan is not active")
	ErrActiveSubscriptionExists    = errors.New("you already have an active subscription")
	ErrSubscriptionOrderNotFound   = errors.New("subscription order not found")
	ErrSubscriptionOrderNotOwned   = errors.New("subscription order does not belong to this user")
	ErrSubscriptionOrderNotCreated = errors.New("subscription order is not in 'created' state")
	ErrSubscriptionOrderExpired    = errors.New("subscription order has expired")
	ErrPaymentAmountMismatch       = errors.New("payment amount does not match order")
	ErrPaymentCurrencyMismatch     = errors.New("payment currency does not match order")
	ErrPaymentOrderMismatch        = errors.New("payment order_id does not match")
	ErrInvalidWebhookSignature     = errors.New("invalid webhook signature")
	ErrTrialAlreadyUsed            = errors.New("free trial has already been used")
	ErrTrialPlanNotFound           = errors.New("free trial plan not found in system")

	// brand
	ErrBrandAlreadyExist = errors.New("brand name already exist")

	// alert
	ErrShopNotFound = errors.New("seller has no shop registered")
)
