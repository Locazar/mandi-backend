package repository

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type bannerRepository struct {
	db *gorm.DB
}

func NewBannerRepository(db *gorm.DB) repo.BannerRepository {
	return &bannerRepository{
		db: db,
	}
}

func (r *bannerRepository) GetActiveBanners(ctx context.Context) ([]domain.Banner, error) {
	var banners []domain.Banner
	err := r.db.WithContext(ctx).Where("active = ?", true).Find(&banners).Error
	return banners, err
}
