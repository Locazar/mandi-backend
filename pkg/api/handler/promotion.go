package handler

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
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

func (h *PromotionHandler) CreatePromotion(ctx *gin.Context) {
	var reqBody struct {
		PromotionCategoryID string `json:"promotion_category_id" binding:"required"`
		PromotionTypeID     string `json:"promotion_type_id" binding:"required"`
		OfferDetails        string `json:"offer_details" binding:"required"`
		ShopID              string `json:"shop_id" binding:"required"`
	}

	if err := ctx.ShouldBindJSON(&reqBody); err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid request body", err, nil)
		return
	}

	// Parse string IDs to uint
	categoryID, err := strconv.ParseUint(reqBody.PromotionCategoryID, 10, 32)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid promotion category ID", err, nil)
		return
	}

	typeID, err := strconv.ParseUint(reqBody.PromotionTypeID, 10, 32)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusBadRequest, "Invalid promotion type ID", err, nil)
		return
	}

	promotion, err := h.promotionUseCase.CreatePromotion(ctx, uint(categoryID), uint(typeID), reqBody.OfferDetails, reqBody.ShopID)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to create promotion", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusCreated, "Promotion created successfully", promotion)
}

func (h *PromotionHandler) GetAllPromotions(ctx *gin.Context) {
	pagination := request.GetPagination(ctx)

	promotions, err := h.promotionUseCase.GetAllPromotions(ctx, pagination)
	if err != nil {
		response.ErrorResponse(ctx, http.StatusInternalServerError, "Failed to retrieve promotions", err, nil)
		return
	}

	response.SuccessResponse(ctx, http.StatusOK, "Promotions retrieved successfully", promotions)
}

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
