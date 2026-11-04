package domain

import "time"

// EnquiryType values for ShopProductEnquiry.
const (
	EnquiryTypeProductAvailability = "product_availability"
	EnquiryTypeCustomerQuery       = "customer_query"
)

// ShopProductEnquiry is submitted by a website visitor on a shop profile page
// to ask the shop about products they carry or anything else they need.
// Admin portal lists these enquiries and can notify the seller to add products.
type ShopProductEnquiry struct {
	ID            string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	ShopID        string    `json:"shop_id" gorm:"type:varchar(100);not null;index"`
	ShopName      string    `json:"shop_name" gorm:"type:varchar(255)"`
	EnquiryType   string    `json:"enquiry_type" gorm:"type:varchar(50);not null"` // see EnquiryType* consts
	Message       string    `json:"message" gorm:"type:text;not null"`
	CustomerPhone string    `json:"customer_phone" gorm:"type:varchar(20)"`
	CustomerCity  string    `json:"customer_city" gorm:"type:varchar(100)"`
	IsActioned    bool      `json:"is_actioned" gorm:"not null;default:false"`
	AdminNote     string    `json:"admin_note" gorm:"type:text"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
