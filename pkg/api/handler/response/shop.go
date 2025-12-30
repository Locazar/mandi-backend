package response

type Shop struct {
	ID           uint    `json:"shop_id"`
	ShopName     string  `json:"shop_name"`
	OwnerName    string  `json:"owner_name"`
	Email        string  `json:"email"`
	Phone        string  `json:"phone"`
	ShopImageUrl string  `json:"shop_image_url"`
	AddressLine1 string  `json:"address_line1"`
	AddressLine2 string  `json:"address_line2"`
	State        string  `json:"state"`
	City         string  `json:"city"`
	Pincode      uint    `json:"pincode"`
	Country      string  `json:"country"`
	Latitude     float64 `json:"latitude"`
	Longitude    float64 `json:"longitude"`
	Verified     bool    `json:"verified"`
	CreatedAt    string  `json:"created_at"`
	UpdatedAt    string  `json:"updated_at"`
}
