package domain

import "time"

type ShopTime struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	ShopID    string    `json:"shop_id" gorm:"type:varchar(32);not null"`
	Status    ShopTimeStatus `json:"status" gorm:"size:20;not null"` // "open" or "close"
	OpenTime  string    `json:"open_time" gorm:"not null"`
	CloseTime string    `json:"close_time" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
