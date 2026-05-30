package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type ShopTimeUseCase interface {
	SetShopTime(ctx context.Context, shopID string, shopTime domain.ShopTime) error
	GetShopTime(ctx context.Context, shopID string) (domain.ShopTime, error)
}
