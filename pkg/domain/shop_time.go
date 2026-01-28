package domain

import "time"

type ShopTime struct {
	ID        uint      `json:"id" gorm:"primaryKey;not null"`
	ShopID    uint      `json:"shop_id" gorm:"not null"`
	Status    string    `json:"status" gorm:"size:20;not null"` // "open" or "close"
	OpenTime  time.Time `json:"open_time" gorm:"not null"`
	CloseTime time.Time `json:"close_time" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}