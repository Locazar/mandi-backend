package handler

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/jinzhu/copier"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/service/token"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
	usecaseInterface "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type offerHandler struct {
	offerUseCase usecaseInterface.OfferUseCase
	tokenService token.TokenService
}

func NewOfferHandler(offerUseCase usecaseInterface.OfferUseCase, tokenService token.TokenService) interfaces.OfferHandler {
	return &offerHandler{
		offerUseCase: offerUseCase,
		tokenService: tokenService,
	}
}

// SaveOffer godoc
//
//	@Summary		Add offer (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to add an offer (Admin)
//	@Id				SaveOffer
//	@Tags			Admin Offers
//	@Param			input	body	request.Offer{}	true	"input field"
//	@Router			/admin/offers [post]
//	@Success		200	{object}	response.Response{}	"Successfully offer added"
//	@Failure		409	{object}	response.Response{}	"Offer already exist"
//	@Failure		400	{object}	response.Response{}	"Invalid inputs"
func (p *offerHandler) SaveOffer(ctx *gin.Context) {

	var body request.Offer

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	err := p.offerUseCase.SaveOffer(ctx, body)
	if err != nil {
		var statusCode int

		switch {
		case errors.Is(err, usecase.ErrOfferNameAlreadyExist):
			statusCode = http.StatusConflict
		case errors.Is(err, usecase.ErrInvalidOfferEndDate):
			statusCode = http.StatusBadRequest
		default:
			statusCode = http.StatusInternalServerError
		}
		response.ErrorResponse(ctx, statusCode, "Failed to add offer", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully offer added", nil)
}

// GetAllOffers godoc
//
//	@Summary		Get all offers (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to get all offers
//	@Id				GetAllOffers
//	@Tags			Admin Offers
//	@Param			page_number	query	int	false	"Page Number"
//	@Param			count		query	int	false	"Count"
//	@Router			/admin/offers [get]
//	@Success		200	{object}	response.Response{}	"Successfully found all promotions"
//	@Failure		500	{object}	response.Response{}	"Failed to get all promotions"
func (c *offerHandler) GetAllOffers(ctx *gin.Context) {
	pagination := request.GetPagination(ctx)

	offersAndPromotions, err := c.offerUseCase.FindAllOffers(ctx, pagination)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve offers", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Offers retrieved successfully", offersAndPromotions)
}

// RemoveOffer godoc
//
//	@summary		Remove offer (Admin)
//	@Security		BearerAuth
//	@Description	API admin to remove an offer
//	@Id				RemoveOffer
//	@Tags			Admin Offers
//	@Param			offer_id	path	int	true	"Offer ID"
//	@Router			/admin/offers/{offer_id} [delete]
//	@Success		200	{object}	response.Response{}	"successfully offer added"
//	@Failure		400	{object}	response.Response{}	"invalid input"
func (c *offerHandler) RemoveOffer(ctx *gin.Context) {

	offerID, err := request.GetParamAsUint(ctx, "product_item_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

	err = c.offerUseCase.RemoveOffer(ctx, offerID)

	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to remove offer", err, nil)
		return
	}

	response.SuccessResponse(ctx, 200, "successfully offer removed", nil)

}

// @Summary		Add category offer (Admin)
// @Security		BearerAuth
// @Description	API for admin to add an offer category
// @Id				SaveCategoryOffer
// @Tags			Admin Offers
// @Param			input	body	request.OfferCategory{}	true	"input field"
// @Router			/admin/offers/category [post]
// @Success		200	{object}	response.Response{}	"successfully offer added for category"
// @Failure		400	{object}	response.Response{}	"invalid input"
func (c *offerHandler) SaveCategoryOffer(ctx *gin.Context) {

	var body request.OfferCategory

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	err := c.offerUseCase.SaveCategoryOffer(ctx, body)
	if err != nil {
		var statusCode int
		switch {
		case errors.Is(err, usecase.ErrOfferAlreadyEnded):
			statusCode = http.StatusBadRequest
		case errors.Is(err, usecase.ErrCategoryOfferAlreadyExist):
			statusCode = http.StatusConflict
		default:
			statusCode = http.StatusInternalServerError
		}
		response.ErrorResponse(ctx, statusCode, "Failed to add offer", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully offer added for given category")
}

// GetAllCategoryOffers godoc
//
//	@Summary		Get all category offers (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to get all category offers
//	@Id				GetAllCategoryOffers
//	@Tags			Admin Offers
//	@Param			page_number	query	int	false	"Page Number"
//	@Param			count		query	int	false	"Count"
//	@Router			/admin/offers/category [get]
//	@Success		200	{object}	response.Response{}	"successfully got all offer_category"
//	@Failure		500	{object}	response.Response{}	"failed to get offers_category"
func (c *offerHandler) GetAllCategoryOffers(ctx *gin.Context) {

	pagination := request.GetPagination(ctx)

	offerCategories, err := c.offerUseCase.FindAllCategoryOffers(ctx, pagination)

	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get offer categories", err, nil)
		return
	}

	if len(offerCategories) == 0 {
		response.SuccessResponse(ctx, http.StatusOK, "No offer categories found", nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully found offers categories", offerCategories)
}

// RemoveCategoryOffer godoc
//
//	@Summary		Remove category offer (Admin)
//	@Security		BearerAuth
//	@Description	API admin to remove a offer from category
//	@Id				RemoveCategoryOffer
//	@Tags			Admin Offers
//	@Param			offer_category_id	path	int	true	"Offer Category ID"
//	@Router			/admin/offers/category/{offer_category_id} [delete]
//	@Success		200	{object}	response.Response{}	"successfully offer added for category"
//	@Failure		400	{object}	response.Response{}	"invalid input"
func (c *offerHandler) RemoveCategoryOffer(ctx *gin.Context) {

	offerCategoryID, err := request.GetParamAsUint(ctx, "offer_category_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

	err = c.offerUseCase.RemoveCategoryOffer(ctx, offerCategoryID)

	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to remove offer form category", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully offer removed from category")
}

// ChangeCategoryOffer godoc
//
//	@Summary		Change product offer (Admin)
//	@Security		BearerAuth
//	@Description	API admin to change product offer to another offer
//	@Id				ChangeCategoryOffer
//	@Tags			Admin Offers
//	@Param			input	body	request.UpdateCategoryOffer{}	true	"input field"
//	@Router			/admin/offers/category [patch]
//	@Success		200	{object}	response.Response{}	"successfully offer replaced for category"
//	@Failure		400	{object}	response.Response{}	"invalid input"
func (c *offerHandler) ChangeCategoryOffer(ctx *gin.Context) {

	var body request.UpdateCategoryOffer

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	err := c.offerUseCase.ChangeCategoryOffer(ctx, body.CategoryOfferID, body.OfferID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to change offer for given category offer", err, nil)
		return
	}

	response.SuccessResponse(ctx, 200, "Successfully offer changed for given category offer")
}

// SaveProductItemOffer godoc
//
//	@Summary		Add product offer (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to add an offer for product
//	@Id				SaveProductItemOffer
//	@Tags			Admin Offers
//	@Param			input	body	request.OfferProduct{}	true	"input field"
//	@Router			/admin/offers/products [post]
//	@Success		200	{object}	response.Response{}	"successfully offer added for product"
//	@Failure		400	{object}	response.Response{}	"invalid input"
func (c *offerHandler) SaveProductItemOffer(ctx *gin.Context) {

	var body request.OfferProduct
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	fmt.Printf("Body after binding: %+v\n", body)

	var offerProduct domain.OfferProduct
	copier.Copy(&offerProduct, &body)

	fmt.Printf("offerProduct after copying: %+v\n", offerProduct)

	err := c.offerUseCase.SaveProductItemOffer(ctx, offerProduct)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to add offer for given product", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully offer added to given product")
}

// GetAllProductsOffers godoc
//
//	@Summary		Get all product offers (Admin)
//	@Security		BearerAuth
//	@Description	API for admin to get all product offers
//	@Id				GetAllProductsOffers
//	@Tags			Admin Offers
//	@Param			page_number	query	int	false	"Page Number"
//	@Param			count		query	int	false	"Count"
//	@Router			/admin/offers/products [get]
//	@Success		200	{object}	response.Response{}	"successfully got all offers_categories"
//	@Failure		500	{object}	response.Response{}	"failed to get offer_products"
func (c *offerHandler) GetAllProductsOffers(ctx *gin.Context) {

	pagination := request.GetPagination(ctx)

	offersOfCategories, err := c.offerUseCase.FindAllProductOffers(ctx, pagination)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get all offer products", err, nil)
		return
	}

	if offersOfCategories == nil {
		response.SuccessResponse(ctx, http.StatusOK, "No offer products found", nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully found all offer products", offersOfCategories)
}

// RemoveProductOffer godoc
//
//	@Summary		Remove product offer (Admin)
//	@Security		BearerAuth
//	@Description	API admin to remove a offer from product
//	@Id				RemoveProductOffer
//	@Tags			Admin Offers
//	@param			offer_product_id	path	int	true	"offer_product_id"
//	@Router			/admin/offers/products/{offer_product_id} [delete]
//	@Success		200	{object}	response.Response{}	"Successfully offer removed from product"
//	@Failure		400	{object}	response.Response{}	"invalid input on params"
func (c *offerHandler) RemoveProductOffer(ctx *gin.Context) {

	offerProductID, err := request.GetParamAsUint(ctx, "offer_product_id")
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindParamFailMessage, err, nil)
		return
	}

	err = c.offerUseCase.RemoveProductOffer(ctx, offerProductID)

	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to remove offer form product", err, nil)
		return
	}

	response.SuccessResponse(ctx, 200, "Successfully offer removed from product")
}

// ChangeProductOffer godoc
//
//	@Summary		Change product offer (Admin)
//	@Security		BearerAuth
//	@Description	API admin to change product offer to another offer
//	@Id				ChangeProductOffer
//	@Tags			Admin Offers
//	@Param			input	body	request.UpdateProductOffer{}	true	"input field"
//	@Router			/admin/offers/products [patch]
//	@Success		200	{object}	response.Response{}	"Successfully offer changed for  given product offer"
//	@Failure		400	{object}	response.Response{}	"invalid input"
func (c *offerHandler) ChangeProductOffer(ctx *gin.Context) {

	var body request.UpdateProductOffer

	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}

	err := c.offerUseCase.ChangeProductOffer(ctx, body.ProductOfferID, body.OfferID)

	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to change offer for given product offer", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully offer changed for  given product offer")
}

func (c *offerHandler) ApplyOfferToShop(ctx *gin.Context) {
	var body request.ApplyOfferToShop
	if err := ctx.ShouldBindJSON(&body); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, BindJsonFailMessage, err, nil)
		return
	}
	tokenStr := ctx.GetHeader("Authorization")
	if tokenStr == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Authorization header missing", errors.New("authorization header missing"), nil)
		return
	}
	adminId := c.tokenService.DecodeTokenData(tokenStr)
	err := c.offerUseCase.ApplyOfferToShop(ctx, adminId, body)

	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Failed to apply offer to shop", err, nil)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully offer applied to shop", nil)
}

// GetActiveOffers returns currently active offers based on start and end date
func (c *offerHandler) GetActiveOffers(ctx *gin.Context) {
	offers, err := c.offerUseCase.FindActiveOffers(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get active offers", err, nil)
		return
	}
	if offers == nil || len(offers) == 0 {
		response.SuccessResponse(ctx, http.StatusOK, "No active offers found", offers)
		return
	}
	response.SuccessResponse(ctx, http.StatusOK, "Successfully found active offers", offers)
}

// GetShopOffers godoc
//
//	@Summary		Get shop offers by shop ID and date range (User)
//	@Security		BearerAuth
//	@Description	API for user to get shop offers within a date range
//	@Id				GetShopOffers
//	@Tags			User Offers
//	@Param			shop_id		query	uint	true	"Shop ID"
//	@Param			start_date	query	string	true	"Start Date (YYYY-MM-DD)"
//	@Param			end_date	query	string	true	"End Date (YYYY-MM-DD)"
//	@Router			/shop-offers [get]
//	@Success		200	{object}	response.Response{}	"Successfully retrieved shop offers"
//	@Failure		400	{object}	response.Response{}	"Invalid inputs"
func (c *offerHandler) GetShopOffers(ctx *gin.Context) {
	shopIDStr := ctx.Query("shop_id")
	startDate := ctx.Query("start_date")
	endDate := ctx.Query("end_date")

	if shopIDStr == "" || startDate == "" || endDate == "" {
		response.ErrorResponse(ctx, http.StatusBadRequest, "shop_id, start_date, and end_date are required", nil, nil)
		return
	}

	var shopID uint
	if _, err := fmt.Sscanf(shopIDStr, "%d", &shopID); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop_id", err, nil)
		return
	}

	shopOffers, err := c.offerUseCase.GetShopOffersByShopIDAndDateRange(ctx, shopID, startDate, endDate)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get shop offers", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Successfully retrieved shop offers", shopOffers)
}

// PostLoginOffer godoc
//
//	@Summary		Get post-login offer for user
//	@Security		BearerAuth
//	@Description	API to decide and return offer to show after login
//	@Id				PostLoginOffer
//	@Tags			User Offers
//	@Router			/user/post-login-offer [get]
//	@Success		200	{object}	response.PostLoginOfferResponse{}	"Offer decision"
//	@Failure		500	{object}	response.Response{}	"Internal server error"
func (c *offerHandler) PostLoginOffer(ctx *gin.Context) {
	// Get user ID from context (set by auth middleware)
	tokenStr := ctx.GetHeader("Authorization")
	if tokenStr == "" {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Authorization header missing", errors.New("authorization header missing"), nil)
		return
	}
	adminIdStr := c.tokenService.DecodeTokenData(tokenStr)
	adminId, err := strconv.ParseUint(adminIdStr, 10, 64)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusUnauthorized, "Invalid token", err, nil)
		return
	}

	offer, err := c.offerUseCase.GetPostLoginOffer(ctx, uint(adminId))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get post-login offer", err, nil)
		return
	}

	fmt.Printf("Post-login offer for user %d: %+v\n", adminId, offer)

	response.SuccessResponse(ctx, http.StatusOK, "Post-login offer retrieved", offer)
}

// GetBanners godoc
// @summary api to get banners
// @id GetBanners
// @tags Offer
// @Router /banner [get]
// @Success 200 {object} response.Response{} "successfully retrieved banners"
// @Failure 500 {object} response.Response{} "failed to get banners"
func (c *offerHandler) GetBanners(ctx *gin.Context) {
	banners, err := c.offerUseCase.GetBanners(ctx)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to get banners", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Banners retrieved successfully", banners)
}
