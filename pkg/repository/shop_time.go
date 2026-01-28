package repository

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type shopTimeRepository struct {
	db *gorm.DB
}

func NewShopTimeRepository(db *gorm.DB) repo.ShopTimeRepository {
	return &shopTimeRepository{
		db: db,
	}
}

func (r *shopTimeRepository) CreateShopTime(ctx context.Context, shopTime domain.ShopTime) error {
	return r.db.WithContext(ctx).Create(&shopTime).Error
}

func (r *shopTimeRepository) GetShopTimeByShopID(ctx context.Context, shopID uint) (domain.ShopTime, error) {
	var shopTime domain.ShopTime
	err := r.db.WithContext(ctx).Where("shop_id = ?", shopID).First(&shopTime).Error
	return shopTime, err
}

func (r *shopTimeRepository) UpdateShopTime(ctx context.Context, shopTime domain.ShopTime) error {
	return r.db.WithContext(ctx).Save(&shopTime).Error
}