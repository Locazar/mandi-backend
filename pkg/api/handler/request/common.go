package request

import (
	"time"
)

type OTPLogin struct {
	Email string `json:"email" binding:"omitempty"`
	Phone string `json:"phone" binding:"omitempty"`
}

type OTPVerify struct {
	Otp   string `json:"otp" binding:"required,min=4,max=8"`
	OtpID string `json:"otp_id" `
}

type BlockUser struct {
	UserID uint `json:"user_id" binding:"required,numeric"`
	Block  bool `json:"block"`
}

type SalesReport struct {
	StartDate  time.Time  `json:"start_date"`
	EndDate    time.Time  `json:"end_date"`
	Pagination Pagination `json:"pagination"`
}

// stock
type UpdateStock struct {
	SKU      string `json:"sku"`
	QtyToAdd uint   `json:"qty_to_add"`
}
