package response

// ShopUpdateCard is one shop's "What's New" entry as the customer app expects
// it. image_url may be a bare CDN key or empty; distance_km is populated only
// when the request supplied coordinates.
type ShopUpdateCard struct {
	ShopID      string                  `json:"shop_id"`
	ShopName    string                  `json:"shop_name"`
	ImageURL    string                  `json:"image_url,omitempty"`
	DistanceKM  *float64                `json:"distance_km,omitempty"`
	ActionLabel string                  `json:"action_label"`
	Products    []ShopUpdateProductCard `json:"products"`
}

// ShopUpdateProductCard is one showcased product. image_url may be null (the
// app renders a text fallback); attribute is an optional display string.
type ShopUpdateProductCard struct {
	ID        string  `json:"id"`
	Title     string  `json:"title"`
	ImageURL  *string `json:"image_url"`
	Attribute *string `json:"attribute,omitempty"`
}
