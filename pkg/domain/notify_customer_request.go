package domain

import "time"

// NotifyCustomerRequest is submitted by a seller on a completed or rejected
// deal to inform Locazar that the customer should be contacted. Admin portal
// shows these requests; admin decides how/when to follow up with the customer.
// EnquiryID is unique — a seller can submit only once per deal.
type NotifyCustomerRequest struct {
	ID           string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	EnquiryID    string    `json:"enquiry_id" gorm:"type:varchar(100);not null;uniqueIndex"`
	ProductID    string    `json:"product_id" gorm:"type:varchar(100);not null"`
	ProductName  string    `json:"product_name" gorm:"type:varchar(255)"`
	ImageURL     string    `json:"image_url" gorm:"type:text"`
	DealAmount   float64   `json:"deal_amount" gorm:"type:decimal(12,2)"`
	CustomerID   string    `json:"customer_id" gorm:"type:varchar(100)"`
	CustomerName string    `json:"customer_name" gorm:"type:varchar(255)"`
	SellerID     string    `json:"seller_id" gorm:"type:varchar(100);not null"`
	ShopID       string    `json:"shop_id" gorm:"type:varchar(100)"`
	Status       string    `json:"status" gorm:"type:varchar(50);not null"` // 'accepted' or 'rejected'
	AdminNote    string    `json:"admin_note" gorm:"type:text"`
	IsActioned   bool      `json:"is_actioned" gorm:"not null;default:false"`
	CreatedAt    time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
