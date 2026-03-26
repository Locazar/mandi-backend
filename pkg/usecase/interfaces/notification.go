package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	notificationSvc "github.com/rohit221990/mandi-backend/pkg/service/notification"
)

type NotificationUseCase interface {
	// Persistence
	SaveNotification(ctx context.Context, notification request.Notification) error
	GetNotificationsBy(ctx context.Context, filter request.GetNotification, pagination request.Pagination) ([]domain.Notification, error)
	MarkNotificationAsRead(ctx context.Context, notificationID uint) error

	// Device token lifecycle
	RegisterDeviceToken(ctx context.Context, req request.NotificationDeviceToken) error
	UnregisterDeviceToken(ctx context.Context, req request.UnregisterDeviceToken) error

	// FCM push delivery
	SendPushNotification(ctx context.Context, req request.SendPushRequest) error

	// StartFirestoreWatcher launches background Firestore listeners for the
	// given rules.  It returns immediately; watchers run until ctx is cancelled.
	// Pass nil to use the default e-commerce rules (orders, products, shops,
	// enquiries).
	StartFirestoreWatcher(ctx context.Context, rules []notificationSvc.WatchRule) error
}
