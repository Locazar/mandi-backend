package interfaces

import (
	"context"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// AlertRepository handles alert-related database operations
type AlertRepository interface {
	// Alert operations
	SaveAlert(ctx context.Context, alert *domain.Alert) error
	GetAlertByID(ctx context.Context, alertID uint) (*domain.Alert, error)
	GetAlertsBySellerID(ctx context.Context, sellerID uint, limit int, offset int) ([]*domain.Alert, int64, error)
	UpdateAlert(ctx context.Context, alert *domain.Alert) error
	DeleteAlert(ctx context.Context, alertID uint) error

	// AlertTemplate operations
	GetActiveAlertTemplates(ctx context.Context) ([]*domain.AlertTemplate, error)
	GetAlertTemplateByKey(ctx context.Context, key string) (*domain.AlertTemplate, error)
	SaveAlertTemplate(ctx context.Context, template *domain.AlertTemplate) error
	UpdateAlertTemplate(ctx context.Context, template *domain.AlertTemplate) error

	// SellerAlertLog operations
	LogAlertAction(ctx context.Context, log *domain.SellerAlertLog) error
	GetLastAlertActionTime(ctx context.Context, sellerID uint, alertKey string) (*time.Time, error)

	// Aggregated data operations
	GetAggregatedDataForSeller(ctx context.Context, adminID uint) (*domain.AggregatedData, error)
}
