package domain

import "time"

type Banner struct {
	ID          uint      `json:"id" gorm:"primaryKey;not null"`
	Title       string    `json:"title" gorm:"size:255;not null"`
	Description string    `json:"description" gorm:"size:500"`
	ImageURL    string    `json:"image_url" gorm:"size:500"`
	Link        string    `json:"link" gorm:"size:500"`
	Active      bool      `json:"active" gorm:"not null;default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
