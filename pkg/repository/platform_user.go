package repository

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type platformUserRepository struct {
	db *gorm.DB
}

func NewPlatformUserRepository(db *gorm.DB) interfaces.PlatformUserRepository {
	return &platformUserRepository{db: db}
}

func (r *platformUserRepository) ListAdmins(ctx context.Context, pagination request.Pagination) ([]domain.Admin, error) {
	var admins []domain.Admin
	err := r.db.WithContext(ctx).
		Offset(int(pagination.Offset)).
		Limit(int(pagination.Limit)).
		Find(&admins).Error
	return admins, err
}

func (r *platformUserRepository) GetAdminByID(ctx context.Context, adminID string) (domain.Admin, error) {
	var admin domain.Admin
	err := r.db.WithContext(ctx).Where("id = ?", adminID).First(&admin).Error
	return admin, err
}

func (r *platformUserRepository) UpdateAdminRole(ctx context.Context, adminID string, role domain.AdminRole) error {
	return r.db.WithContext(ctx).
		Model(&domain.Admin{}).
		Where("id = ?", adminID).
		Update("role", role).Error
}

func (r *platformUserRepository) DeactivateAdmin(ctx context.Context, adminID string) error {
	return r.db.WithContext(ctx).
		Model(&domain.Admin{}).
		Where("id = ?", adminID).
		Update("status", domain.AdminStatusSuspended).Error
}
