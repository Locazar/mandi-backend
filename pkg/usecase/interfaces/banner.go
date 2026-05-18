package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// BannerUseCase defines the business logic contract for banner operations exposed to the user API.
type BannerUseCase interface {
	GetFilteredBanners(ctx context.Context, req request.BannerFilterRequest) ([]domain.Banner, error)
}
