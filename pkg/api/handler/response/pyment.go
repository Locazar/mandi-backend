package response

import "github.com/rohit221990/mandi-backend/pkg/domain"

type OrderPayment struct {
	PaymentType  domain.PaymentType `json:"payment_type"`
	PaymentOrder any                `json:"payment_order"`
}
