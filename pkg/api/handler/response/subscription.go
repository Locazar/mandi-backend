package response

type SubscriptionPrefill struct {
	Email   string `json:"email"`
	Contact string `json:"contact"`
}

type SubscriptionOrderResponse struct {
	OrderID     string              `json:"order_id"`
	KeyID       string              `json:"key_id"`
	Amount      uint                `json:"amount"`
	Currency    string              `json:"currency"`
	ShopOrderID uint                `json:"shop_order_id"`
	Prefill     SubscriptionPrefill `json:"prefill"`
}

type SubscriptionVerificationResponse struct {
	Status      string `json:"status"`
	Plan        string `json:"plan"`
	ActivatedAt string `json:"activated_at,omitempty"`
}
