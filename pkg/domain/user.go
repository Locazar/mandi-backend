package domain

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	ID            uint           `json:"id" gorm:"primaryKey;unique"`
	DateOfBirth   time.Time      `json:"date_of_birth" gorm:"type:date"`
	FirstName     string         `json:"first_name"`
	LastName      string         `json:"last_name"`
	Email         string         `json:"email"`
	Phone         string         `json:"phone" gorm:"unique"`
	Password      string         `json:"-"`
	EmailVerified bool           `json:"email_verified" gorm:"default:false"`
	PhoneVerified bool           `json:"phone_verified" gorm:"default:false"`
	BlockStatus   bool           `json:"block_status" gorm:"default:false"`
	TrialUsed     bool           `json:"trial_used" gorm:"default:false"`
	CreatedAt     time.Time      `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     gorm.DeletedAt `json:"-" gorm:"index"`
}

// AgeYears derives the user's whole-year age as of `now`.
func (u User) AgeYears(now time.Time) int {
	if u.DateOfBirth.IsZero() {
		return 0
	}
	y := now.Year() - u.DateOfBirth.Year()
	if now.YearDay() < u.DateOfBirth.YearDay() {
		y--
	}
	return y
}

// many to many join
type UserAddress struct {
	ID        uint      `json:"id" gorm:"primaryKey;unique"`
	UserID    uint      `json:"user_id" gorm:"not null"`
	AddressID uint      `json:"address_id" gorm:"not null"`
	IsDefault bool      `json:"is_default"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type Address struct {
	ID           uint           `json:"id" gorm:"primaryKey;unique"`
	UserID       uint           `json:"user_id" gorm:"not null"`
	LandMark     string         `json:"land_mark" gorm:"not null"`
	Area         *string        `json:"area" gorm:"type:varchar(255);default:null"`
	City         string         `json:"city" gorm:"not null"`
	Pincode      string         `json:"pincode" gorm:"size:10"`
	CountryID    uint           `json:"country_id" gorm:"not null"`
	Country      Country        `json:"country"`
	Latitude     *float64       `json:"latitude" gorm:"type:decimal(10,7);"`
	Longitude    *float64       `json:"longitude" gorm:"type:decimal(10,7);"`
	PhoneNumber  string         `json:"phone_number" gorm:"not null"`
	AddressType  AddressType    `json:"address_type" gorm:"not null"`
	AddressLine1 string         `json:"address_line1" gorm:"not null"`
	AddressLine2 string         `json:"address_line2" gorm:"not null"`
	IsDefault    *bool          `json:"is_default" gorm:"type:boolean"`
	UpdatedAt    time.Time      `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt    gorm.DeletedAt `json:"-" gorm:"index"`
}

type Country struct {
	ID          uint      `json:"id" gorm:"primaryKey;unique;"`
	CountryName string    `json:"country_name" gorm:"unique;not null"`
	ISOCode     string    `json:"iso_code" gorm:"unique;not null"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
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
	CreatedAt     time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type Cart struct {
	ID              uint      `json:"id" gorm:"primaryKey;not null"`
	UserID          uint      `json:"user_id" gorm:"not null"`
	TotalPrice      Money     `json:"total_price" gorm:"embedded;embeddedPrefix:total_price_"`
	AppliedCouponID uint      `json:"applied_coupon_id"`
	DiscountAmount  Money     `json:"discount_amount" gorm:"embedded;embeddedPrefix:discount_amount_"`
	CreatedAt       time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type CartItem struct {
	ID            uint `json:"id" gorm:"primaryKey;not null"`
	CartID        uint `json:"cart_id"`
	Cart          Cart
	ProductItemID uint        `json:"product_item_id" gorm:"not null"`
	ProductItem   ProductItem `json:"-"`
	Qty           uint        `json:"qty" gorm:"not null"`
	CreatedAt     time.Time   `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time   `json:"updated_at" gorm:"autoUpdateTime"`
}

// wallet start
// for ENUM Data type

type Wallet struct {
	ID          uint      `json:"wallet_id" gorm:"primaryKey;not null"`
	UserID      uint      `json:"user_id" gorm:"not null"`
	User        User      `json:"-"`
	TotalAmount Money     `json:"total_amount" gorm:"embedded;embeddedPrefix:total_amount_"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

type TransactionType string

const (
	Debit  TransactionType = "DEBIT"
	Credit TransactionType = "CREDIT"
)

func (t TransactionType) IsValid() bool {
	switch t {
	case Debit, Credit:
		return true
	}
	return false
}

type Transaction struct {
	TransactionID   uint            `json:"transaction_id" gorm:"primaryKey;not null"`
	WalletID        uint            `json:"wallet_id" gorm:"not null"`
	Wallet          Wallet          `json:"-"`
	TransactionDate time.Time       `json:"transaction_time" gorm:"not null"`
	Amount          Money           `json:"amount" gorm:"embedded;embeddedPrefix:amount_"`
	TransactionType TransactionType `json:"transaction_type" gorm:"not null"`
}

// wallet end
