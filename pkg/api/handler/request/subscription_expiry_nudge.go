package request

// UpdateSubscriptionExpiryNudgeSettings configures the automated sweep.
type UpdateSubscriptionExpiryNudgeSettings struct {
	Enabled          bool `json:"enabled"`
	DaysBeforeExpiry int  `json:"days_before_expiry"`
}

// UpdateSubscriptionExpiryNudgeTemplate edits the push notification content.
type UpdateSubscriptionExpiryNudgeTemplate struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	ImageURL string `json:"image_url"`
	Route    string `json:"route"`
}
