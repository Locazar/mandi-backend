package request

type CreateSubscriptionOrderRequest struct {
	PlanID uint `json:"plan_id" binding:"required"`
}

type VerifySubscriptionPaymentRequest struct {
	OrderID   string `json:"razorpay_order_id" binding:"required"`
	PaymentID string `json:"razorpay_payment_id" binding:"required"`
	Signature string `json:"razorpay_signature" binding:"required"`
}

type PaymentFailureRequest struct {
	OrderID string `json:"order_id"`
	Code    int    `json:"code"`
	Message string `json:"message"`
}
