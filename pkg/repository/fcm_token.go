package repository

import (
	"fmt"
	"strconv"
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
	if fcmToken.OwnerID == "" {
		if fcmToken.ShopID != 0 {
			fcmToken.OwnerID = strconv.FormatUint(uint64(fcmToken.ShopID), 10)
		} else if fcmToken.AdminID != 0 {
			fcmToken.OwnerID = strconv.FormatUint(uint64(fcmToken.AdminID), 10)
		}
	}

	if fcmToken.OwnerType == "" {
		fcmToken.OwnerType = "seller"
	}

	if fcmToken.Token == "" || fcmToken.OwnerID == "" {
		return fcmToken, fmt.Errorf("token and owner_id are required")
	}

	fcmToken.IsActive = true
	now := time.Now()
	fcmToken.UpdatedAt = &now

	err := r.db.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "token"}},
			DoUpdates: clause.AssignmentColumns([]string{"device", "platform", "owner_id", "owner_type", "is_active", "updated_at"}),
		}).
		Create(&fcmToken).Error
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
