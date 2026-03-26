package domain

import "time"

type Admin struct {
	ID       uint   `json:"id" gorm:"primaryKey;not null"`
	FullName string `json:"full_name"  binding:"min=2,max=50"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"password" gorm:"not null" binding:"required,min=5,max=30"`

	AddressLine1    string  `json:"address_line1" gorm:"size:255"`
	AddressLine2    string  `json:"address_line2" gorm:"size:255" binding:"omitempty"`
	City            string  `json:"city" gorm:"size:50" `
	State           string  `json:"state" gorm:"size:50" `
	Country         string  `json:"country" gorm:"size:50" `
	Pincode         string  `json:"pincode" gorm:"size:50" `
	Mobile          string  `json:"mobile" gorm:"size:50"`
	ProfileImageUrl string  `json:"profile_image_url" gorm:"size:255" binding:"omitempty"`
	Latitude        float64 `json:"latitude" gorm:"type:decimal(10,7);"`
	Longitude       float64 `json:"longitude" gorm:"type:decimal(10,7);"`

	PaymentStatus bool      `json:"payment_status" gorm:"not null;default:false"`
	PaymentType   string    `json:"payment_type" gorm:"size:50"`
	PaymentDate   time.Time `json:"payment_date" gorm:""`
	StartDate     time.Time `json:"start_date" gorm:""`
	ExpiryDate    time.Time `json:"expiry_date" gorm:""`

	BankAccountNumber string `json:"bank_account_number" gorm:"size:50" binding:"omitempty"`
	BankIFSC          string `json:"bank_ifsc" gorm:"size:20" binding:"omitempty"`

	PAN    string `json:"pan" gorm:"size:20" binding:"omitempty"`
	Aadhar string `json:"aadhar" gorm:"size:20" binding:"omitempty"`

	AgreeToTerms bool `json:"agree_to_terms" gorm:"size:50"`

	CreatedAt time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	VerifiedSeller bool   `json:"verified_seller" gorm:"not null;default:false"` // e.g. "yes", "no", "pending"
	Status         string `json:"status" gorm:"size:50"`                         // e.g. "active", "inactive", "suspended"
}

type ShopVerification struct {
	ID                 uint      `json:"id" gorm:"primaryKey;not null"`
	AdminID            string    `json:"admin_id" binding:"required"`
	ShopID             uint      `json:"shop_id"`
	ShopName           string    `json:"shop_name"`
	VerificationStatus bool      `json:"verification_status" gorm:"not null;default:false"`
	Remarks            string    `json:"remarks" binding:"omitempty"`
	AgentID            uint      `json:"agent_id" binding:"omitempty"`
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type ShopVerificationHistory struct {
	ID                 uint      `json:"id" gorm:"primaryKey;not null"`
	AdminID            string    `json:"admin_id" gorm:"not null"`
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

type Advertisement struct {
	ID              uint      `json:"id" gorm:"primaryKey;not null"`
	Title           string    `json:"title" gorm:"size:100" binding:"required"`
	Content         string    `json:"content" gorm:"type:text" binding:"required"`
	ImageURL        string    `json:"image_url" gorm:"size:255" binding:"omitempty"`
	TargetURL       string    `json:"target_url" gorm:"size:255" binding:"omitempty"`
	StartDate       time.Time `json:"start_date" gorm:"not null" binding:"required"`
	EndDate         time.Time `json:"end_date" gorm:"not null" binding:"required"`
	CreatedAt       time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedByAdmin  uint      `json:"created_by_admin" gorm:"not null"`
	AdminID         string    `json:"admin_id" gorm:"not null"`
	AreaTargeted    string    `json:"area_targeted" gorm:"size:255" binding:"omitempty"`
	PincodeTargeted string    `json:"pincode_targeted" gorm:"size:20" binding:"omitempty"`
	Latitude        float64   `json:"latitude" gorm:"type:decimal(10,7);"`
	Longitude       float64   `json:"longitude" gorm:"type:decimal(10,7);"`
	DistanceKM      float64   `json:"distance_km" gorm:"type:decimal(10,2);"`
	Status          string    `json:"status" gorm:"size:50"`   // e.g. "active", "inactive", "expired"
	Priority        string    `json:"priority" gorm:"size:20"` // e.g. "high", "medium", "low"
}

type SubTypeAttributes struct {
	ID            uint   `json:"id" gorm:"primaryKey;not null"`
	SubCategoryID uint   `json:"sub_category_id" gorm:"not null" binding:"required,numeric"`
	FieldName     string `json:"field_name" gorm:"size:50" binding:"required"`
	FieldType     string `json:"field_type" gorm:"size:20" binding:"required"` // dropdown, number, text
	IsRequired    bool   `json:"is_required" gorm:"not null;default:true"`
	SortOrder     int    `json:"sort_order" gorm:"not null;default:0"`
}
type SubTypeAttributeOptions struct {
	ID                 uint `json:"id" gorm:"primaryKey;not null"`
	SubTypeAttributeID uint `json:"sub_type_attribute_id" gorm:"not null" binding:"required,numeric"`
	SubTypeAttribute   SubTypeAttributes
	OptionValue        string `json:"option_value" gorm:"size:50" binding:"required"`
	SortOrder          int    `json:"sort_order" gorm:"not null;default:0"`
}

type CategoryImage struct {
	ID         uint      `json:"id" gorm:"primaryKey;not null"`
	CategoryID uint      `json:"category_id" gorm:"not null" binding:"required,numeric"`
	ImageURL   string    `json:"image_url" gorm:"not null" binding:"required"`
	AltText    string    `json:"alt_text" gorm:"size:255" binding:"omitempty"`
	SortOrder  int       `json:"sort_order" gorm:"not null;default:0"`
	IsActive   bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
