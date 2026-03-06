package repository

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"gorm.io/gorm"
)

type ShopSocialRepository struct {
	DB *gorm.DB
}

func NewShopSocialRepository(db *gorm.DB) *ShopSocialRepository {
	return &ShopSocialRepository{DB: db}
}

// GetShopSocialDetails returns all social details for a shop
func (r *ShopSocialRepository) GetShopSocialDetails(ctx context.Context, shopID uint) ([]domain.ShopSocial, error) {
	var details []domain.ShopSocial
	if err := r.DB.WithContext(ctx).Where("shop_id = ?", shopID).Find(&details).Error; err != nil {
		return nil, err
	}
	return details, nil
}
