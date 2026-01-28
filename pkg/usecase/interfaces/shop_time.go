package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type ShopTimeUseCase interface {
	SetShopTime(ctx context.Context, shopID uint, shopTime domain.ShopTime) error
	GetShopTime(ctx context.Context, shopID uint) (domain.ShopTime, error)
}