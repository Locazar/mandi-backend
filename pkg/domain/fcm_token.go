package domain

import "time"

type FcmToken struct {
	ID        string     `gorm:"primaryKey;type:varchar(32)" json:"id"`
	Token     string     `gorm:"column:token;unique;not null" json:"token"`
	Device    string     `gorm:"column:device" json:"device"`
	Platform  string     `gorm:"column:platform" json:"platform"`
	OwnerID   string     `gorm:"column:owner_id;not null" json:"owner_id"`
	OwnerType string     `gorm:"column:owner_type;not null" json:"owner_type"`
	IsActive  bool       `gorm:"column:is_active;not null;default:true" json:"is_active"`
	CreatedAt *time.Time `gorm:"column:created_at" json:"created_at,omitempty"`
	UpdatedAt *time.Time `gorm:"column:updated_at" json:"updated_at,omitempty"`

	// Compatibility fields used by existing handlers/usecases for owner resolution.
	ShopID  string `gorm:"-" json:"shop_id"`
	AdminID string `gorm:"-" json:"admin_id"`
}
