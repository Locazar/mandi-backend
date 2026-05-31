package repository

import (
	"context"
	"fmt"

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

func (r *shopTimeRepository) GetShopTimeByShopID(ctx context.Context, shopID string) (domain.ShopTime, error) {
	var shopTime domain.ShopTime
	result := r.db.WithContext(ctx).Where("shop_id = ?", shopID).Find(&shopTime)
	if result.Error != nil {
		return shopTime, result.Error
	}
	if result.RowsAffected == 0 {
		return shopTime, gorm.ErrRecordNotFound
	}
	return shopTime, nil
}

func (r *shopTimeRepository) UpdateShopTime(ctx context.Context, shopTime domain.ShopTime) error {
	fmt.Printf("Updating shop time for shop ID %s: %+v\n", shopTime.ShopID, shopTime)
	return r.db.WithContext(ctx).Save(&shopTime).Error
}
