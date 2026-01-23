package usecase

import (
	"context"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
)

type promotionUseCase struct {
	promotionRepo interfaces.PromotionRepository
}

func NewPromotionUseCase(promotionRepo interfaces.PromotionRepository) *promotionUseCase {
	return &promotionUseCase{
		promotionRepo: promotionRepo,
	}
}

func (u *promotionUseCase) FindAllPromotionCategories(ctx context.Context, pagination request.Pagination) ([]response.PromotionCategory, error) {
	return u.promotionRepo.FindAllPromotionCategories(ctx, pagination)
}

func (u *promotionUseCase) FindPromotionCategoryByID(ctx context.Context, categoryID uint) (response.PromotionCategory, error) {
	return u.promotionRepo.FindPromotionCategoryByID(ctx, categoryID)
}

func (u *promotionUseCase) FindAllPromotionTypes(ctx context.Context, pagination request.Pagination) ([]response.PromotionsType, error) {
	return u.promotionRepo.FindAllPromotionTypes(ctx, pagination)
}

func (u *promotionUseCase) FindPromotionTypesByCategoryID(ctx context.Context, categoryID uint, pagination request.Pagination) ([]response.PromotionsType, error) {
	return u.promotionRepo.FindPromotionTypesByCategoryID(ctx, categoryID, pagination)
}

func (u *promotionUseCase) FindPromotionTypeByID(ctx context.Context, typeID uint) (response.PromotionsType, error) {
	return u.promotionRepo.FindPromotionTypeByID(ctx, typeID)
}

func (u *promotionUseCase) CreatePromotion(ctx context.Context, promotionCategoryID, promotionTypeID uint, offerDetails domain.PromotionOfferDetails, shopID uint, isActive bool) (response.Promotion, error) {

	promotion := domain.Promotion{
		PromotionCategoryID:    promotionCategoryID,
		PromotionTypeID:        promotionTypeID,
		OfferName:              offerDetails.OfferName,
		Description:            offerDetails.Description,
		DiscountRate:           offerDetails.DiscountRate,
		StartDate:              offerDetails.StartDate,
		EndDate:                offerDetails.EndDate,
		MinimumPurchaseAmount:  offerDetails.MinimumPurchaseAmount,
		TierQuantity:           offerDetails.TierQuantity,
		BogoGetQuantity:        offerDetails.BogoGetQuantity,
		BogoBuyQuantity:        offerDetails.BogoBuyQuantity,
		BogoCombinationEnabled: offerDetails.BogoCombinationEnabled,
		GiftDescription:        offerDetails.GiftDescription,
		ShopID:                 shopID,
		IsActive:               isActive,
		CreatedAt:              time.Now().UTC(),
		UpdatedAt:              time.Now().UTC(),
	}

	return u.promotionRepo.CreatePromotion(ctx, promotion)
}

func (u *promotionUseCase) GetAllPromotions(ctx context.Context, pagination request.Pagination) ([]response.Promotion, error) {
	return u.promotionRepo.GetAllPromotions(ctx, pagination)
}

func (u *promotionUseCase) GetPromotionByID(ctx context.Context, promotionID uint) (response.Promotion, error) {
	return u.promotionRepo.GetPromotionByID(ctx, promotionID)
}

func (u *promotionUseCase) DeletePromotion(ctx context.Context, promotionID uint) error {
	return u.promotionRepo.DeletePromotion(ctx, promotionID)
}
