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

// AdvertisementAudience distinguishes who sees the advertisement.
type AdvertisementAudience string

const (
	AdvertisementAudienceCustomer AdvertisementAudience = "customer"
	AdvertisementAudienceSeller   AdvertisementAudience = "seller"
)

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
	// Audience segmentation
	Audience     AdvertisementAudience `json:"audience" gorm:"size:20;default:'customer'"`
	DepartmentID string                `json:"department_id" gorm:"type:varchar(32);default:null"`
	CategoryID   string                `json:"category_id" gorm:"type:varchar(32);default:null"`
}

// AdvertisementRequestStatus is the review state of a seller's ad request.
type AdvertisementRequestStatus string

const (
	AdvertRequestStatusPending  AdvertisementRequestStatus = "pending"
	AdvertRequestStatusApproved AdvertisementRequestStatus = "approved"
	AdvertRequestStatusRejected AdvertisementRequestStatus = "rejected"
)

// AdvertRequestPaymentStatus tracks the payment lifecycle of an approved request.
type AdvertRequestPaymentStatus string

const (
	AdvertPaymentUnpaid  AdvertRequestPaymentStatus = "unpaid"
	AdvertPaymentPending AdvertRequestPaymentStatus = "pending" // order created, awaiting verification
	AdvertPaymentPaid    AdvertRequestPaymentStatus = "paid"
	AdvertPaymentFailed  AdvertRequestPaymentStatus = "failed"
)

// AdvertisementRequest is a seller-raised request to run an advertisement
// for a date range at a quoted price plan.
type AdvertisementRequest struct {
	ID           string                     `json:"id" gorm:"primaryKey;type:varchar(32)"`
	AdminID      string                     `json:"admin_id" gorm:"type:varchar(32);index;not null"`
	ShopID       string                     `json:"shop_id" gorm:"type:varchar(32);index"`
	ShopName     string                     `json:"shop_name" gorm:"-"`
	Title        string                     `json:"title" gorm:"size:100"`
	Content      string                     `json:"content" gorm:"type:text"`
	StartDate    time.Time                  `json:"start_date" gorm:"not null"`
	EndDate      time.Time                  `json:"end_date" gorm:"not null"`
	PlanKey      string                     `json:"plan_key" gorm:"size:20;not null"`
	PriceMinor   int64                      `json:"price_minor" gorm:"not null"`
	Status       AdvertisementRequestStatus `json:"status" gorm:"size:20;not null;default:'pending'"`
	AdminComment string                     `json:"admin_comment" gorm:"type:text"`
	CreatedAt    time.Time                  `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt    time.Time                  `json:"updated_at" gorm:"autoUpdateTime"`

	// Payment tracking (populated once the request is approved and paid for).
	PaymentStatus  AdvertRequestPaymentStatus `json:"payment_status" gorm:"size:20;not null;default:'unpaid'"`
	PaymentOrderID string                     `json:"payment_order_id" gorm:"size:64"`
	PaymentID      string                     `json:"payment_id" gorm:"size:64"`
	PaidAt         *time.Time                 `json:"paid_at"`
}

// AdvertisementInvoice is the full charge breakdown for an approved request.
// All amounts are in minor units (paise).
type AdvertisementInvoice struct {
	RequestID        string  `json:"request_id"`
	PlanKey          string  `json:"plan_key"`
	PlanName         string  `json:"plan_name"`
	Days             int     `json:"days"`
	RatePerDay       int64   `json:"rate_per_day"`
	BaseMinor        int64   `json:"base_minor"`
	PlatformFeeMinor int64   `json:"platform_fee_minor"`
	GSTRatePercent   float64 `json:"gst_rate_percent"`
	GSTMinor         int64   `json:"gst_minor"`
	TotalMinor       int64   `json:"total_minor"`
}

// FeatureFlag is an admin-managed on/off switch for a client-side feature.
type FeatureFlag struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	FlagKey     string    `json:"flag_key" gorm:"size:64;uniqueIndex;not null"`
	Enabled     bool      `json:"enabled" gorm:"not null;default:false"`
	Description string    `json:"description" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// HelpSettings holds the singleton support-contact configuration shown on the
// Help & Support screen in client apps (seller app, and future customer app).
type HelpSettings struct {
	ID             string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	SupportPhone   string    `json:"support_phone" gorm:"size:20"`
	SupportEmail   string    `json:"support_email" gorm:"size:120"`
	WhatsAppNumber string    `json:"whatsapp_number" gorm:"size:20"`
	SupportHours   string    `json:"support_hours" gorm:"size:120"`
	AboutText      string    `json:"about_text" gorm:"type:text"`
	CreatedAt      time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// HelpFAQ is a single admin-managed frequently-asked-question entry shown on
// the Help & Support screen. SortOrder controls display order (ascending).
type HelpFAQ struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	Question  string    `json:"question" gorm:"type:text;not null"`
	Answer    string    `json:"answer" gorm:"type:text;not null"`
	SortOrder int       `json:"sort_order" gorm:"not null;default:0"`
	IsActive  bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// HelpCenter bundles contact settings + active FAQs for the single public
// read endpoint consumed by client apps.
type HelpCenter struct {
	Settings HelpSettings `json:"settings"`
	FAQs     []HelpFAQ    `json:"faqs"`
}

// AppConfig is an admin-managed key/value setting stored in the database.
type AppConfig struct {
	ID          string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	ConfigKey   string    `json:"config_key" gorm:"size:64;uniqueIndex;not null"`
	Value       string    `json:"value" gorm:"type:text;not null"`
	Description string    `json:"description" gorm:"type:text"`
	Enabled     bool      `json:"enabled" gorm:"not null;default:true"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// globalAppConfigKey is the fixed app_configs.config_key under which the
// single structured GlobalAppConfig JSON blob is stored (see GET/PUT /api/admin/config).
const GlobalAppConfigKey = "global_app_config"

// GlobalAppConfig is the structured, admin-editable configuration that drives
// client app behavior (CDN URLs, feature flags, HTTP timeouts, etc.) without
// requiring a new app release. Stored as a single JSON blob in the app_configs
// table under GlobalAppConfigKey.
type GlobalAppConfig struct {
	ImageBaseURL     string `json:"imageBaseUrl"`
	ShopDeepLinkBase string `json:"shopDeepLinkBase"`
	AppDownloadURL   string `json:"appDownloadUrl"`

	// JSON key is "features" (not "featureFlags") to match the seller-app's
	// already-shipped RemoteConfigService contract (GET /api/app-config).
	FeatureFlags struct {
		ShowProductItemOffers     bool `json:"showProductItemOffers"`
		ShowOfferBadgesOnProducts bool `json:"showOfferBadgesOnProducts"`
	} `json:"features"`

	// JSON key is "links" (not "social") to match RemoteConfigService.
	Social struct {
		Instagram      string `json:"instagram"`
		Facebook       string `json:"facebook"`
		SupportEmail   string `json:"supportEmail"`
		PrivacyPolicy  string `json:"privacyPolicy"`
		TermsOfService string `json:"termsOfService"`
	} `json:"links"`

	// Field names match RemoteConfigService.httpConfig exactly.
	Http struct {
		TimeoutSeconds       int `json:"timeoutSeconds"`
		UploadTimeoutSeconds int `json:"uploadTimeoutSeconds"`
		MaxRetryAttempts     int `json:"maxRetryAttempts"`
		RetryDelaySeconds    int `json:"retryDelaySeconds"`
	} `json:"http"`

	ImageUpload struct {
		MaxWidthPx  int `json:"maxWidthPx"`
		MaxHeightPx int `json:"maxHeightPx"`
		Quality     int `json:"quality"`
	} `json:"imageUpload"`

	// OpenAI is server-side only (never shipped to client apps) but is
	// admin-editable here so it can change without a backend redeploy.
	OpenAI struct {
		Model     string `json:"model"`
		ImageSize string `json:"imageSize"`
	} `json:"openai"`
}

// DefaultGlobalAppConfig mirrors the values currently hardcoded in the seller
// app's AppConfig (lib/core/config/app_config.dart), used as the fallback
// when no global_app_config row exists yet.
func DefaultGlobalAppConfig() GlobalAppConfig {
	cfg := GlobalAppConfig{
		ImageBaseURL:     "https://innoida.utho.io/locazar-dev",
		ShopDeepLinkBase: "https://locazar.com/shop",
		AppDownloadURL:   "https://locazar.app",
	}
	cfg.FeatureFlags.ShowProductItemOffers = false
	cfg.FeatureFlags.ShowOfferBadgesOnProducts = false
	cfg.Social.Instagram = "https://www.instagram.com/official_locazar"
	cfg.Social.Facebook = "https://www.facebook.com/profile.php?id=61586142238127"
	cfg.Social.SupportEmail = "support@locazar.co.in"
	cfg.Social.PrivacyPolicy = "https://locazar.co.in/privacy-policy"
	cfg.Social.TermsOfService = "https://locazar.co.in/terms-of-service"
	cfg.Http.TimeoutSeconds = 15
	cfg.Http.UploadTimeoutSeconds = 60
	cfg.Http.MaxRetryAttempts = 3
	cfg.Http.RetryDelaySeconds = 2
	cfg.ImageUpload.MaxWidthPx = 1024
	cfg.ImageUpload.MaxHeightPx = 1024
	cfg.ImageUpload.Quality = 85
	cfg.OpenAI.Model = "gpt-image-1"
	cfg.OpenAI.ImageSize = "1024x1024"
	return cfg
}

// AdvertisementPlanConfig is an admin-managed price plan stored in the DB.
// Sellers see only active plans; rates are in minor units (paise).
type AdvertisementPlanConfig struct {
	ID              string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	PlanKey         string    `json:"plan_key" gorm:"size:20;uniqueIndex;not null"`
	Name            string    `json:"name" gorm:"size:100;not null"`
	Description     string    `json:"description" gorm:"type:text"`
	RatePerDayMinor int64     `json:"rate_per_day_minor" gorm:"not null"`
	SortOrder       int       `json:"sort_order" gorm:"not null;default:0"`
	IsActive        bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// AdvertisementPricingConfig is the admin-managed singleton holding the
// tax/fee percentages applied on top of the base plan price.
type AdvertisementPricingConfig struct {
	ID                 string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	GSTRatePercent     float64   `json:"gst_rate_percent" gorm:"type:numeric(5,2);not null"`
	PlatformFeePercent float64   `json:"platform_fee_percent" gorm:"type:numeric(5,2);not null"`
	UpdatedAt          time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// AdvertisementPricePlan is a quoted plan for a requested date range.
type AdvertisementPricePlan struct {
	PlanKey      string `json:"plan_key"`      // high | medium | low
	Name         string `json:"name"`          // display name
	Description  string `json:"description"`   // what the plan includes
	Days         int    `json:"days"`          // number of days in the range
	RatePerDay   int64  `json:"rate_per_day"`  // minor units (paise)
	TotalMinor   int64  `json:"total_minor"`   // minor units (paise)
}

// AdvertisementFilter holds query-time filter parameters for admin list.
type AdvertisementFilter struct {
	// existing app-facing fields
	Latitude  float64
	Longitude float64
	RadiusKM  float64               // 0 means no geo filter
	Pincode   string
	AppType   AdvertisementAudience // "seller" | "customer" | "" (all)

	// admin filter fields
	DepartmentID    string
	CategoryID      string
	PincodeTargeted string
	Status          AdvertisementStatus
	Audience        AdvertisementAudience
	Priority        AdvertisementPriority
	AdminID         string
	// date range: zero value means no filter
	StartDateFrom time.Time // ads whose start_date >= this value
	EndDateTo     time.Time // ads whose end_date <= this value
	// geo proximity: all three must be non-zero to apply
	FilterLatitude  float64
	FilterLongitude float64
	DistanceKM      float64
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
