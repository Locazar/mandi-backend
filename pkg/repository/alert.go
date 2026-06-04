package repository

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repointerfacespkg "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
	"gorm.io/gorm"
)

// AlertRepositoryImpl implements AlertRepository interface
type AlertRepositoryImpl struct {
	db *gorm.DB
}

// NewAlertRepository creates a new alert repository
func NewAlertRepository(db *gorm.DB) repointerfacespkg.AlertRepository {
	return &AlertRepositoryImpl{
		db: db,
	}
}

// SaveAlert saves a new alert
func (r *AlertRepositoryImpl) SaveAlert(ctx context.Context, alert *domain.Alert) error {
	return r.db.WithContext(ctx).Create(alert).Error
}

// GetAlertByID retrieves an alert by ID
func (r *AlertRepositoryImpl) GetAlertByID(ctx context.Context, alertID string) (*domain.Alert, error) {
	var alert domain.Alert
	err := r.db.WithContext(ctx).
		Preload("Actions").
		First(&alert, alertID).Error
	if err != nil {
		return nil, err
	}
	return &alert, nil
}

// GetAlertsBySellerID retrieves alerts for a seller with pagination
func (r *AlertRepositoryImpl) GetAlertsBySellerID(ctx context.Context, sellerID string, limit int, offset int) ([]*domain.Alert, int64, error) {
	var alerts []*domain.Alert
	var total int64

	err := r.db.WithContext(ctx).
		Model(&domain.Alert{}).
		Where("seller_id = ?", sellerID).
		Count(&total).
		Limit(limit).
		Offset(offset).
		Preload("Actions").
		Order("priority DESC, created_at DESC").
		Find(&alerts).Error

	if err != nil {
		return nil, 0, err
	}
	return alerts, total, nil
}

// UpdateAlert updates an existing alert
func (r *AlertRepositoryImpl) UpdateAlert(ctx context.Context, alert *domain.Alert) error {
	return r.db.WithContext(ctx).Save(alert).Error
}

// DeleteAlert deletes an alert
func (r *AlertRepositoryImpl) DeleteAlert(ctx context.Context, alertID string) error {
	return r.db.WithContext(ctx).Delete(&domain.Alert{}, alertID).Error
}

// GetActiveAlertTemplates retrieves all active alert templates
func (r *AlertRepositoryImpl) GetActiveAlertTemplates(ctx context.Context) ([]*domain.AlertTemplate, error) {
	var templates []*domain.AlertTemplate
	err := r.db.WithContext(ctx).
		Where("is_active = ? AND enabled = ?", true, true).
		Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// GetAlertTemplateByKey retrieves a template by key
func (r *AlertRepositoryImpl) GetAlertTemplateByKey(ctx context.Context, key string) (*domain.AlertTemplate, error) {
	var template domain.AlertTemplate
	err := r.db.WithContext(ctx).
		Where("key = ?", key).
		First(&template).Error
	if err != nil {
		return nil, err
	}
	return &template, nil
}

// SaveAlertTemplate saves a new alert template
func (r *AlertRepositoryImpl) SaveAlertTemplate(ctx context.Context, template *domain.AlertTemplate) error {
	return r.db.WithContext(ctx).Create(template).Error
}

// UpdateAlertTemplate updates an alert template
func (r *AlertRepositoryImpl) UpdateAlertTemplate(ctx context.Context, template *domain.AlertTemplate) error {
	return r.db.WithContext(ctx).Save(template).Error
}

// GetAllAlertTemplates retrieves all alert templates regardless of active/enabled status (admin view)
func (r *AlertRepositoryImpl) GetAllAlertTemplates(ctx context.Context) ([]*domain.AlertTemplate, error) {
	var templates []*domain.AlertTemplate
	err := r.db.WithContext(ctx).
		Order("priority DESC, created_at DESC").
		Find(&templates).Error
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// DeleteAlertTemplate hard-deletes a template by its key
func (r *AlertRepositoryImpl) DeleteAlertTemplate(ctx context.Context, key string) error {
	result := r.db.WithContext(ctx).
		Unscoped().
		Where("key = ?", key).
		Delete(&domain.AlertTemplate{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

// LogAlertAction logs an alert action (shown, dismissed, clicked)
func (r *AlertRepositoryImpl) LogAlertAction(ctx context.Context, log *domain.SellerAlertLog) error {
	return r.db.WithContext(ctx).Create(log).Error
}

// GetLastAlertActionTime gets the last time an alert action was recorded
func (r *AlertRepositoryImpl) GetLastAlertActionTime(ctx context.Context, sellerID string, alertKey string) (*time.Time, error) {
	var log domain.SellerAlertLog
	err := r.db.WithContext(ctx).
		Where("seller_id = ? AND alert_key = ? AND action = ?", sellerID, alertKey, "shown").
		Order("created_at DESC").
		First(&log).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // No record found
		}
		return nil, err
	}
	return &log.CreatedAt, nil
}

// GetLastAlertActionTimes fetches the most-recent "shown" timestamp for each of the given alert keys
// in a single query, returning a map of alertKey → last shown time (nil if never shown).
func (r *AlertRepositoryImpl) GetLastAlertActionTimes(ctx context.Context, sellerID string, alertKeys []string) (map[string]*time.Time, error) {
	result := make(map[string]*time.Time, len(alertKeys))
	if len(alertKeys) == 0 {
		return result, nil
	}

	type row struct {
		AlertKey  string    `gorm:"column:alert_key"`
		CreatedAt time.Time `gorm:"column:created_at"`
	}
	var rows []row
	err := r.db.WithContext(ctx).
		Model(&domain.SellerAlertLog{}).
		Select("DISTINCT ON (alert_key) alert_key, created_at").
		Where("seller_id = ? AND alert_key IN ? AND action = ?", sellerID, alertKeys, "shown").
		Order("alert_key, created_at DESC").
		Scan(&rows).Error
	if err != nil {
		return nil, err
	}

	for _, r := range rows {
		t := r.CreatedAt
		result[r.AlertKey] = &t
	}
	return result, nil
}

// GetStepCompletions returns the step numbers completed by a seller for a given onboarding flow key
func (r *AlertRepositoryImpl) GetStepCompletions(ctx context.Context, sellerID string, flowKey string) ([]int, error) {
	var logs []domain.SellerAlertLog
	// alert_key stores the flow key for step completion logs
	err := r.db.WithContext(ctx).
		Where("seller_id = ? AND alert_key = ? AND action = ?", sellerID, flowKey, "step_completed").
		Find(&logs).Error
	if err != nil {
		return nil, err
	}

	var rawSteps []int
	for _, log := range logs {
		if log.Metadata == nil {
			continue
		}
		var meta struct {
			Step int `json:"step"`
		}
		if err := json.Unmarshal(log.Metadata, &meta); err != nil {
			continue
		}
		// Steps are 1-indexed; entries with step=0 are ignored
		if meta.Step > 0 {
			rawSteps = append(rawSteps, meta.Step)
		}
	}

	seen := make(map[int]bool)
	steps := make([]int, 0)
	for _, s := range rawSteps {
		if !seen[s] {
			seen[s] = true
			steps = append(steps, s)
		}
	}
	return steps, nil
}

// GetAggregatedDataForSeller fetches all required data for rule evaluation in minimal queries
func (r *AlertRepositoryImpl) GetAggregatedDataForSeller(ctx context.Context, adminID string) (*domain.AggregatedData, error) {
	// Fetch shop details
	var shop domain.ShopDetails
	err := r.db.WithContext(ctx).
		Where("admin_id = ?", adminID).
		First(&shop).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, usecase.ErrShopNotFound
		}
		return nil, err
	}

	// Fetch product count for this shop
	var productCount int64
	err = r.db.WithContext(ctx).
		Model(&domain.ProductItem{}).
		Where("shop_id = ?", shop.ID).
		Count(&productCount).Error
	if err != nil {
		return nil, err
	}

	// Determine verification status
	isVerified := shop.ShopVerificationStatus

	// Check if shop has a photo
	hasShopPhoto := shop.Shop_Image_URL != ""

	aggregatedData := &domain.AggregatedData{
		Shop:         shop,
		ProductCount: int(productCount),
		IsVerified:   isVerified,
		HasShopPhoto: hasShopPhoto,
		AdminID:      adminID,
		ShopID:       shop.ID,
		FetchedAt:    time.Now(),
	}

	return aggregatedData, nil
}
