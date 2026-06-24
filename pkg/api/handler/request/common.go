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

// SignupOtpSend is the body for the seller signup/login OTP-send endpoint.
// It accepts the phone number under either "mobile" (signup) or "phone"
// (login/resend); full_name and password are only used when creating a new seller.
type SignupOtpSend struct {
	Mobile   string `json:"mobile"`
	Phone    string `json:"phone"`
	FullName string `json:"full_name"`
	Password string `json:"password"`
}

// PhoneNumber returns the phone number regardless of which field the client sent.
func (s SignupOtpSend) PhoneNumber() string {
	if s.Mobile != "" {
		return s.Mobile
	}
	return s.Phone
}

type BlockUser struct {
	UserID string `json:"user_id" binding:"required"`
	Block  bool   `json:"block"`
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
