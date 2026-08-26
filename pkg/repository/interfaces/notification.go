package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type NotificationRepository interface {
	SaveNotification(ctx context.Context, notification domain.Notification) error
	GetNotifications(ctx context.Context, filter request.GetNotification, pagination request.Pagination) ([]domain.Notification, error)
	MarkNotificationAsRead(ctx context.Context, notificationID string) error

	// FCM token management in Postgres
	SaveDeviceToken(ctx context.Context, token domain.NotificationDeviceToken) error
	GetActiveTokensByOwner(ctx context.Context, ownerID, ownerType string) ([]string, error)
	DeleteDeviceToken(ctx context.Context, ownerID, ownerType, token string) error

	// Shop launch announcements — radius-targeted "this shop just opened" pushes.
	GetShopAnnouncementTarget(ctx context.Context, shopID string) (domain.ShopAnnouncementTarget, error)
	GetCustomerDevicesInRadius(ctx context.Context, lat, lng, radiusKm float64) ([]domain.CustomerDevice, error)
	SaveNotificationsBatch(ctx context.Context, notifications []domain.Notification) error
}
