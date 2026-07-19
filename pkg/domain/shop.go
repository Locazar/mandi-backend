package domain

import (
	"time"

	"gorm.io/gorm"
)

type ShopDetails struct {
	ID        string `json:"id" gorm:"primaryKey;type:varchar(32)"`
	ShopID    string `json:"shop_id" gorm:"-"`
	AdminID   string `json:"admin_id" gorm:"type:varchar(32);uniqueIndex"`
	ShopName  string `json:"shop_name" gorm:"size:100;"`
	OwnerName string `json:"owner_name" gorm:"size:100;"`
	Email     string `json:"email" gorm:"size:100;"`
	Phone     string `json:"phone" gorm:"size:50;"`

	AddressLine1 string  `json:"address_line1" gorm:"size:255"`
	AddressLine2 string  `json:"address_line2" gorm:"size:255" binding:"omitempty"`
	City         string  `json:"city" gorm:"size:50"`
	State        string  `json:"state" gorm:"size:50"`
	Country      string  `json:"country" gorm:"size:50"`
	Pincode      string  `json:"pincode" gorm:"size:50"`
	Latitude     float64 `json:"latitude" gorm:"type:decimal(10,7);"`
	Longitude    float64 `json:"longitude" gorm:"type:decimal(10,7);"`

	// PhoneVisibleConsent records the seller's opt-in, given on the first
	// onboarding step right after phone OTP verification, to show Phone to
	// customers on the shop listing ("visible to customers to discover you").
	// Defaults to true (both here and in the DB column default) — sellers
	// are opted in unless they explicitly uncheck the consent box.
	PhoneVisibleConsent bool `json:"phone_visible_consent" gorm:"not null;default:true" binding:"omitempty"`

	ShopDescription      string           `json:"shop_description" gorm:"type:text" binding:"omitempty"`
	ShopVerificationDocs string           `json:"shop_verification_docs" gorm:"type:text;" binding:"omitempty"`
	Document_Type        ShopDocumentType `json:"document_type" gorm:"size:50" binding:"omitempty"`
	Document_Value       string           `json:"document_value" gorm:"type:text" binding:"omitempty"`
	PanNumber            string           `json:"pan_number" gorm:"type:text" binding:"omitempty"`    // encrypted at rest
	ITRDocuments         string           `json:"itr_documents" gorm:"type:text" binding:"omitempty"` // encrypted at rest

	ShopType   ShopType       `json:"shop_type" gorm:"size:50" binding:"omitempty"`
	ShopStatus ShopStatusType `json:"shop_status" gorm:"size:50" binding:"omitempty"`

	BankAccountNumber string `json:"bank_account_number" gorm:"type:text" binding:"omitempty"` // encrypted at rest
	BankIFSC          string `json:"bank_ifsc" gorm:"type:text" binding:"omitempty"`           // encrypted at rest
	Shop_Image_URL    string `json:"shop_image_url" gorm:"size:255" binding:"omitempty"`

	ShopVerificationStatus     bool   `json:"shop_verification_status" gorm:"not null;default:false" binding:"omitempty"`
	ShopVerificationRemarks    string `json:"shop_verification_remarks" gorm:"not null;default:false" binding:"omitempty"`
	Photo_Shop_Verification    bool   `json:"photo_shop_verification" gorm:"not null;default:false" binding:"omitempty"`
	Business_Doc_Verification  bool   `json:"business_doc_verification" gorm:"not null;default:false" binding:"omitempty"`
	Identity_Doc_Verification  bool   `json:"identity_doc_verification" gorm:"not null;default:false" binding:"omitempty"`
	Address_Proof_Verification bool   `json:"address_proof_verification" gorm:"not null;default:false" binding:"omitempty"`

	Offers    []Offer `json:"offers" gorm:"many2many:shop_offers;"`
	HasOffers bool    `json:"has_offers" gorm:"column:has_offers"`

	// ProductLimit is the owning seller's admins.product_limit, joined in for
	// admin-panel display/edit; not a real shop_details column.
	ProductLimit int `json:"product_limit" gorm:"column:product_limit"`

	// ReferralCouponID (input-only, not persisted here) is the optional
	// referral code the seller enters on the final onboarding step. A match
	// against an Admin.ReferralCouponID creates a ShopReferral row instead
	// of a shop_details column, so re-submits never need to mutate this
	// table. ReferralStatus (output-only) reports the outcome back to the
	// client: "attached" or "invalid"; empty when no code was submitted.
	ReferralCouponID string `json:"referral_coupon_id,omitempty" gorm:"-" binding:"omitempty"`
	ReferralStatus   string `json:"referral_status,omitempty" gorm:"-"`

	CreatedAt time.Time      `json:"created_at" gorm:";autoCreateTime"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type ShopOffer struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	ShopID    string    `json:"shop_id" gorm:"type:varchar(32);not null"`
	OfferID   string    `json:"offer_id" gorm:"type:varchar(32);not null"`
	AdminID   string    `json:"admin_id" gorm:"type:varchar(32);not null"`
	StartDate time.Time `json:"start_date" gorm:"not null"`
	EndDate   time.Time `json:"end_date" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"not null;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type ShopDepartment struct {
	ID            string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	AdminID       string    `json:"admin_id" gorm:"type:varchar(32);not null;index"`
	ShopID        string    `json:"shop_id" gorm:"type:varchar(32);not null;index;uniqueIndex:idx_shop_department"`
	DepartmentID  string    `json:"department_id" gorm:"type:varchar(32);not null;uniqueIndex:idx_shop_department"`
	CategoryID    string    `json:"category_id" gorm:"type:varchar(32);not null;uniqueIndex:idx_shop_department"`
	SubCategoryID string    `json:"sub_category_id" gorm:"type:varchar(32);not null;uniqueIndex:idx_shop_department"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// TableName specifies the table name for ShopDepartment
func (ShopDepartment) TableName() string {
	return "shop_departments"
}

// ShopConflict is one pair of onboarded shops whose pinned locations fall
// within the configured detection radius of each other.
type ShopConflict struct {
	ShopAID     string    `json:"shop_a_id"`
	ShopAName   string    `json:"shop_a_name"`
	ShopAAdmin  string    `json:"shop_a_admin_id"`
	ShopBID     string    `json:"shop_b_id"`
	ShopBName   string    `json:"shop_b_name"`
	ShopBAdmin  string    `json:"shop_b_admin_id"`
	City        string    `json:"city"`
	Pincode     string    `json:"pincode"`
	DistanceM   float64   `json:"distance_m"`
	DetectedAt  time.Time `json:"detected_at"`
}
