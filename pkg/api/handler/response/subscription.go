package response

type SubscriptionPrefill struct {
	Email   string `json:"email"`
	Contact string `json:"contact"`
}

type SubscriptionOrderResponse struct {
	OrderID     string `json:"order_id"`
	KeyID       string `json:"key_id"`
	Amount      uint   `json:"amount"`
	Currency    string `json:"currency"`
	ShopOrderID string `json:"shop_order_id"`
	// GSTRateBasisPoints is the GST rate applied to this order (1800 = 18.00%),
	// and GSTAmount is the GST portion already included in Amount (minor units).
	GSTRateBasisPoints int                 `json:"gst_rate_basis_points"`
	GSTAmount          uint                `json:"gst_amount"`
	Prefill            SubscriptionPrefill `json:"prefill"`
}

type SubscriptionVerificationResponse struct {
	Status      string `json:"status"`
	Plan        string `json:"plan"`
	ActivatedAt string `json:"activated_at,omitempty"`
}

type SubscriptionStatusResponse struct {
	Status        string `json:"status"`
	PlanName      string `json:"plan_name,omitempty"`
	StartDate     string `json:"start_date,omitempty"`
	EndDate       string `json:"end_date,omitempty"`
	DaysRemaining int    `json:"days_remaining"`
	IsTrial       bool   `json:"is_trial"`
}

type SubscriptionPlanResponse struct {
	ID           string   `json:"id"`
	Name         string   `json:"name"`
	PriceMonthly uint     `json:"price_monthly"`
	DurationDays uint     `json:"duration_days"`
	Description  string   `json:"description,omitempty"`
	Features     []string `json:"features,omitempty"`
}

// BillingPaymentResponse is one paid subscription payment in the billing
// history. Amounts are in minor units (paise). GST is already included in
// AmountMinor (prices are tax-inclusive); GSTRateBasisPoints and GSTAmount are
// the values snapshotted on the order when it was created (1800 = 18.00%).
type BillingPaymentResponse struct {
	ID                 string  `json:"id"`
	PlanName           string  `json:"plan_name"`
	AmountMinor        int64   `json:"amount_minor"`
	Currency           string  `json:"currency"`
	BaseAmountMinor    int64   `json:"base_amount_minor"`
	GSTAmount          int64   `json:"gst_amount"`
	GSTRateBasisPoints int     `json:"gst_rate_basis_points"`
	GSTInclusive       bool    `json:"gst_inclusive"`
	Status             string  `json:"status"`
	RazorpayPaymentID  *string `json:"razorpay_payment_id"`
	RazorpayOrderID    string  `json:"razorpay_order_id"`
	PaidAt             string  `json:"paid_at"`
}

// BillingHistoryResponse is the payload for GET /subscriptions/billing-history.
// Each payment carries its own GSTRateBasisPoints (rates are snapshotted per
// order), so there is no single top-level rate.
type BillingHistoryResponse struct {
	GSTInclusive bool                     `json:"gst_inclusive"`
	LastPayment  *BillingPaymentResponse  `json:"last_payment"`
	Payments     []BillingPaymentResponse `json:"payments"`
}
