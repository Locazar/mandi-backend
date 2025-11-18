package repository

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type notificationRepository struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) interfaces.NotificationRepository {
	return &notificationRepository{
		db: db,
	}
}

func (r *notificationRepository) SaveNotification(ctx context.Context, notification domain.Notification) error {
	result := r.db.WithContext(ctx).Create(&notification)
	return result.Error
}

func (r *notificationRepository) GetNotification(ctx context.Context, filter request.Notification) error {
	var notifications []domain.Notification
	query := r.db.WithContext(ctx).Model(&domain.Notification{})

	if filter.ReceiverID != 0 {
		query = query.Where("receiver_id = ?", filter.ReceiverID)
	}
	if filter.IsRead {
		query = query.Where("is_read = ?", filter.IsRead)
	}
	if filter.SenderID != 0 {
		query = query.Where("sender_id = ?", filter.SenderID)
	}
	if filter.ShopID != 0 {
		query = query.Where("shop_id = ?", filter.ShopID)
	}
	if filter.Status != "" {
		query = query.Where("status = ?", filter.Status)
	}
	if filter.ProductID != 0 {
		query = query.Where("product_id = ?", filter.ProductID)
	}
	if filter.OrderID != 0 {
		query = query.Where("order_id = ?", filter.OrderID)
	}
	if filter.VariationID != 0 {
		query = query.Where("variation_id = ?", filter.VariationID)
	}

	result := query.Find(&notifications)
	return result.Error
}

func (r *notificationRepository) MarkNotificationAsRead(ctx context.Context, notificationID uint) error {
	result := r.db.WithContext(ctx).Model(&domain.Notification{}).Where("id = ?", notificationID).Update("is_read", true)
	return result.Error
}

func (r *notificationRepository) GenerateFCMToken(ctx context.Context, req request.NotificationDeviceToken) error {
	// Save or update token in DB (upsert example)
	err := r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "token"}},
		DoUpdates: clause.AssignmentColumns([]string{"owner_id", "owner_type", "platform", "updated_at"}),
	}).Create(&req).Error
	if err != nil {
		return err
	}

	return nil
}
