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
}

// AlertUseCase defines the alert business logic interface
type AlertUseCase interface {
	GetSellerAlerts(ctx context.Context, sellerID string) ([]*domain.Alert, error)
	DismissAlert(ctx context.Context, sellerID string, alertKey string) error
	LogAlertView(ctx context.Context, sellerID string, alertKey string) error
}
