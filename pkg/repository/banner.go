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

func (r *bannerRepository) GetActiveBanners(ctx context.Context, departmentID, categoryID *string) ([]domain.Banner, error) {
	var banners []domain.Banner
	query := r.db.WithContext(ctx).Where("active = ?", true)

	if departmentID != nil && *departmentID != "" {
		query = query.Where("department_id = ?", *departmentID)
	}
	if categoryID != nil && *categoryID != "" {
		query = query.Where("category_id = ?", *categoryID)
	}

	err := query.Find(&banners).Error
	return banners, err
}
