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
		Select("id, user_name, full_name, email, mobile, profile_image_url, status, role, created_at, updated_at").
		Where("role <> ?", domain.AdminRoleSeller).
		Order("created_at DESC").
		Offset(int(pagination.Offset)).
		Limit(int(pagination.Limit)).
		Find(&admins).Error
	return admins, err
}

func (r *platformUserRepository) UpdateAdminRole(ctx context.Context, adminID string, role domain.AdminRole) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Admin{}).
		Where("id = ?", adminID).
		Update("role", role)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *platformUserRepository) UpdateAdminDetails(ctx context.Context, adminID string, updates map[string]interface{}) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Admin{}).
		Where("id = ?", adminID).
		Updates(updates)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *platformUserRepository) DeleteAdmin(ctx context.Context, adminID string) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", adminID).
		Delete(&domain.Admin{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (r *platformUserRepository) DeactivateAdmin(ctx context.Context, adminID string) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Admin{}).
		Where("id = ?", adminID).
		Update("status", domain.AdminStatusSuspended)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}
