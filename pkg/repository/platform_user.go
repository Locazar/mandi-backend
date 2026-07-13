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

func (r *platformUserRepository) ListAdmins(ctx context.Context, roleFilter string, pagination request.Pagination) ([]domain.Admin, error) {
	var admins []domain.Admin
	query := r.db.WithContext(ctx).
		Select("id, user_name, full_name, email, mobile, profile_image_url, status, role, referral_coupon_id, created_at, updated_at")

	if roleFilter != "" {
		query = query.Where("role = ?", roleFilter)
	} else {
		// Exclude sellers explicitly, and also NULL/blank role — a plain
		// "role <> 'seller'" never matches NULL or '' in SQL (NULL <> x is
		// NULL, not true), so a legacy admin row with an unset role would
		// silently leak into this "platform users only" listing instead of
		// being excluded like every other seller account.
		query = query.Where("role <> ? AND role IS NOT NULL AND role <> ''", domain.AdminRoleSeller)
	}

	err := query.
		Order("created_at DESC").
		Offset(int(pagination.Offset)).
		Limit(int(pagination.Limit)).
		Find(&admins).Error
	if err != nil {
		return nil, err
	}

	if err := r.attachShopsOnboardedCounts(ctx, admins); err != nil {
		return nil, err
	}
	return admins, nil
}

// attachShopsOnboardedCounts batch-populates ShopsOnboardedCount for every
// admin in the slice that carries a referral coupon — one grouped COUNT
// query instead of one query per row.
func (r *platformUserRepository) attachShopsOnboardedCounts(ctx context.Context, admins []domain.Admin) error {
	ids := make([]string, 0, len(admins))
	for _, a := range admins {
		if a.ReferralCouponID != "" {
			ids = append(ids, a.ID)
		}
	}
	if len(ids) == 0 {
		return nil
	}

	var rows []struct {
		PlatformUserID string
		Count          int64
	}
	if err := r.db.WithContext(ctx).Model(&domain.ShopReferral{}).
		Select("platform_user_id, COUNT(*) as count").
		Where("platform_user_id IN ? AND status = ?", ids, domain.ShopReferralStatusActive).
		Group("platform_user_id").
		Scan(&rows).Error; err != nil {
		return err
	}

	counts := make(map[string]int64, len(rows))
	for _, row := range rows {
		counts[row.PlatformUserID] = row.Count
	}
	for i := range admins {
		admins[i].ShopsOnboardedCount = counts[admins[i].ID]
	}
	return nil
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
