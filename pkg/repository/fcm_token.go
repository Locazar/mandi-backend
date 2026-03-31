package repository

import (
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type fcmTokenRepository struct {
	db *gorm.DB
}

func NewFcmTokenRepository(db *gorm.DB) interfaces.FcmTokenRepository {
	return &fcmTokenRepository{db}
}

func (r *fcmTokenRepository) SaveFcmToken(fcmToken domain.FcmToken) (domain.FcmToken, error) {
	// Try to find by Token, ShopID, or AdminID
	var existing domain.FcmToken
	err := r.db.Where("token = ?", fcmToken.Token).First(&existing).Error
	if err == nil {
		// Update existing by token
		existing.Device = fcmToken.Device
		existing.Platform = fcmToken.Platform
		existing.ShopID = fcmToken.ShopID
		existing.AdminID = fcmToken.AdminID
		err = r.db.Save(&existing).Error
		return existing, err
	}
	// If not found by token, try by ShopID or AdminID
	tx := r.db
	if fcmToken.ShopID != 0 {
		tx = tx.Or("shop_id = ?", fcmToken.ShopID)
	}
	err = tx.First(&existing).Error
	if err == nil {
		// Update existing by shop/admin
		existing.Token = fcmToken.Token
		existing.Device = fcmToken.Device
		existing.Platform = fcmToken.Platform
		err = r.db.Save(&existing).Error
		return existing, err
	}
	// Not found, create new
	err = r.db.Create(&fcmToken).Error
	return fcmToken, err
}

// UpsertDeviceToken writes the token into notification_device_tokens so that
// SendPushNotification can look up tokens from Postgres without requiring Firestore.
func (r *fcmTokenRepository) UpsertDeviceToken(token, ownerID, ownerType, platform string) error {
	now := time.Now()
	record := domain.NotificationDeviceToken{
		OwnerID:   ownerID,
		OwnerType: ownerType,
		Token:     token,
		Platform:  platform,
		IsActive:  true,
		UpdatedAt: &now,
	}
	return r.db.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "token"}},
			DoUpdates: clause.AssignmentColumns([]string{"owner_id", "owner_type", "platform", "is_active", "updated_at"}),
		}).
		Create(&record).Error
}
