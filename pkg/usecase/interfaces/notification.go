package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type NotificationUseCase interface {
	SaveNotification(ctx context.Context, notification request.Notification) error
	MarkNotificationAsRead(ctx context.Context, notificationID uint) error
	GetNotificationsBy(ctx context.Context, filter request.GetNotification, pagination request.Pagination) ([]domain.Notification, error)
	SendNotificationToDevice(ctx context.Context, notification request.Notification) ([]domain.Notification, error)
	GenerateFCMToken(ctx context.Context, request request.NotificationDeviceToken) error
}
