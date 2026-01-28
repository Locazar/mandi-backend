package usecase

import (
	"context"
	"errors"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"gorm.io/gorm"
)

type shopTimeUseCase struct {
	shopTimeRepo repo.ShopTimeRepository
}

func NewShopTimeUseCase(shopTimeRepo repo.ShopTimeRepository) interfaces.ShopTimeUseCase {
	return &shopTimeUseCase{
		shopTimeRepo: shopTimeRepo,
	}
}

func (u *shopTimeUseCase) SetShopTime(ctx context.Context, shopID uint, shopTime domain.ShopTime) error {
	shopTime.ShopID = shopID
	existing, err := u.shopTimeRepo.GetShopTimeByShopID(ctx, shopID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}
	if existing.ID == 0 {
		return u.shopTimeRepo.CreateShopTime(ctx, shopTime)
	}
	shopTime.ID = existing.ID
	return u.shopTimeRepo.UpdateShopTime(ctx, shopTime)
}

func (u *shopTimeUseCase) GetShopTime(ctx context.Context, shopID uint) (domain.ShopTime, error) {
	return u.shopTimeRepo.GetShopTimeByShopID(ctx, shopID)
}
