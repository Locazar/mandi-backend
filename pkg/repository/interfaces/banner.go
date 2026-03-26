package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type BannerRepository interface {
	GetActiveBanners(ctx context.Context) ([]domain.Banner, error)
}
