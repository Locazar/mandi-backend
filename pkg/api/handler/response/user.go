package response

import "time"

// user details response
type User struct {
	ID          uint      `json:"id" copier:"must"`
	GoogleImage string    `json:"google_profile_image"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Age         uint      `json:"age"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Verified    bool      `json:"verified"`
	BlockStatus bool      `json:"block_status"`
	CreatedAt   time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type CartItem struct {
	ProductItemId uint   `json:"product_item_id"`
	ProductName   string `json:"product_name"`
	Price         uint   `json:"price"`
	DiscountPrice uint   `json:"discount_price"`
	QtyInStock    uint   `json:"qty_in_stock"`
	Qty           uint   `json:"qty"`
	SubTotal      uint   `json:"sub_total"`
}

type Cart struct {
	CartItems       []CartItem
	AppliedCouponID uint `json:"applied_coupon_id"`
	TotalPrice      uint `json:"total_price"`
	DiscountAmount  uint `json:"discount_amount"`
}

// address
type Address struct {
	ID           uint     `json:"address_id"`
	LandMark     string   `json:"land_mark"`
	City         string   `json:"city"`
	Pincode      int64    `json:"pincode"`
	CountryID    uint     `json:"country_id"`
	CountryName  string   `json:"country_name"`
	Latitude     *float64 `json:"latitude"`
	Longitude    *float64 `json:"longitude"`
	PhoneNumber  string   `json:"phone_number"`
	AddressType  string   `json:"address_type"`
	AddressLine1 string   `json:"address_line1"`
	AddressLine2 string   `json:"address_line2"`
	IsDefault    *bool    `json:"is_default"`
}

// wish list response
type WishListItem struct {
	ID            uint   `json:"wish_list_id"`
	ProductItemID uint   `json:"product_item_id"`
	Name          string `json:"product_name"`
	ProductID     uint   `json:"product_id"`
}

type Admin struct {
	ID        uint      `json:"admin_id"`
	ShopID    uint      `json:"shop_id"`
	ShopName  string    `json:"shop_name"`
	Email     string    `json:"email"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	Latitude  float64   `json:"latitude"`
	Longitude float64   `json:"longitude"`
	Verified  string    `json:"verified"`
	Status    string    `json:"status"`
}
