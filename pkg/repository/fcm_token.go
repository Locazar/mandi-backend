package repository

import (
	"fmt"
	"strings"
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
		if fcmToken.ShopID != "" {
			fcmToken.OwnerID = fcmToken.ShopID
		} else if fcmToken.AdminID != "" {
			fcmToken.OwnerID = fcmToken.AdminID
		}
	}
	if fcmToken.ShopID == "" && fcmToken.OwnerType == "seller" {
		fcmToken.ShopID = fcmToken.OwnerID
	}
	if fcmToken.AdminID == "" {
		if fcmToken.OwnerType == "seller" {
			fcmToken.AdminID = fcmToken.OwnerID
		} else {
			fcmToken.AdminID = fcmToken.ShopID
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

func (r *fcmTokenRepository) UnregisterFcmToken(fcmToken domain.FcmToken) error {
	fmt.Printf("Unregistering FCM token: %s for owner_id: %s owner_type: %s\n", fcmToken.Token, fcmToken.OwnerID, fcmToken.OwnerType)
	if fcmToken.Token == "" {
		return fmt.Errorf("token is required")
	}

	if err := r.ensureNotificationDeviceTokenTable(); err != nil {
		return err
	}

	shopID := fcmToken.ShopID
	if shopID == "" && fcmToken.OwnerID != "" && (fcmToken.OwnerType == "" || fcmToken.OwnerType == "seller") {
		shopID = fcmToken.OwnerID
	}

	fcmQuery := r.db.Where("token = ?", fcmToken.Token)
	if shopID != "" {
		fcmQuery = fcmQuery.Where("owner_id = ?", shopID)
	}
	if err := fcmQuery.Delete(&domain.FcmToken{}).Error; err != nil {
		return err
	}

	ndtQuery := r.db.Where("token = ?", fcmToken.Token)
	if shopID != "" {
		ndtQuery = ndtQuery.Where("shop_id = ?", shopID)
	}
	if err := ndtQuery.Delete(&domain.NotificationDeviceToken{}).Error; err != nil {
		if isUndefinedTableError(err) {
			if ensureErr := r.ensureNotificationDeviceTokenTable(); ensureErr != nil {
				return ensureErr
			}
			if retryErr := ndtQuery.Delete(&domain.NotificationDeviceToken{}).Error; retryErr != nil {
				return retryErr
			}
			return nil
		}
		return err
	}

	return nil
}

// UpsertDeviceToken writes the token into notification_device_tokens so that
// SendPushNotification can look up tokens from Postgres without requiring Firestore.
func (r *fcmTokenRepository) UpsertDeviceToken(token, ownerID, ownerType, platform, shopID, adminID string) error {
	fmt.Printf("Upserting device token into notification_device_tokens: token=%s, owner_id=%s, owner_type=%s, platform=%s, shop_id=%s, admin_id=%s\n", token, ownerID, ownerType, platform, shopID, adminID)
	if err := r.ensureNotificationDeviceTokenTable(); err != nil {
		return err
	}
	if shopID == "" {
		if ownerType == "seller" {
			shopID = ownerID
		} else {
			shopID = adminID
		}
	}
	if adminID == "" && ownerType == "seller" {
		adminID = shopID
	}
	if adminID == "" {
		adminID = "0"
	}
	if shopID == "" {
		shopID = "0"
	}
	now := time.Now()
	record := domain.NotificationDeviceToken{
		OwnerID:   ownerID,
		OwnerType: ownerType,
		ShopID:    shopID,
		AdminID:   adminID,
		Token:     token,
		Platform:  platform,
		IsActive:  true,
		UpdatedAt: &now,
	}
	err := r.db.
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "token"}},
			DoUpdates: clause.AssignmentColumns([]string{"owner_id", "owner_type", "shop_id", "admin_id", "platform", "is_active", "updated_at"}),
		}).
		Create(&record).Error

	if isUndefinedTableError(err) {
		if ensureErr := r.ensureNotificationDeviceTokenTable(); ensureErr != nil {
			return ensureErr
		}
		return r.db.
			Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "token"}},
				DoUpdates: clause.AssignmentColumns([]string{"owner_id", "owner_type", "shop_id", "admin_id", "platform", "is_active", "updated_at"}),
			}).
			Create(&record).Error
	}

	return err
}

func (r *fcmTokenRepository) ensureNotificationDeviceTokenTable() error {
	if r.db.Migrator().HasTable(&domain.NotificationDeviceToken{}) {
		return nil
	}
	return r.db.AutoMigrate(&domain.NotificationDeviceToken{})
}

func isUndefinedTableError(err error) bool {
	if err == nil {
		return false
	}
	errMsg := strings.ToLower(err.Error())
	return strings.Contains(errMsg, "does not exist") || strings.Contains(errMsg, "sqlstate 42p01")
}

