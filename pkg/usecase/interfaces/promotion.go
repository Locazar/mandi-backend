package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type PromotionUseCase interface {
	FindAllPromotionCategories(ctx context.Context, pagination request.Pagination) ([]response.PromotionCategory, error)
	FindPromotionCategoryByID(ctx context.Context, categoryID string) (response.PromotionCategory, error)

	FindAllPromotionTypes(ctx context.Context, pagination request.Pagination) ([]response.PromotionsType, error)
	FindPromotionTypesByCategoryID(ctx context.Context, categoryID string, pagination request.Pagination) ([]response.PromotionsType, error)
	FindPromotionTypeByID(ctx context.Context, typeID string) (response.PromotionsType, error)

	CreatePromotion(ctx context.Context, promotionCategoryID, promotionTypeID string, offerDetails domain.PromotionOfferDetails, shopID string, isActive bool) (response.Promotion, error)
	GetAllPromotions(ctx context.Context, pagination request.Pagination) ([]response.Promotion, error)
	GetPromotionByID(ctx context.Context, promotionID string) (response.Promotion, error)
	DeletePromotion(ctx context.Context, promotionID string) error
}
