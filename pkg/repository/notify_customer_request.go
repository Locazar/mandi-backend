package repository

import (
	"context"
	"fmt"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type notifyCustomerRequestRepository struct {
	db *gorm.DB
}

func NewNotifyCustomerRequestRepository(db *gorm.DB) interfaces.NotifyCustomerRequestRepository {
	return &notifyCustomerRequestRepository{db: db}
}

func (r *notifyCustomerRequestRepository) Create(ctx context.Context, req domain.NotifyCustomerRequest) error {
	if err := r.db.WithContext(ctx).Create(&req).Error; err != nil {
		return fmt.Errorf("failed to create notify customer request: %w", err)
	}
	return nil
}

func (r *notifyCustomerRequestRepository) ExistsByEnquiryID(ctx context.Context, enquiryID string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&domain.NotifyCustomerRequest{}).
		Where("enquiry_id = ?", enquiryID).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check notify customer request: %w", err)
	}
	return count > 0, nil
}

func (r *notifyCustomerRequestRepository) List(ctx context.Context, actioned *bool) ([]domain.NotifyCustomerRequest, error) {
	var requests []domain.NotifyCustomerRequest
	q := r.db.WithContext(ctx).Order("created_at DESC")
	if actioned != nil {
		q = q.Where("is_actioned = ?", *actioned)
	}
	if err := q.Find(&requests).Error; err != nil {
		return nil, fmt.Errorf("failed to list notify customer requests: %w", err)
	}
	return requests, nil
}

func (r *notifyCustomerRequestRepository) MarkActioned(ctx context.Context, id string, note string) error {
	result := r.db.WithContext(ctx).Model(&domain.NotifyCustomerRequest{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_actioned": true,
			"admin_note":  note,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to mark notify customer request as actioned: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("notify customer request not found: %s", id)
	}
	return nil
}
