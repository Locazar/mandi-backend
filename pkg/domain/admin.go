package domain

import "time"

type Admin struct {
	ID       uint   `json:"id" gorm:"primaryKey;not null"`
	UserName string `json:"user_name" gorm:"not null" binding:"required,min=3,max=15"`
	Email    string `json:"email" gorm:"not null;uniqueIndex" binding:"required,email"`
	Password string `json:"password" gorm:"not null" binding:"required,min=5,max=30"`

	ShopName string `json:"shop_name" gorm:"null" binding:"required"`
	GSTIN    string `json:"gstin" gorm:"size:20" binding:"omitempty"`   // GST number (optional)
	ShopID   string `json:"shop_id" gorm:"size:50" binding:"omitempty"` // Shop registration number

	AddressLine1 string  `json:"address_line1" gorm:"size:255" binding:"required"`
	AddressLine2 string  `json:"address_line2" gorm:"size:255" binding:"omitempty"`
	City         string  `json:"city" gorm:"size:50" binding:"required"`
	State        string  `json:"state" gorm:"size:50" binding:"required"`
	Country      string  `json:"country" gorm:"size:50" binding:"required"`
	Pincode      string  `json:"pincode" gorm:"size:50" binding:"required"`
	Mobile       string  `json:"mobile" gorm:"size:50" binding:"required"`
	Latitude     float64 `json:"latitude" gorm:"type:decimal(10,7);"`
	Longitude    float64 `json:"longitude" gorm:"type:decimal(10,7);"`

	BankAccountNumber string `json:"bank_account_number" gorm:"size:50" binding:"omitempty"`
	BankIFSC          string `json:"bank_ifsc" gorm:"size:20" binding:"omitempty"`

	PAN    string `json:"pan" gorm:"size:20" binding:"omitempty"`
	Aadhar string `json:"aadhar" gorm:"size:20" binding:"omitempty"`

	AgreeToTerms bool `json:"agree_to_terms" gorm:"size:50" binding:"required"`

	CreatedAt time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	Verified string `json:"verified" gorm:"size:20"` // e.g. "yes", "no", "pending"
	Status   string `json:"status" gorm:"size:50"`   // e.g. "active", "inactive", "suspended"
}
