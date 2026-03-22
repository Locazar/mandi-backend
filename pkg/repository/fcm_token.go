package repository

import (
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
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
	if fcmToken.AdminID != 0 {
		tx = tx.Or("admin_id = ?", fcmToken.AdminID)
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
