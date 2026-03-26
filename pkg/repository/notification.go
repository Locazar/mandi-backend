package repository

import (
	"context"
	"fmt"

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
	return r.db.WithContext(ctx).Create(&notification).Error
}

func (r *notificationRepository) GetNotifications(ctx context.Context, filter request.GetNotification, pagination request.Pagination) ([]domain.Notification, error) {
	var notifications []domain.Notification
	query := r.db.WithContext(ctx).Model(&domain.Notification{})

	if filter.UserID != 0 {
		query = query.Where("receiver_id = ? AND receiver_type = 'user'", filter.UserID)
	}
	if filter.AdminID != 0 {
		query = query.Where("receiver_id = ? AND receiver_type = 'admin'", filter.AdminID)
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
	if filter.IsRead != nil {
		query = query.Where("is_read = ?", *filter.IsRead)
	}

	query = query.Limit(int(pagination.Limit)).Offset(int(pagination.Offset))

	if err := query.Order("created_at DESC").Find(&notifications).Error; err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *notificationRepository) MarkNotificationAsRead(ctx context.Context, notificationID uint) error {
	result := r.db.WithContext(ctx).
		Model(&domain.Notification{}).
		Where("id = ?", notificationID).
		Update("is_read", true)
	if result.RowsAffected == 0 {
		return fmt.Errorf("notification %d not found", notificationID)
	}
	return result.Error
}

// SaveDeviceToken upserts an FCM device token for a user or seller in Postgres.
func (r *notificationRepository) SaveDeviceToken(ctx context.Context, token domain.NotificationDeviceToken) error {
	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "token"}},
			DoUpdates: clause.AssignmentColumns([]string{"owner_id", "owner_type", "platform", "is_active", "updated_at"}),
		}).
		Create(&token).Error
}

// GetActiveTokensByOwner returns all active FCM tokens for a given owner (user/seller).
func (r *notificationRepository) GetActiveTokensByOwner(ctx context.Context, ownerID, ownerType string) ([]string, error) {
	var tokens []string
	err := r.db.WithContext(ctx).
		Model(&domain.NotificationDeviceToken{}).
		Where("owner_id = ? AND owner_type = ? AND is_active = true", ownerID, ownerType).
		Pluck("token", &tokens).Error
	return tokens, err
}

// DeleteDeviceToken marks an FCM token as inactive (soft delete).
func (r *notificationRepository) DeleteDeviceToken(ctx context.Context, ownerID, ownerType, token string) error {
	return r.db.WithContext(ctx).
		Model(&domain.NotificationDeviceToken{}).
		Where("owner_id = ? AND owner_type = ? AND token = ?", ownerID, ownerType, token).
		Update("is_active", false).Error
}

