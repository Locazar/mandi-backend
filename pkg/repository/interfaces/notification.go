package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type NotificationRepository interface {
	SaveNotification(ctx context.Context, notification domain.Notification) error
	GetNotification(ctx context.Context, filter request.Notification) error
	MarkNotificationAsRead(ctx context.Context, notificationID uint) error
	GenerateFCMToken(ctx context.Context, request request.NotificationDeviceToken) error
}
