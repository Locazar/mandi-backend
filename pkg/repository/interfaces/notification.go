package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type NotificationRepository interface {
	SaveNotification(ctx context.Context, notification domain.Notification) error
	GetNotifications(ctx context.Context, filter request.GetNotification, pagination request.Pagination) ([]domain.Notification, error)
	MarkNotificationAsRead(ctx context.Context, notificationID uint) error

	// FCM token management in Postgres
	SaveDeviceToken(ctx context.Context, token domain.NotificationDeviceToken) error
	GetActiveTokensByOwner(ctx context.Context, ownerID, ownerType string) ([]string, error)
	DeleteDeviceToken(ctx context.Context, ownerID, ownerType, token string) error
}
