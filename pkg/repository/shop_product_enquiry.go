package repository

import (
	"context"
	"fmt"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type shopProductEnquiryRepository struct {
	db *gorm.DB
}

func NewShopProductEnquiryRepository(db *gorm.DB) interfaces.ShopProductEnquiryRepository {
	return &shopProductEnquiryRepository{db: db}
}

func (r *shopProductEnquiryRepository) Create(ctx context.Context, e domain.ShopProductEnquiry) error {
	if err := r.db.WithContext(ctx).Create(&e).Error; err != nil {
		return fmt.Errorf("failed to create shop product enquiry: %w", err)
	}
	return nil
}

func (r *shopProductEnquiryRepository) GetByID(ctx context.Context, id string) (domain.ShopProductEnquiry, error) {
	var e domain.ShopProductEnquiry
	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&e).Error; err != nil {
		return e, fmt.Errorf("shop product enquiry not found: %w", err)
	}
	return e, nil
}

func (r *shopProductEnquiryRepository) List(ctx context.Context, shopID string, actioned *bool, enquiryType string) ([]domain.ShopProductEnquiry, error) {
	var results []domain.ShopProductEnquiry
	q := r.db.WithContext(ctx).Order("created_at DESC")
	if shopID != "" {
		q = q.Where("shop_id = ?", shopID)
	}
	if actioned != nil {
		q = q.Where("is_actioned = ?", *actioned)
	}
	if enquiryType != "" {
		q = q.Where("enquiry_type = ?", enquiryType)
	}
	if err := q.Find(&results).Error; err != nil {
		return nil, fmt.Errorf("failed to list shop product enquiries: %w", err)
	}
	return results, nil
}

func (r *shopProductEnquiryRepository) MarkActioned(ctx context.Context, id string, note string) error {
	result := r.db.WithContext(ctx).Model(&domain.ShopProductEnquiry{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{
			"is_actioned": true,
			"admin_note":  note,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to mark shop product enquiry as actioned: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("shop product enquiry not found: %s", id)
	}
	return nil
}
