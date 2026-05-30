package domain

import (
	"time"
)

type Coupon struct {
	CouponID   string `json:"coupon_id" gorm:"primaryKey;type:varchar(32)"`
	CouponName string `json:"coupon_name" gorm:"unique;not null" binding:"required,min=3,max=25"`
	CouponCode string `json:"coupon_code" gorm:"unique;not null"`

	ExpireDate       time.Time `json:"expire_date" gorm:"not null"`
	Description      string    `json:"description" gorm:"not null" binding:"required,min=6,max=150"`
	DiscountRate     uint      `json:"discount_rate" gorm:"not null" binding:"required,numeric,min=1,max=100"`
	MinimumCartPrice Money     `json:"minimum_cart_price" gorm:"embedded;embeddedPrefix:minimum_cart_price_"`
	Image            string    `json:"image" binding:"required"`
	BlockStatus      bool      `json:"block_status" gorm:"not null"`
	CreatedBy        string    `json:"created_by" gorm:"type:varchar(32)"`
	UpdatedBy        string    `json:"updated_by" gorm:"type:varchar(32)"`
	CreatedAt        time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt        time.Time `json:"updated_at"`
}

// which is for store the user who are used coupon
type CouponUses struct {
	CouponUsesID string    `json:"coupon_uses_id" gorm:"primaryKey;type:varchar(32)"`
	CouponID     string    `json:"coupon_id" gorm:"type:varchar(32);not null"`
	Coupon       Coupon    `json:"-"`
	UserID       string    `json:"user_id" gorm:"type:varchar(32);not null"`
	User         User      `json:"-"`
	UsedAt       time.Time `json:"used_at" gorm:"not null"`
}
