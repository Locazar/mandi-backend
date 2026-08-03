package domain

import "time"

// Language is a selectable UI/preference language offered to sellers during
// onboarding (feature-flag gated). Read-only from the client.
type Language struct {
	ID         string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	Code       string    `json:"code" gorm:"size:10;uniqueIndex;not null"`
	Name       string    `json:"name" gorm:"size:50;not null"`
	NativeName string    `json:"native_name" gorm:"size:50;not null"`
	SortOrder  int       `json:"sort_order" gorm:"not null;default:0"`
	IsActive   bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
