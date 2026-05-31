package domain

import (
	"time"

	"gorm.io/gorm"
)

// AdminStatus is the lifecycle state of an admin/seller account.
type AdminStatus string

const (
	AdminStatusActive    AdminStatus = "active"
	AdminStatusInactive  AdminStatus = "inactive"
	AdminStatusSuspended AdminStatus = "suspended"
)

func (s AdminStatus) IsValid() bool {
	switch s {
	case AdminStatusActive, AdminStatusInactive, AdminStatusSuspended:
		return true
	}
	return false
}

// AdminRole is the permission level of an admin account.
type AdminRole string

const (
	AdminRoleSuperAdmin       AdminRole = "super_admin"
	AdminRoleSupportStaff     AdminRole = "support_staff"
	AdminRoleCatalogManager   AdminRole = "catalog_manager"
	AdminRoleMarketingManager AdminRole = "marketing_manager"
	AdminRoleSeller           AdminRole = "seller"
)

func (r AdminRole) IsValid() bool {
	switch r {
	case AdminRoleSuperAdmin, AdminRoleSupportStaff, AdminRoleCatalogManager, AdminRoleMarketingManager, AdminRoleSeller:
		return true
	}
	return false
}

// VerificationStatusType is the shop verification review state.
type VerificationStatusType string

const (
	VerificationStatusVerified    VerificationStatusType = "verified"
	VerificationStatusUnverified  VerificationStatusType = "unverified"
	VerificationStatusUnderReview VerificationStatusType = "under_review"
)

func (s VerificationStatusType) IsValid() bool {
	switch s {
	case VerificationStatusVerified, VerificationStatusUnverified, VerificationStatusUnderReview:
		return true
	}
	return false
}

// AdvertisementStatus is the lifecycle state of an advertisement.
type AdvertisementStatus string

const (
	AdvertisementStatusActive   AdvertisementStatus = "active"
	AdvertisementStatusInactive AdvertisementStatus = "inactive"
	AdvertisementStatusExpired  AdvertisementStatus = "expired"
)

func (s AdvertisementStatus) IsValid() bool {
	switch s {
	case AdvertisementStatusActive, AdvertisementStatusInactive, AdvertisementStatusExpired:
		return true
	}
	return false
}

// AdvertisementPriority is the display priority of an advertisement.
type AdvertisementPriority string

const (
	AdvertisementPriorityHigh   AdvertisementPriority = "high"
	AdvertisementPriorityMedium AdvertisementPriority = "medium"
	AdvertisementPriorityLow    AdvertisementPriority = "low"
)

func (p AdvertisementPriority) IsValid() bool {
	switch p {
	case AdvertisementPriorityHigh, AdvertisementPriorityMedium, AdvertisementPriorityLow:
		return true
	}
	return false
}

type Admin struct {
	ID       string `json:"id" gorm:"primaryKey;type:varchar(32)"`
	AdminID  string `json:"admin_id" gorm:"-"`
	UserName string `json:"user_name"`
	FullName string `json:"full_name"`
	Email    string `json:"email" binding:"omitempty,email"`
	Password string `json:"-"`

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

	// Encrypted at rest (AES-256-GCM) by the repository; columns hold ciphertext.
	BankAccountNumber string `json:"bank_account_number" gorm:"type:text" binding:"omitempty"`
	BankIFSC          string `json:"bank_ifsc" gorm:"type:text" binding:"omitempty"`
	PAN               string `json:"pan" gorm:"type:text" binding:"omitempty"`

	// Aadhaar is minimized per DPDP: the full number is never persisted, only
	// the last four digits plus a verification flag.
	AadhaarLast4    string `json:"aadhaar_last4" gorm:"size:4" binding:"omitempty"`
	AadhaarVerified bool   `json:"aadhaar_verified" gorm:"not null;default:false"`

	CreatedAt time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`

	VerifiedSeller bool        `json:"verified_seller" gorm:"not null;default:false"`
	Status         AdminStatus `json:"status" gorm:"size:50"`
	Role           AdminRole   `json:"role" gorm:"size:50;not null;default:'super_admin'"`

	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type ShopVerification struct {
	ID                 string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	AdminID            string    `json:"admin_id" gorm:"type:varchar(32)" binding:"required"`
	ShopID             string    `json:"shop_id" gorm:"type:varchar(32)"`
	ShopName           string    `json:"shop_name"`
	VerificationStatus bool      `json:"verification_status" gorm:"not null;default:false"`
	Remarks            string    `json:"remarks" binding:"omitempty"`
	AgentID            string    `json:"agent_id" gorm:"type:varchar(32)" binding:"omitempty"`
	CreatedBy          string    `json:"created_by" gorm:"type:varchar(32)"`
	UpdatedBy          string    `json:"updated_by" gorm:"type:varchar(32)"`
	CreatedAt          time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type ShopVerificationHistory struct {
	ID                 string                 `json:"id" gorm:"primaryKey;type:varchar(32)"`
	AdminID            string                 `json:"admin_id" gorm:"type:varchar(32);not null"`
	ShopID             string                 `json:"shop_id" gorm:"type:varchar(32);not null"`
	VerificationStatus VerificationStatusType `json:"verification_status" gorm:"not null"`
	Remarks            string                 `json:"remarks" gorm:"size:255"`
	ChangedAt          time.Time              `json:"changed_at" gorm:"not null;autoCreateTime"`
}

type Advertisement struct {
	ID              string                `json:"id" gorm:"primaryKey;type:varchar(32)"`
	Title           string                `json:"title" gorm:"size:100" binding:"required"`
	Content         string                `json:"content" gorm:"type:text" binding:"required"`
	ImageURL        string                `json:"image_url" gorm:"size:255" binding:"omitempty"`
	TargetURL       string                `json:"target_url" gorm:"size:255" binding:"omitempty"`
	StartDate       time.Time             `json:"start_date" gorm:"not null" binding:"required"`
	EndDate         time.Time             `json:"end_date" gorm:"not null" binding:"required"`
	CreatedAt       time.Time             `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt       time.Time             `json:"updated_at" gorm:"autoUpdateTime"`
	CreatedByAdmin  string                `json:"created_by_admin" gorm:"type:varchar(32);not null"`
	AdminID         string                `json:"admin_id" gorm:"type:varchar(32);index"`
	AreaTargeted    string                `json:"area_targeted" gorm:"size:255" binding:"omitempty"`
	PincodeTargeted string                `json:"pincode_targeted" gorm:"size:20" binding:"omitempty"`
	Latitude        float64               `json:"latitude" gorm:"type:decimal(10,7);"`
	Longitude       float64               `json:"longitude" gorm:"type:decimal(10,7);"`
	DistanceKM      float64               `json:"distance_km" gorm:"type:decimal(10,2);"`
	Status          AdvertisementStatus   `json:"status" gorm:"size:50"`
	Priority        AdvertisementPriority `json:"priority" gorm:"size:20"`
	CreatedBy       string                `json:"created_by" gorm:"type:varchar(32)"`
	UpdatedBy       string                `json:"updated_by" gorm:"type:varchar(32)"`
}

type SubTypeAttributes struct {
	ID            string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	SubCategoryID string    `json:"sub_category_id" gorm:"type:varchar(32);not null" binding:"required"`
	FieldName     string    `json:"field_name" gorm:"size:50" binding:"required"`
	FieldType     FieldType `json:"field_type" gorm:"size:20" binding:"required"` // dropdown, number, text
	IsRequired    bool      `json:"is_required" gorm:"not null;default:true"`
	SortOrder     int       `json:"sort_order" gorm:"not null;default:0"`
}
type SubTypeAttributeOptions struct {
	ID                 string `json:"id" gorm:"primaryKey;type:varchar(32)"`
	SubTypeAttributeID string `json:"sub_type_attribute_id" gorm:"type:varchar(32);not null" binding:"required"`
	SubTypeAttribute   SubTypeAttributes
	OptionValue        string `json:"option_value" gorm:"size:50" binding:"required"`
	SortOrder          int    `json:"sort_order" gorm:"not null;default:0"`
}

type CategoryImage struct {
	ID         string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	CategoryID string    `json:"category_id" gorm:"type:varchar(32);not null" binding:"required"`
	ImageURL   string    `json:"image_url" gorm:"not null" binding:"required"`
	AltText    string    `json:"alt_text" gorm:"size:255" binding:"omitempty"`
	SortOrder  int       `json:"sort_order" gorm:"not null;default:0"`
	IsActive   bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt  time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type DashboardStats struct {
	TotalSellers         int64   `json:"total_sellers"`
	ActiveSellers        int64   `json:"active_sellers"`
	TotalShops           int64   `json:"total_shops"`
	VerifiedShops        int64   `json:"verified_shops"`
	PendingVerifications int64   `json:"pending_verifications"`
	TotalOrders          int64   `json:"total_orders"`
	TotalCustomers       int64   `json:"total_customers"`
	TotalRevenue         float64 `json:"total_revenue"`
	TotalProducts        int64   `json:"total_products"`
}
