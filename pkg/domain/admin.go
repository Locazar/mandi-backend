package domain

import "time"

type Admin struct {
	ID       uint   `json:"id" gorm:"primaryKey;not null"`
	UserName string `json:"user_name" gorm:"not null" binding:"required,min=3,max=15"`
	Email    string `json:"email" gorm:"not null;uniqueIndex" binding:"required,email"`
	Password string `json:"password" gorm:"not null" binding:"required,min=5,max=30"`

	ShopName                string `json:"shop_name" gorm:"null" binding:"required"`
	GSTIN                   string `json:"gstin" gorm:"size:20" binding:"omitempty"`             // GST number (optional)
	ShopID                  string `json:"shop_id" gorm:"size:50" binding:"omitempty"`           // Shop registration number
	ElectricityBill         string `json:"electricity_bill" gorm:"size:255" binding:"omitempty"` // URL or path to electricity bill document
	ShopType                string `json:"shop_type" gorm:"size:50"`                             // e.g. "retail", "wholesale"
	ShopVerificationStatus  string `json:"shop_verification_status" gorm:"size:50"`              // e.g. "verified", "unverified", "under_review"
	ShopVerificationRemarks string `json:"shop_verification_remarks" gorm:"size:255"`            // remarks regarding verification

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

type ShopVerification struct {
	ID                 uint      `json:"id" gorm:"primaryKey;not null"`
	AdminID            uint      `json:"admin_id" binding:"required"`
	ShopID             string    `json:"shop_id" binding:"required"`
	ShopName           string    `json:"shop_name" binding:"required"`
	VerificationStatus string    `json:"verification_status" binding:"required"` // e.g. "verified", "unverified", "under_review"
	Remarks            string    `json:"remarks" binding:"omitempty"`
	AgentID            uint      `json:"agent_id" binding:"omitempty"`
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type ShopVerificationHistory struct {
	ID                 uint      `json:"id" gorm:"primaryKey;not null"`
	AdminID            uint      `json:"admin_id" gorm:"not null"`
	ShopID             uint      `json:"shop_id" gorm:"not null"`
	VerificationStatus string    `json:"verification_status" gorm:"not null"` // e.g. "verified", "unverified", "under_review"
	Remarks            string    `json:"remarks" gorm:"size:255"`
	ChangedAt          time.Time `json:"changed_at" gorm:"not null;autoCreateTime"`
}

type AgentDetails struct {
	ID        uint   `json:"id" gorm:"primaryKey;not null"`
	FirstName string `json:"first_name" gorm:"size:50" binding:"required"`
	LastName  string `json:"last_name" gorm:"size:50" binding:"required"`
	Email     string `json:"email" gorm:"size:100;uniqueIndex" binding:"required,email"`
	Phone     string `json:"phone" gorm:"size:15;uniqueIndex" binding:"required"`
}
