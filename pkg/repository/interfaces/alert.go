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
	GetAlertByID(ctx context.Context, alertID string) (*domain.Alert, error)
	GetAlertsBySellerID(ctx context.Context, sellerID string, limit int, offset int) ([]*domain.Alert, int64, error)
	UpdateAlert(ctx context.Context, alert *domain.Alert) error
	DeleteAlert(ctx context.Context, alertID string) error

	// AlertTemplate operations
	GetActiveAlertTemplates(ctx context.Context) ([]*domain.AlertTemplate, error)
	GetAllAlertTemplates(ctx context.Context) ([]*domain.AlertTemplate, error)
	GetAlertTemplateByKey(ctx context.Context, key string) (*domain.AlertTemplate, error)
	SaveAlertTemplate(ctx context.Context, template *domain.AlertTemplate) error
	UpdateAlertTemplate(ctx context.Context, template *domain.AlertTemplate) error
	DeleteAlertTemplate(ctx context.Context, key string) error

	// SellerAlertLog operations
	LogAlertAction(ctx context.Context, log *domain.SellerAlertLog) error
	GetLastAlertActionTime(ctx context.Context, sellerID string, alertKey string) (*time.Time, error)
	GetLastAlertActionTimes(ctx context.Context, sellerID string, alertKeys []string) (map[string]*time.Time, error)
	GetStepCompletions(ctx context.Context, sellerID string, flowKey string) ([]int, error)

	// Aggregated data operations
	GetAggregatedDataForSeller(ctx context.Context, adminID string) (*domain.AggregatedData, error)
}
