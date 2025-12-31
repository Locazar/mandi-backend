package response

import "time"

type Shop struct {
	ID                     uint      `json:"shop_id"`
	ShopName               string    `json:"shop_name"`
	Email                  string    `json:"email"`
	Phone                  string    `json:"phone"`
	AddressLine1           string    `json:"address_line1"`
	AddressLine2           string    `json:"address_line2"`
	City                   string    `json:"city"`
	State                  string    `json:"state"`
	Country                string    `json:"country"`
	Pincode                string    `json:"pincode"`
	ShopType               string    `json:"shop_type"`
	ShopVerificationStatus string    `json:"shop_verification_status"`
	ShopImageURL           string    `json:"shop_image_url"`
	Latitude               float64   `json:"latitude"`
	Longitude              float64   `json:"longitude"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at"`
}
