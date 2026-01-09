package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
)

type PromotionUseCase interface {
	FindAllPromotionCategories(ctx context.Context, pagination request.Pagination) ([]response.PromotionCategory, error)
	FindPromotionCategoryByID(ctx context.Context, categoryID uint) (response.PromotionCategory, error)

	FindAllPromotionTypes(ctx context.Context, pagination request.Pagination) ([]response.PromotionsType, error)
	FindPromotionTypesByCategoryID(ctx context.Context, categoryID uint, pagination request.Pagination) ([]response.PromotionsType, error)
	FindPromotionTypeByID(ctx context.Context, typeID uint) (response.PromotionsType, error)

	CreatePromotion(ctx context.Context, promotionCategoryID, promotionTypeID uint, offerDetails string, shopID string) (response.Promotion, error)
	GetAllPromotions(ctx context.Context, pagination request.Pagination) ([]response.Promotion, error)
	GetPromotionByID(ctx context.Context, promotionID uint) (response.Promotion, error)
	DeletePromotion(ctx context.Context, promotionID uint) error
}
