package request

// UpdateDistrictNudgeTemplate edits the district-growth-nudge template.
type UpdateDistrictNudgeTemplate struct {
	Title    string `json:"title"`
	Body     string `json:"body"`
	ImageURL string `json:"image_url"`
	Route    string `json:"route"`
}

// UpdateDistrictNudgeSettings edits the district-growth-nudge schedule.
type UpdateDistrictNudgeSettings struct {
	Enabled  bool    `json:"enabled"`
	State    string  `json:"state"`
	RadiusKm float64 `json:"radius_km"`
}
