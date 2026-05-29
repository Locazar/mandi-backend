package domain

import (
	"time"

	"gorm.io/gorm"
)

// for defining ENUM stasues
type OrderStatusType string

// payment types
type PaymentType string

const (
	// order status
	StatusPaymentPending  OrderStatusType = "payment pending"
	StatusOrderPlaced     OrderStatusType = "order placed"
	StatusOrderCancelled  OrderStatusType = "order cancelled"
	StatusOrderDelivered  OrderStatusType = "order delivered"
	StatusReturnRequested OrderStatusType = "return requested"
	StatusReturnApproved  OrderStatusType = "return approved"
	StatusReturnCancelled OrderStatusType = "return cancelled"
	StatusOrderReturned   OrderStatusType = "order returned"

	// payment type
	RazopayPayment        PaymentType = "razor pay"
	RazorPayMaximumAmount             = 50000 // this is only for initial admin can later change this
	CodPayment            PaymentType = "cod"
	CodMaximumAmount                  = 20000
	StripePayment         PaymentType = "stripe"
	StripeMaximumAmount               = 50000
)

func (s OrderStatusType) IsValid() bool {
	switch s {
	case StatusPaymentPending, StatusOrderPlaced, StatusOrderCancelled, StatusOrderDelivered,
		StatusReturnRequested, StatusReturnApproved, StatusReturnCancelled, StatusOrderReturned:
		return true
	}
	return false
}

func (p PaymentType) IsValid() bool {
	switch p {
	case RazopayPayment, CodPayment, StripePayment:
		return true
	}
	return false
}

type PaymentMethod struct {
	ID            uint        `json:"id" gorm:"primaryKey;not null"`
	Name          PaymentType `json:"name" gorm:"unique;not null"`
	BlockStatus   bool        `json:"block_status" gorm:"not null;default:false"`
	MaximumAmount Money       `json:"maximum_amount" gorm:"embedded;embeddedPrefix:maximum_amount_"`
}

// ShopOrder.Status replaces the former OrderStatus lookup table +
// OrderStatusID FK. The column is CHECK-constrained in the DB migration.
type ShopOrder struct {
	ID              uint            `json:"shop_order_id" gorm:"primaryKey;not null"`
	UserID          uint            `json:"user_id" gorm:"not null"`
	User            User            `json:"-"`
	OrderDate       time.Time       `json:"order_date" gorm:"not null"`
	AddressID       uint            `json:"address_id" gorm:"not null"`
	Address         Address         `json:"-"`
	OrderTotal      Money           `json:"order_total" gorm:"embedded;embeddedPrefix:order_total_"`
	Discount        Money           `json:"discount" gorm:"embedded;embeddedPrefix:discount_"`
	Status          OrderStatusType `json:"status" gorm:"type:varchar(50);not null"`
	PaymentMethodID uint            `json:"payment_method_id"`
	PaymentMethod   PaymentMethod   `json:"-"`
	ShopID          uint            `json:"shop_id"`

	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

type OrderLine struct {
	ID            uint      `json:"id" gorm:"primaryKey;not null"`
	ProductItemID uint      `json:"product_item_id" gorm:"not null"`
	ShopOrderID   uint      `json:"shop_order_id" gorm:"not null"`
	ShopOrder     ShopOrder `json:"-"`
	Qty           uint      `json:"qty" gorm:"not null"`
	Price         Money     `json:"price" gorm:"embedded;embeddedPrefix:price_"`
}

type OrderReturn struct {
	ID           uint      `json:"id" gorm:"primaryKey;not null"`
	ShopOrderID  uint      `json:"shop_order_id" gorm:"not null;unique"`
	ShopOrder    ShopOrder `json:"-"`
	RequestDate  time.Time `json:"request_date" gorm:"not null"`
	ReturnReason string    `json:"return_reason" gorm:"not null"`
	RefundAmount Money     `json:"refund_amount" gorm:"embedded;embeddedPrefix:refund_amount_"`

	IsApproved   bool      `json:"is_approved"`
	ReturnDate   time.Time `json:"return_date"`
	ApprovalDate time.Time `json:"approval_date"`
	AdminComment string    `json:"admin_comment"`
}
