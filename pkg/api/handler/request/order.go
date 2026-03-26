package request

import (
	"time"
)

type UpdateOrder struct {
	ShopOrderID   uint `json:"shop_order_id" binding:"required"`
	OrderStatusID uint `json:"order_status_id"`
}

// return request
type Return struct {
	ShopOrderID  uint   `json:"shop_order_id" binding:"required"`
	ReturnReason string `json:"return_reason" binding:"required,min=6,max=150"`
}

type UpdateOrderReturn struct {
	OrderReturnID uint      `json:"order_return_id" binding:"required"`
	OrderStatusID uint      `json:"order_status_id" binding:"required"`
	ReturnDate    time.Time `json:"return_date" binding:"omitempty"`
	AdminComment  string    `json:"admin_comment" binding:"required,min=6,max=150"`
}

type OrderPayment struct {
	ShopOrderID     uint `json:"shop_order_id" binding:"required" `
	PaymentMethodID uint `json:"payment_method_id"  binding:"required"`
}

// type RazorpayVeification struct {
// 	RazorpayPaymentID string `json:"razorpay_payment_id"`
// 	RazorpayOrderID   string `json:"razorpay_order_id"`
// 	RazorpaySignature string `json:"razorpay_signature"`
// 	UserID            uint   `json:"user_id"`
// 	ShopOrderID       uint   `json:"shop_order_id"`
// }

type ShoppingFeedback struct {
	ShopOrderID    uint   `json:"shop_order_id" binding:"required"`
	Rating         uint   `json:"rating" binding:"required,min=1,max=5"`
	Comment        string `json:"comment" binding:"omitempty,max=255"`
	ShopID         uint   `json:"shop_id" `
	Amount         int    `json:"amount" binding:"required"`
	AdminID        uint   `json:"admin_id" `
	SellerVerified bool   `json:"verified" default:"false"`
}
