package repository

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repointerfacespkg "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type leadRequestRepository struct {
	db *gorm.DB
}

func NewLeadRequestRepository(db *gorm.DB) repointerfacespkg.LeadRequestRepository {
	return &leadRequestRepository{db: db}
}

func (r *leadRequestRepository) Create(ctx context.Context, lead *domain.LeadRequest) error {
	return r.db.WithContext(ctx).Create(lead).Error
}

func (r *leadRequestRepository) List(ctx context.Context, city, status string, limit, offset int) ([]domain.LeadRequest, int64, error) {
	var leads []domain.LeadRequest
	var total int64

	q := r.db.WithContext(ctx).Model(&domain.LeadRequest{})
	if city != "" {
		q = q.Where("city = ?", city)
	}
	if status != "" {
		q = q.Where("status = ?", status)
	}

	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	if err := q.Order("created_at DESC").Limit(limit).Offset(offset).Find(&leads).Error; err != nil {
		return nil, 0, err
	}
	return leads, total, nil
}

func (r *leadRequestRepository) UpdateStatus(ctx context.Context, id uint, status, adminNote string) error {
	updates := map[string]interface{}{
		"admin_note": adminNote, // always written — empty string clears an existing note
	}
	if status != "" {
		updates["status"] = status
	}
	return r.db.WithContext(ctx).Model(&domain.LeadRequest{}).Where("id = ?", id).Updates(updates).Error
}
