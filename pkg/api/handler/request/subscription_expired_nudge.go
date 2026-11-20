package request

// UpdateSubscriptionExpiredNudgeSettings configures the automated sweep for already-expired sellers.
type UpdateSubscriptionExpiredNudgeSettings struct {
	Enabled       bool `json:"enabled"`
	FireHourIST   int  `json:"fire_hour_ist"`
	FireMinuteIST int  `json:"fire_minute_ist"`
}

// UpdateSubscriptionExpiredNudgeTemplate edits the push notification content.
type UpdateSubscriptionExpiredNudgeTemplate struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	ImageURL string `json:"image_url"`
	Route    string `json:"route"`
}
