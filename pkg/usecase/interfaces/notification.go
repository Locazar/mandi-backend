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
	MarkNotificationAsRead(ctx context.Context, notificationID string) error

	// Device token lifecycle
	RegisterDeviceToken(ctx context.Context, req request.NotificationDeviceToken) error
	UnregisterDeviceToken(ctx context.Context, req request.UnregisterDeviceToken) error

	// FCM push delivery. delivered reports whether a device was actually
	// reached (true) versus a successful no-op because the owner has no
	// registered device right now (false, err == nil) — callers that send to
	// many owners in a loop (e.g. "send to all sellers") need this to report
	// real per-recipient status instead of a blanket "sent".
	SendPushNotification(ctx context.Context, req request.SendPushRequest) (delivered bool, err error)

	// SendBroadcast delivers a notification to a whole audience via an FCM topic.
	SendBroadcast(ctx context.Context, req request.SendBroadcastRequest) error

	// Radius-targeted announcements. SendRadiusAnnouncement centres on raw
	// coordinates; the shop-launch pair centres on a shop's pinned location.
	SendRadiusAnnouncement(ctx context.Context, adminID string, req request.RadiusAnnouncement) (domain.RadiusAnnouncementResult, error)
	PreviewShopLaunchAudience(ctx context.Context, shopID string, radiusKm float64) (domain.RadiusAnnouncementResult, error)
	SendShopLaunchAnnouncement(ctx context.Context, adminID string, req request.ShopLaunchAnnouncement) (domain.RadiusAnnouncementResult, error)

	// StartFirestoreWatcher launches background Firestore listeners for the
	// given rules.  It returns immediately; watchers run until ctx is cancelled.
	// Pass nil to use the default e-commerce rules (orders, products, shops,
	// enquiries).
	StartFirestoreWatcher(ctx context.Context, rules []notificationSvc.WatchRule) error
}
