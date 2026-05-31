package request

import (
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type UpdateOrder struct {
	ShopOrderID string                 `json:"shop_order_id" binding:"required"`
	OrderStatus domain.OrderStatusType `json:"order_status"`
}

// return request
type Return struct {
	ShopOrderID  string `json:"shop_order_id" binding:"required"`
	ReturnReason string `json:"return_reason" binding:"required,min=6,max=150"`
}

type UpdateOrderReturn struct {
	OrderReturnID string                 `json:"order_return_id" binding:"required"`
	OrderStatus   domain.OrderStatusType `json:"order_status" binding:"required"`
	ReturnDate    time.Time              `json:"return_date" binding:"omitempty"`
	AdminComment  string                 `json:"admin_comment" binding:"required,min=6,max=150"`
}

type OrderPayment struct {
	ShopOrderID     string `json:"shop_order_id" binding:"required"`
	PaymentMethodID string `json:"payment_method_id" binding:"required"`
}

type ShoppingFeedback struct {
	ShopOrderID    string `json:"shop_order_id" binding:"required"`
	Rating         uint   `json:"rating" binding:"required,min=1,max=5"`
	Comment        string `json:"comment" binding:"omitempty,max=255"`
	ShopID         string `json:"shop_id"`
	Amount         int    `json:"amount" binding:"required"`
	AdminID        string `json:"admin_id"`
	SellerVerified bool   `json:"verified" default:"false"`
}
