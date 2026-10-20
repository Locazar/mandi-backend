package interfaces

import (
	"context"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// AlertHandler defines the interface for alert handlers
type AlertHandler interface {
	GetSellerAlerts(ctx *gin.Context)
	DismissAlert(ctx *gin.Context)
	// MarkAlertShown records that an alert was displayed, so its configured
	// frequency (once/daily/weekly) is honoured on the next fetch.
	MarkAlertShown(ctx *gin.Context)
}

// AlertUseCase defines the alert business logic interface
type AlertUseCase interface {
	GetSellerAlerts(ctx context.Context, sellerID string) ([]*domain.Alert, error)
	DismissAlert(ctx context.Context, sellerID string, alertKey string) error
	LogAlertView(ctx context.Context, sellerID string, alertKey string) error
}
