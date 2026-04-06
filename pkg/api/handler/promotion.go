package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	usecaseInterface "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type PromotionHandler struct {
	promotionUseCase usecaseInterface.PromotionUseCase
}

func NewPromotionHandler(promotionUseCase usecaseInterface.PromotionUseCase) *PromotionHandler {
	return &PromotionHandler{
		promotionUseCase: promotionUseCase,
	}
}

// GetAllPromotionCategories godoc
//
//	@Summary		Get all promotion categories
//	@Security		BearerAuth
//	@ID				GetAllPromotionCategories
//	@Tags			Admin Promotions
//	@Produce		json
//	@Param			pagination	query	request.Pagination	false	"Pagination"
//	@Router			/admin/promotions/categories/ [get]
//	@Success		200	{object}	response.Response{}	"Promotion categories retrieved"
//	@Failure		500	{object}	response.Response{}	"Internal server error"
func (h *PromotionHandler) GetAllPromotionCategories(ctx *gin.Context) {
	pagination := request.GetPagination(ctx)

	categories, err := h.promotionUseCase.FindAllPromotionCategories(ctx, pagination)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve promotion categories", err, nil)
		return
	}

	// Generate icon paths if not set
	for i := range categories {
		if categories[i].IconPath == "" {
			// Replace spaces with underscores and add .png extension
			iconName := strings.ReplaceAll(categories[i].Name, " ", "_")
			categories[i].IconPath = fmt.Sprintf("/uploads/promotions/loyalty/%s.png", iconName)
		}
	}

	response.SuccessResponse(ctx, http.StatusOK, "Promotion categories retrieved successfully", categories)
}

// GetPromotionCategoryByID godoc
//
//	@Summary		Get a promotion category by ID
//	@Security		BearerAuth
//	@ID				GetPromotionCategoryByID
//	@Tags			Admin Promotions
//	@Produce		json
//	@Param			category_id	path	int	true	"Category ID"
//	@Router			/admin/promotions/categories/{category_id} [get]
//	@Success		200	{object}	response.Response{}	"Promotion category retrieved"
//	@Failure		400	{object}	response.Response{}	"Invalid category ID"
//	@Failure		500	{object}	response.Response{}	"Internal server error"
func (h *PromotionHandler) GetPromotionCategoryByID(ctx *gin.Context) {
	categoryIDStr := ctx.Param("category_id")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid category ID", err, nil)
		return
	}

	category, err := h.promotionUseCase.FindPromotionCategoryByID(ctx, uint(categoryID))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve promotion category", err, nil)
		return
	}

	// Generate icon path if not set
	if category.IconPath == "" {
		// Replace spaces with underscores and add .png extension
		iconName := strings.ReplaceAll(category.Name, " ", "_")
		category.IconPath = fmt.Sprintf("/uploads/promotions/loyalty/%s.png", iconName)
	}

	response.SuccessResponse(ctx, http.StatusOK, "Promotion category retrieved successfully", category)
}

// GetAllPromotionTypes godoc
//
//	@Summary		Get all promotion types
//	@Security		BearerAuth
//	@ID				GetAllPromotionTypes
//	@Tags			Admin Promotions
//	@Produce		json
//	@Param			pagination	query	request.Pagination	false	"Pagination"
//	@Router			/admin/promotions/types/ [get]
//	@Success		200	{object}	response.Response{}	"Promotion types retrieved"
//	@Failure		500	{object}	response.Response{}	"Internal server error"
func (h *PromotionHandler) GetAllPromotionTypes(ctx *gin.Context) {
	pagination := request.GetPagination(ctx)

	types, err := h.promotionUseCase.FindAllPromotionTypes(ctx, pagination)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve promotion types", err, nil)
		return
	}

	// Generate icon paths if not set
	for i := range types {
		if types[i].IconPath == "" && types[i].CategoryName != "" {
			// Replace spaces with underscores for both category and type names
			categoryName := strings.ReplaceAll(types[i].CategoryName, " ", "_")
			typeName := strings.ReplaceAll(types[i].Name, " ", "_")
			types[i].IconPath = fmt.Sprintf("/uploads/promotions/%s/%s.png", categoryName, typeName)
		}
	}

	response.SuccessResponse(ctx, http.StatusOK, "Promotion types retrieved successfully", types)
}

// GetPromotionTypesByCategoryID godoc
//
//	@Summary		Get promotion types by category ID
//	@Security		BearerAuth
//	@ID				GetPromotionTypesByCategoryID
//	@Tags			Admin Promotions
//	@Produce		json
//	@Param			category_id	path	int					true	"Category ID"
//	@Param			pagination	query	request.Pagination	false	"Pagination"
//	@Router			/admin/promotions/types/category/{category_id} [get]
//	@Success		200	{object}	response.Response{}	"Promotion types retrieved"
//	@Failure		400	{object}	response.Response{}	"Invalid category ID"
//	@Failure		500	{object}	response.Response{}	"Internal server error"
func (h *PromotionHandler) GetPromotionTypesByCategoryID(ctx *gin.Context) {
	categoryIDStr := ctx.Param("category_id")
	categoryID, err := strconv.ParseUint(categoryIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid category ID", err, nil)
		return
	}

	pagination := request.GetPagination(ctx)

	types, err := h.promotionUseCase.FindPromotionTypesByCategoryID(ctx, uint(categoryID), pagination)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve promotion types by category", err, nil)
		return
	}

	// Generate icon paths if not set
	for i := range types {
		if types[i].IconPath == "" && types[i].CategoryName != "" {
			// Replace spaces with underscores for both category and type names
			categoryName := strings.ReplaceAll(types[i].CategoryName, " ", "_")
			typeName := strings.ReplaceAll(types[i].Name, " ", "_")
			types[i].IconPath = fmt.Sprintf("/uploads/promotions/%s/%s.png", categoryName, typeName)
		}
	}

	response.SuccessResponse(ctx, http.StatusOK, "Promotion types retrieved successfully", types)
}

// GetPromotionTypeByID godoc
//
//	@Summary		Get a promotion type by ID
//	@Security		BearerAuth
//	@ID				GetPromotionTypeByID
//	@Tags			Admin Promotions
//	@Produce		json
//	@Param			type_id	path	int	true	"Type ID"
//	@Router			/admin/promotions/types/{type_id} [get]
//	@Success		200	{object}	response.Response{}	"Promotion type retrieved"
//	@Failure		400	{object}	response.Response{}	"Invalid type ID"
//	@Failure		500	{object}	response.Response{}	"Internal server error"
func (h *PromotionHandler) GetPromotionTypeByID(ctx *gin.Context) {
	typeIDStr := ctx.Param("type_id")
	typeID, err := strconv.ParseUint(typeIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid type ID", err, nil)
		return
	}

	promotionType, err := h.promotionUseCase.FindPromotionTypeByID(ctx, uint(typeID))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve promotion type", err, nil)
		return
	}

	// Generate icon path if not set
	if promotionType.IconPath == "" && promotionType.CategoryName != "" {
		// Replace spaces with underscores for both category and type names
		categoryName := strings.ReplaceAll(promotionType.CategoryName, " ", "_")
		typeName := strings.ReplaceAll(promotionType.Name, " ", "_")
		promotionType.IconPath = fmt.Sprintf("/uploads/promotions/%s/%s.png", categoryName, typeName)
	}

	response.SuccessResponse(ctx, http.StatusOK, "Promotion type retrieved successfully", promotionType)
}

// CreatePromotion godoc
//
//	@Summary		Create a new promotion
//	@Security		BearerAuth
//	@ID				CreatePromotion
//	@Tags			Admin Promotions
//	@Accept			json
//	@Produce		json
//	@Param			input	body	request.CreatePromotionRequest	true	"Promotion details"
//	@Router			/admin/promotions/ [post]
//	@Success		201	{object}	response.Response{}	"Promotion created"
//	@Failure		400	{object}	response.Response{}	"Invalid request body"
//	@Failure		500	{object}	response.Response{}	"Internal server error"
func (h *PromotionHandler) CreatePromotion(ctx *gin.Context) {
	var reqBody request.CreatePromotionRequest

	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err, nil)
		return
	}

	// Parse string IDs to uint
	shopID, err := strconv.ParseUint(reqBody.ShopID, 10, 32)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid shop_id", err, nil)
		return
	}

	categoryID, err := strconv.ParseUint(reqBody.PromotionCategoryID, 10, 32)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid promotion_category_id", err, nil)
		return
	}

	typeID, err := strconv.ParseUint(reqBody.PromotionTypeID, 10, 32)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid promotion_type_id", err, nil)
		return
	}

	// Construct offer_details struct
	offerDetails := domain.PromotionOfferDetails{
		OfferName:              reqBody.OfferName,
		Description:            reqBody.Description,
		DiscountRate:           reqBody.DiscountRate,
		StartDate:              reqBody.StartDate,
		EndDate:                reqBody.EndDate,
		MinimumPurchaseAmount:  reqBody.MinimumPurchaseAmount,
		TierQuantity:           reqBody.TierQuantity,
		BogoGetQuantity:        reqBody.BogoGetQuantity,
		BogoBuyQuantity:        reqBody.BogoBuyQuantity,
		BogoCombinationEnabled: reqBody.BogoCombinationEnabled,
		GiftDescription:        reqBody.GiftDescription,
	}

	promotion, err := h.promotionUseCase.CreatePromotion(ctx, uint(categoryID), uint(typeID), offerDetails, uint(shopID), reqBody.IsActive)
	if err != nil {
		fmt.Printf("Error creating promotion: %v\n", err)
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create promotion", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Promotion created successfully", promotion)
}

// GetAllPromotions godoc
//
//	@Summary		Get all promotions
//	@Security		BearerAuth
//	@ID				GetAllPromotions
//	@Tags			Admin Promotions
//	@Produce		json
//	@Param			pagination	query	request.Pagination	false	"Pagination"
//	@Router			/admin/promotions/ [get]
//	@Success		200	{object}	response.Response{}	"Promotions retrieved"
//	@Failure		500	{object}	response.Response{}	"Internal server error"
func (h *PromotionHandler) GetAllPromotions(ctx *gin.Context) {
	pagination := request.GetPagination(ctx)

	promotions, err := h.promotionUseCase.GetAllPromotions(ctx, pagination)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve promotions", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Promotions retrieved successfully", promotions)
}

// GetPromotionByID godoc
//
//	@Summary		Get a promotion by ID
//	@Security		BearerAuth
//	@ID				GetPromotionByID
//	@Tags			Admin Promotions
//	@Produce		json
//	@Param			promotion_id	path	int	true	"Promotion ID"
//	@Router			/admin/promotions/{promotion_id} [get]
//	@Success		200	{object}	response.Response{}	"Promotion retrieved"
//	@Failure		400	{object}	response.Response{}	"Invalid promotion ID"
//	@Failure		500	{object}	response.Response{}	"Internal server error"
func (h *PromotionHandler) GetPromotionByID(ctx *gin.Context) {
	promotionIDStr := ctx.Param("promotion_id")
	promotionID, err := strconv.ParseUint(promotionIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid promotion ID", err, nil)
		return
	}

	promotion, err := h.promotionUseCase.GetPromotionByID(ctx, uint(promotionID))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve promotion", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Promotion retrieved successfully", promotion)
}

// DeletePromotion godoc
//
//	@Summary		Delete a promotion by ID
//	@Security		BearerAuth
//	@ID				DeletePromotion
//	@Tags			Admin Promotions
//	@Produce		json
//	@Param			promotion_id	path	int	true	"Promotion ID"
//	@Router			/admin/promotions/{promotion_id} [delete]
//	@Success		200	{object}	response.Response{}	"Promotion deleted"
//	@Failure		400	{object}	response.Response{}	"Invalid promotion ID"
//	@Failure		500	{object}	response.Response{}	"Internal server error"
func (h *PromotionHandler) DeletePromotion(ctx *gin.Context) {
	promotionIDStr := ctx.Param("promotion_id")
	promotionID, err := strconv.ParseUint(promotionIDStr, 10, 32)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid promotion ID", err, nil)
		return
	}

	err = h.promotionUseCase.DeletePromotion(ctx, uint(promotionID))
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to delete promotion", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Promotion deleted successfully", nil)
}
