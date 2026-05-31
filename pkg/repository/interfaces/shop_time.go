package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type ShopTimeRepository interface {
	CreateShopTime(ctx context.Context, shopTime domain.ShopTime) error
	GetShopTimeByShopID(ctx context.Context, shopID string) (domain.ShopTime, error)
	UpdateShopTime(ctx context.Context, shopTime domain.ShopTime) error
}
