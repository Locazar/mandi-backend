package response

// PostLoginOfferResponse represents the decision for post-login offer
type PostLoginOfferResponse struct {
	ShowOffer         bool   `json:"show_offer"`
	OfferType         string `json:"offer_type"` // BOTTOM_SHEET, FULL_SCREEN, BANNER, NONE
	OfferID           string `json:"offer_id,omitempty"`
	Title             string `json:"title,omitempty"`
	Description       string `json:"description,omitempty"`
	CTA               string `json:"cta,omitempty"`
	DiscountType      string `json:"discount_type,omitempty"` // PERCENT, FIXED
	DiscountValue     int    `json:"discount_value,omitempty"`
	FrequencyCapHours int    `json:"frequency_cap_hours,omitempty"`
	ExperimentVariant string `json:"experiment_variant,omitempty"`
}

// Banner represents a banner for display
type Banner struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	ImageURL    string `json:"image_url"`
	Link        string `json:"link"`
}
