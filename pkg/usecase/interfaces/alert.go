package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// AlertUseCase defines the alert business logic interface
type AlertUseCase interface {
	// GetSellerAlerts fetches all alerts for a seller
	GetSellerAlerts(ctx context.Context, sellerID string) ([]*domain.Alert, error)
	// DismissAlert marks an alert as dismissed
	DismissAlert(ctx context.Context, sellerID string, alertKey string) error
	// LogAlertView logs when an alert is viewed by seller
	LogAlertView(ctx context.Context, sellerID string, alertKey string) error
}
