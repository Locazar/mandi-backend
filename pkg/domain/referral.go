package domain

import (
	"time"

	"gorm.io/gorm"
)

// ShopReferralStatus is the lifecycle state of a shop's attachment to a
// Sales Executive via referral coupon.
type ShopReferralStatus string

const (
	ShopReferralStatusActive   ShopReferralStatus = "active"
	ShopReferralStatusInactive ShopReferralStatus = "inactive"
)

func (s ShopReferralStatus) IsValid() bool {
	switch s {
	case ShopReferralStatusActive, ShopReferralStatusInactive:
		return true
	}
	return false
}

// ShopReferral attaches a shop to the Sales Executive — a platform user
// (domain.Admin) holding a ReferralCouponID — whose referral code the
// seller entered during onboarding. One row per shop: a shop is attached to
// at most one Sales Executive at a time, which is what drives that
// executive's "shops attached to me" view and commission/report rollups.
type ShopReferral struct {
	ID               string             `json:"id" gorm:"primaryKey;type:varchar(32)"`
	ReferralCouponID string             `json:"referral_coupon_id" gorm:"size:50;index;not null" binding:"required"`
	PlatformUserID   string             `json:"platform_user_id" gorm:"type:varchar(32);index;not null" binding:"required"`
	ShopID           string             `json:"shop_id" gorm:"type:varchar(32);uniqueIndex;not null" binding:"required"`
	SellerAdminID    string             `json:"seller_admin_id" gorm:"type:varchar(32);index"`
	Status           ShopReferralStatus `json:"status" gorm:"size:20;not null;default:'active'"`
	CreatedAt        time.Time          `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt        time.Time          `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt        gorm.DeletedAt     `json:"-" gorm:"index"`
}
