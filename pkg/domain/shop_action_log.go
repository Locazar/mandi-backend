package domain

import "time"

const (
	ShopActionCall       = "call"
	ShopActionWhatsApp   = "whatsapp"
	ShopActionDirections = "directions"
	ShopActionCopyNumber = "copy_number"
)

// ShopActionLog records a customer interaction on a shop profile page:
// directions, WhatsApp, call, or phone number copy.
type ShopActionLog struct {
	ID            string    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Action        string    `json:"action" gorm:"type:varchar(50);not null;index"`
	ShopID        string    `json:"shop_id" gorm:"type:varchar(100);not null;index"`
	ShopName      string    `json:"shop_name" gorm:"type:varchar(255)"`
	SellerID      string    `json:"seller_id" gorm:"type:varchar(100);index"`
	CustomerID    string    `json:"customer_id" gorm:"type:varchar(100);index"`
	CustomerPhone string    `json:"customer_phone" gorm:"type:varchar(20)"`
	CustomerLat   *float64  `json:"customer_lat,omitempty" gorm:"type:decimal(10,7)"`
	CustomerLon   *float64  `json:"customer_lon,omitempty" gorm:"type:decimal(10,7)"`
	LocationLabel string    `json:"location_label" gorm:"type:varchar(255)"`
	SessionID     string    `json:"session_id" gorm:"type:varchar(100);index"`
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
}
