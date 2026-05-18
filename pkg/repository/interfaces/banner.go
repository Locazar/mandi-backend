package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type BannerRepository interface {
	GetActiveBanners(ctx context.Context, departmentID, categoryID *string) ([]domain.Banner, error)
	GetBannersForUser(ctx context.Context, req request.BannerFilterRequest) ([]domain.Banner, error)
}
