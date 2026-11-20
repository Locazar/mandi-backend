package usecase

import (
	"context"
	"fmt"
	"strings"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"gorm.io/gorm"
)

type ShopActionLogService struct {
	db *gorm.DB
}

func NewShopActionLogService(db *gorm.DB) *ShopActionLogService {
	return &ShopActionLogService{db: db}
}

type ShopActionLogFilter struct {
	Action   string
	ShopID   string
	SellerID string
	DateFrom string
	DateTo   string
	Limit    int
	Offset   int
}

// Record inserts one log entry. Resolves SellerID from shop_details when absent.
func (s *ShopActionLogService) Record(ctx context.Context, entry domain.ShopActionLog) error {
	if entry.SellerID == "" && entry.ShopID != "" {
		var shop domain.ShopDetails
		s.db.WithContext(ctx).Select("admin_id").Where("id = ?", entry.ShopID).First(&shop)
		entry.SellerID = shop.AdminID
	}
	if err := s.db.WithContext(ctx).Create(&entry).Error; err != nil {
		return fmt.Errorf("shop_action_log: insert: %w", err)
	}
	return nil
}

// List returns logs matching the filter with total count for pagination.
func (s *ShopActionLogService) List(ctx context.Context, f ShopActionLogFilter) ([]domain.ShopActionLog, int64, error) {
	q := s.db.WithContext(ctx).Model(&domain.ShopActionLog{}).Order("created_at DESC")

	if v := strings.TrimSpace(f.Action); v != "" {
		q = q.Where("action = ?", v)
	}
	if v := strings.TrimSpace(f.ShopID); v != "" {
		q = q.Where("shop_id = ?", v)
	}
	if v := strings.TrimSpace(f.SellerID); v != "" {
		q = q.Where("seller_id = ?", v)
	}
	if v := strings.TrimSpace(f.DateFrom); v != "" {
		q = q.Where("created_at >= ?", v)
	}
	if v := strings.TrimSpace(f.DateTo); v != "" {
		q = q.Where("created_at < ?", v)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, fmt.Errorf("shop_action_log: count: %w", err)
	}

	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	offset := f.Offset
	if offset < 0 {
		offset = 0
	}

	var logs []domain.ShopActionLog
	if err := q.Limit(limit).Offset(offset).Find(&logs).Error; err != nil {
		return nil, 0, fmt.Errorf("shop_action_log: list: %w", err)
	}
	return logs, total, nil
}
