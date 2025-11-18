package usecase

import (
	"context"
	"errors"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	service "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type notificationUseCase struct {
	notificationRepo interfaces.NotificationRepository
}

func NewNotificationUseCase(repo interfaces.NotificationRepository) service.NotificationUseCase {
	return &notificationUseCase{
		notificationRepo: repo,
	}
}

func (c *notificationUseCase) SaveNotification(ctx context.Context, notification request.Notification) error {

	newNotification := domain.Notification{
		SenderType:           notification.SenderType,
		ReceiverType:         notification.ReceiverType,
		SenderID:             notification.SenderID,
		Title:                notification.Title,
		Message:              notification.Message,
		Body:                 notification.Body,
		IsRead:               false,
		ReceiverID:           notification.ReceiverID,
		ShopID:               notification.ShopID,
		OrderID:              notification.OrderID,
		ProductID:            notification.ProductID,
		OfferID:              notification.OfferID,
		CategoryID:           notification.CategoryID,
		NotificationMetaData: notification.NotificationMetaData,
	}

	err := c.notificationRepo.SaveNotification(ctx, newNotification)
	if err != nil {
		return errors.New("failed to save notification")
	}

	return nil
}

func (c *notificationUseCase) GetNotificationsBy(ctx context.Context, filter request.GetNotification, pagination request.Pagination) ([]domain.Notification, error) {
	// Implementation goes here
	err := c.notificationRepo.GetNotification(ctx, request.Notification{})
	if err != nil {
		return nil, errors.New("failed to get notifications by filter")
	}
	return []domain.Notification{}, nil
}

func (c *notificationUseCase) MarkNotificationAsRead(ctx context.Context, notificationID uint) error {
	// Implementation goes here
	err := c.notificationRepo.MarkNotificationAsRead(ctx, notificationID)
	if err != nil {
		return errors.New("failed to mark notification as read")
	}
	return nil
}

func (c *notificationUseCase) SendNotificationToDevice(ctx context.Context, notification request.Notification) ([]domain.Notification, error) {
	// Implementation for sending notification to device
	err := c.notificationRepo.GetNotification(ctx, notification)
	if err != nil {
		return nil, errors.New("failed to send notification to device")
	}
	return []domain.Notification{}, nil
}

func (c *notificationUseCase) GenerateFCMToken(ctx context.Context, tokenRequest request.NotificationDeviceToken) error {
	// Implementation for generating FCM token
	err := c.notificationRepo.GenerateFCMToken(ctx, tokenRequest)
	if err != nil {
		return errors.New("failed to generate FCM token")
	}
	return nil
}
