package request

// UpdateOnboardingNudgeTemplate edits one of the four fixed template slots.
// Key selects which one (from the URL path); it is never itself editable.
type UpdateOnboardingNudgeTemplate struct {
	Title    string `json:"title" binding:"required"`
	Body     string `json:"body" binding:"required"`
	ImageURL string `json:"image_url"`
	// Route is optional — an empty value means the tap just opens the app.
	Route string `json:"route"`
}

// UpdateOnboardingNudgeSettings edits the single schedule-config row.
type UpdateOnboardingNudgeSettings struct {
	Enabled      bool `json:"enabled"`
	GapHours     int  `json:"gap_hours" binding:"required,min=1"`
	DurationDays int  `json:"duration_days" binding:"required,min=1"`
}
