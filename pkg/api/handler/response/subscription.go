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
	ID           string `json:"id"`
	Name         string `json:"name"`
	PriceMonthly uint   `json:"price_monthly"`
	DurationDays uint   `json:"duration_days"`
}
