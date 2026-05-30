package domain

import "time"

type Banner struct {
	ID           string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	Title        string    `json:"title" gorm:"size:255;not null"`
	Description  string    `json:"description" gorm:"size:500"`
	ImageURL     string    `json:"image_url" gorm:"size:500"`
	Link         string    `json:"link" gorm:"size:500"`
	Active       bool      `json:"active" gorm:"not null;default:true"`
	DepartmentID *string   `json:"department_id" gorm:"type:varchar(32);index"`
	CategoryID   *string   `json:"category_id" gorm:"type:varchar(32);index"`
	AppType      BannerAppType `json:"app_type" gorm:"size:50"`
	Latitude     float64   `json:"latitude" gorm:"type:decimal(10,7);default:0"`
	Longitude    float64   `json:"longitude" gorm:"type:decimal(10,7);default:0"`
	Pincode      string    `json:"pincode" gorm:"size:20"`
	RadiusKm     float64   `json:"radius_km" gorm:"type:decimal(10,2);default:0"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
