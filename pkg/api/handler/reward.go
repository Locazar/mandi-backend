package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

// RewardHandler serves the in-store rewards program to three audiences:
// customers (/api/rewards), sellers (/api/admin/rewards) and platform admins
// (/api/admin/rewards/program).
type RewardHandler struct {
	uc *usecase.RewardUseCase
}

func NewRewardHandler(uc *usecase.RewardUseCase) *RewardHandler {
	return &RewardHandler{uc: uc}
}

// ── customer ────────────────────────────────────────────────────────────────

// GetEligibility godoc
//
//	@Summary	Check whether the customer can record a purchase at a shop
//	@Security	BearerAuth
//	@Tags		User Rewards
//	@Param		shop_id	path	string	true	"Shop ID"
//	@Param		lat		query	number	true	"Customer latitude"
//	@Param		lng		query	number	true	"Customer longitude"
//	@Router		/rewards/shops/{shop_id}/eligibility [get]
//	@Success	200	{object}	response.Response{}
func (h *RewardHandler) GetEligibility(ctx *gin.Context) {
	lat, errLat := strconv.ParseFloat(ctx.Query("lat"), 64)
	lng, errLng := strconv.ParseFloat(ctx.Query("lng"), 64)
	if errLat != nil || errLng != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "lat and lng query parameters are required", errors.New("invalid coordinates"), nil)
		return
	}
	res, err := h.uc.GetEligibility(ctx, utils.GetUserIdFromContext(ctx), ctx.Param("shop_id"), lat, lng)
	if err != nil {
		errResponse(ctx, "Failed to check rewards eligibility", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully checked rewards eligibility", res)
}

// SubmitPurchase godoc
//
//	@Summary	Record an in-store purchase (customer)
//	@Security	BearerAuth
//	@Tags		User Rewards
//	@Param		input	body	request.SubmitShopPurchaseRequest	true	"Purchase"
//	@Router		/rewards/purchases [post]
//	@Success	201	{object}	response.Response{}
func (h *RewardHandler) SubmitPurchase(ctx *gin.Context) {
	var body request.SubmitShopPurchaseRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	p, err := h.uc.SubmitPurchase(ctx, usecase.SubmitPurchaseInput{
		CustomerID: utils.GetUserIdFromContext(ctx), ShopID: body.ShopID, ClientRequestID: body.ClientRequestID,
		BillAmountPaise: body.BillAmountPaise, Lat: *body.Latitude, Lng: *body.Longitude,
	})
	if err != nil {
		errResponse(ctx, "Failed to record purchase", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusCreated, "Purchase recorded, waiting for the shop to confirm", p)
}

// ListMyPurchases godoc
//
//	@Summary	List my in-store purchases (customer)
//	@Security	BearerAuth
//	@Tags		User Rewards
//	@Router		/rewards/purchases [get]
//	@Success	200	{object}	response.Response{}
func (h *RewardHandler) ListMyPurchases(ctx *gin.Context) {
	views, err := h.uc.ListCustomerPurchases(ctx, utils.GetUserIdFromContext(ctx), request.GetPagination(ctx))
	if err != nil {
		errResponse(ctx, "Failed to list purchases", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully listed purchases", views)
}

// ── seller ──────────────────────────────────────────────────────────────────

// GetSellerSettings godoc
//
//	@Summary	Get my shop's Locazar discount setting (seller)
//	@Security	BearerAuth
//	@Tags		Seller Rewards
//	@Router		/admin/rewards/settings [get]
//	@Success	200	{object}	response.Response{}
func (h *RewardHandler) GetSellerSettings(ctx *gin.Context) {
	s, err := h.uc.GetSellerSettings(ctx, utils.GetUserIdFromContext(ctx))
	if err != nil {
		errResponse(ctx, "Failed to load rewards settings", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully loaded rewards settings", s)
}

// UpdateSellerSettings godoc
//
//	@Summary	Opt in/out of the Locazar discount and pick the % (seller)
//	@Security	BearerAuth
//	@Tags		Seller Rewards
//	@Param		input	body	request.UpdateShopRewardSettingsRequest	true	"Settings"
//	@Router		/admin/rewards/settings [put]
//	@Success	200	{object}	response.Response{}
func (h *RewardHandler) UpdateSellerSettings(ctx *gin.Context) {
	var body request.UpdateShopRewardSettingsRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	s, err := h.uc.UpdateSellerSettings(ctx, utils.GetUserIdFromContext(ctx), *body.DiscountEnabled, body.DiscountPercent)
	if err != nil {
		errResponse(ctx, "Failed to save rewards settings", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully saved rewards settings", s)
}

// ListSellerPurchases godoc
//
//	@Summary	List Locazar purchases at my shop (seller)
//	@Security	BearerAuth
//	@Tags		Seller Rewards
//	@Param		status	query	string	false	"pending|claimed|rejected|expired"
//	@Router		/admin/rewards/purchases [get]
//	@Success	200	{object}	response.Response{}
func (h *RewardHandler) ListSellerPurchases(ctx *gin.Context) {
	views, err := h.uc.ListSellerPurchases(ctx, utils.GetUserIdFromContext(ctx),
		domain.ShopPurchaseStatus(ctx.Query("status")), request.GetPagination(ctx))
	if err != nil {
		errResponse(ctx, "Failed to list purchases", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully listed purchases", views)
}

// ClaimPurchase godoc
//
//	@Summary	Confirm a purchase and claim reward points (seller)
//	@Security	BearerAuth
//	@Tags		Seller Rewards
//	@Param		purchase_id	path	string	true	"Purchase ID"
//	@Router		/admin/rewards/purchases/{purchase_id}/claim [post]
//	@Success	200	{object}	response.Response{}
func (h *RewardHandler) ClaimPurchase(ctx *gin.Context) {
	p, err := h.uc.ClaimPurchase(ctx, utils.GetUserIdFromContext(ctx), ctx.Param("purchase_id"))
	if err != nil {
		errResponse(ctx, "Failed to claim rewards", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Rewards claimed", p)
}

// RejectPurchase godoc
//
//	@Summary	Reject a purchase that did not happen (seller)
//	@Security	BearerAuth
//	@Tags		Seller Rewards
//	@Param		purchase_id	path	string								true	"Purchase ID"
//	@Param		input		body	request.RejectShopPurchaseRequest	false	"Reason"
//	@Router		/admin/rewards/purchases/{purchase_id}/reject [post]
//	@Success	200	{object}	response.Response{}
func (h *RewardHandler) RejectPurchase(ctx *gin.Context) {
	var body request.RejectShopPurchaseRequest
	if ctx.Request.ContentLength > 0 {
		if err := ctx.ShouldBindJSON(&body); err != nil {
			response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
			return
		}
	}
	p, err := h.uc.RejectPurchase(ctx, utils.GetUserIdFromContext(ctx), ctx.Param("purchase_id"), body.Reason)
	if err != nil {
		errResponse(ctx, "Failed to reject purchase", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Purchase rejected", p)
}

// GetSellerAccount godoc
//
//	@Summary	My reward points summary (seller)
//	@Security	BearerAuth
//	@Tags		Seller Rewards
//	@Router		/admin/rewards/account [get]
//	@Success	200	{object}	response.Response{}
func (h *RewardHandler) GetSellerAccount(ctx *gin.Context) {
	s, err := h.uc.GetSellerAccount(ctx, utils.GetUserIdFromContext(ctx))
	if err != nil {
		errResponse(ctx, "Failed to load reward points", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully loaded reward points", s)
}

// ListSellerLedger godoc
//
//	@Summary	My reward points history (seller)
//	@Security	BearerAuth
//	@Tags		Seller Rewards
//	@Router		/admin/rewards/account/ledger [get]
//	@Success	200	{object}	response.Response{}
func (h *RewardHandler) ListSellerLedger(ctx *gin.Context) {
	entries, err := h.uc.ListSellerLedger(ctx, utils.GetUserIdFromContext(ctx), request.GetPagination(ctx))
	if err != nil {
		errResponse(ctx, "Failed to load points history", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully loaded points history", entries)
}

// ── platform admin ──────────────────────────────────────────────────────────

// GetProgramConfig godoc
//
//	@Summary	Get rewards program configuration (Admin)
//	@Security	BearerAuth
//	@Tags		Admin Rewards
//	@Router		/admin/rewards/program/config [get]
//	@Success	200	{object}	response.Response{}
func (h *RewardHandler) GetProgramConfig(ctx *gin.Context) {
	cfg, err := h.uc.GetProgramConfig(ctx)
	if err != nil {
		errResponse(ctx, "Failed to load rewards configuration", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully loaded rewards configuration", cfg)
}

// UpdateProgramConfig godoc
//
//	@Summary	Update rewards program configuration (Admin)
//	@Security	BearerAuth
//	@Tags		Admin Rewards
//	@Param		input	body	domain.RewardProgramConfig	true	"Full configuration"
//	@Router		/admin/rewards/program/config [put]
//	@Success	200	{object}	response.Response{}
func (h *RewardHandler) UpdateProgramConfig(ctx *gin.Context) {
	var body domain.RewardProgramConfig
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	cfg, err := h.uc.UpdateProgramConfig(ctx, utils.GetUserIdFromContext(ctx), body)
	if err != nil {
		errResponse(ctx, "Failed to update rewards configuration", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully updated rewards configuration", cfg)
}

// ListAllPurchases godoc
//
//	@Summary	List all Locazar in-store purchases (Admin)
//	@Security	BearerAuth
//	@Tags		Admin Rewards
//	@Param		shop_id		query	string	false	"Shop ID"
//	@Param		customer_id	query	string	false	"Customer ID"
//	@Param		status		query	string	false	"pending|claimed|rejected|expired"
//	@Router		/admin/rewards/program/purchases [get]
//	@Success	200	{object}	response.Response{}
func (h *RewardHandler) ListAllPurchases(ctx *gin.Context) {
	views, err := h.uc.ListAllPurchases(ctx, domain.ShopPurchaseFilter{
		ShopID: ctx.Query("shop_id"), CustomerID: ctx.Query("customer_id"),
		Status: domain.ShopPurchaseStatus(ctx.Query("status")),
	}, request.GetPagination(ctx))
	if err != nil {
		errResponse(ctx, "Failed to list purchases", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully listed purchases", views)
}

// AdjustAccount godoc
//
//	@Summary	Manually credit/debit a reward account (Admin)
//	@Security	BearerAuth
//	@Tags		Admin Rewards
//	@Param		account_id	path	string								true	"Reward account ID"
//	@Param		input		body	request.AdjustRewardAccountRequest	true	"Adjustment"
//	@Router		/admin/rewards/program/accounts/{account_id}/adjust [post]
//	@Success	200	{object}	response.Response{}
func (h *RewardHandler) AdjustAccount(ctx *gin.Context) {
	var body request.AdjustRewardAccountRequest
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	acct, err := h.uc.AdjustAccount(ctx, utils.GetUserIdFromContext(ctx), ctx.Param("account_id"), body.DeltaPoints, body.Reason)
	if err != nil {
		errResponse(ctx, "Failed to adjust reward points", err)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully adjusted reward points", acct)
}
