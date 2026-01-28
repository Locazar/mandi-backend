package domain

import "time"

type User struct {
	ID          uint      `json:"id" gorm:"primaryKey;unique"`
	Age         uint      `json:"age"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone" gorm:"unique"`
	Password    string    `json:"password"`
	Verified    bool      `json:"verified" gorm:"default:false"`
	BlockStatus bool      `json:"block_status" gorm:"default:false"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// many to many join
type UserAddress struct {
	ID        uint `json:"id" gorm:"primaryKey;unique"`
	UserID    uint `json:"user_id" gorm:"not null"`
	AddressID uint `json:"address_id" gorm:"not null"`
	IsDefault bool `json:"is_default"`
}

type Address struct {
	ID           uint    `json:"id" gorm:"primaryKey;unique"`
	UserID       uint    `json:"user_id" gorm:"not null"`
	Name         string  `json:"name" gorm:"not null"`
	PhoneNumber  string  `json:"phone_number" gorm:"not null"`
	House        *string `json:"house" gorm:"column:house"`
	AddressLine1 string  `json:"address_line1" gorm:"not null" binding:"required"`
	AddressLine2 string  `json:"address_line2" gorm:"not null" binding:"omitempty"`
	Area         string  `json:"area" gorm:"not null"`
	LandMark     string  `json:"land_mark" gorm:"not null" binding:"omitempty"`
	City         string  `json:"city" gorm:"not null"`
	Pincode      uint    `json:"pincode" gorm:"not null" binding:"required,numeric,min=6,max=6"`
	CountryID    uint    `json:"country_id" gorm:"not null" binding:"required,numeric"`
	Country      Country
	Latitude     float64   `json:"latitude" gorm:"type:decimal(10,7);"`
	Longitude    float64   `json:"longitude" gorm:"type:decimal(10,7);"`
	CreatedAt    time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Country struct {
	ID          uint   `json:"id" gorm:"primaryKey;unique;"`
	CountryName string `json:"country_name" gorm:"unique;not null"`
	ISOCode     string `json:"iso_code" gorm:"unique;not null"`
}

// Wish List
type WishList struct {
	ID            uint `json:"id" gorm:"primaryKey;not null"`
	UserID        uint `json:"user_id" gorm:"not null"`
	ShopID        uint `json:"shop_id" gorm:"not null"`
	AdminID       uint `json:"admin_id" gorm:"not null"`
	User          User
	ProductItemID uint `json:"product_item_id" gorm:"not null"`
	ProductItem   ProductItem
}

type Cart struct {
	ID              uint `json:"id" gorm:"primaryKey;not null"`
	UserID          uint `json:"user_id" gorm:"not null"`
	TotalPrice      uint `json:"total_price" gorm:"not null"`
	AppliedCouponID uint `json:"applied_coupon_id"`
	DiscountAmount  uint `json:"discount_amount"`
}

type CartItem struct {
	ID            uint `json:"id" gorm:"primaryKey;not null"`
	CartID        uint `json:"cart_id"`
	Cart          Cart
	ProductItemID uint        `json:"product_item_id" gorm:"not null"`
	ProductItem   ProductItem `json:"-"`
	Qty           uint        `json:"qty" gorm:"not null"`
}

// wallet start
// for ENUM Data type

type Wallet struct {
	ID          uint `json:"wallet_id" gorm:"primaryKey;not null"`
	UserID      uint `json:"user_id" gorm:"not null"`
	User        User `json:"-"`
	TotalAmount uint `json:"total_amount" gorm:"not null"`
}

type TransactionType string

const (
	Debit  TransactionType = "DEBIT"
	Credit TransactionType = "CREDIT"
)

type Transaction struct {
	TransactionID   uint            `json:"transaction_id" gorm:"primaryKey;not null"`
	WalletID        uint            `json:"wallet_id" gorm:"not null"`
	Wallet          Wallet          `json:"-"`
	TransactionDate time.Time       `json:"transaction_time" gorm:"not null"`
	Amount          uint            `json:"amount" gorm:"not null"`
	TransactionType TransactionType `json:"transaction_type" gorm:"not null"`
}

// wallet end

// UserOfferHistory tracks user interactions with offers for frequency capping and analytics
type UserOfferHistory struct {
	ID            uint      `json:"id" gorm:"primaryKey;not null"`
	UserID        uint      `json:"user_id" gorm:"not null"`
	OfferID       uint      `json:"offer_id" gorm:"not null"`
	EventType     string    `json:"event_type" gorm:"not null"` // 'shown', 'clicked', 'dismissed', 'applied'
	ExperimentVariant string `json:"experiment_variant" gorm:"not null"` // 'A', 'B', etc.
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}
