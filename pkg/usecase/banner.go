package usecase

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type bannerUseCase struct {
	bannerRepo repo.BannerRepository
}

// NewBannerUseCase creates a new BannerUseCase instance.
func NewBannerUseCase(bannerRepo repo.BannerRepository) interfaces.BannerUseCase {
	return &bannerUseCase{
		bannerRepo: bannerRepo,
	}
}

// GetFilteredBanners delegates to the repository which applies all filters
// (geo-location via Haversine in Go, pincode, department_id, category_id, app_type).
func (u *bannerUseCase) GetFilteredBanners(ctx context.Context, req request.BannerFilterRequest) ([]domain.Banner, error) {
	return u.bannerRepo.GetBannersForUser(ctx, req)
}
