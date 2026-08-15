package response

// PublicShop is the SEO-safe subset of a shop exposed on public directory pages
// (e.g. locazar.com/{city}/{category}). It deliberately excludes owner PII
// (owner name, email), full address lines, and all financial/KYC fields. Phone
// is included only when the seller opted in to public visibility
// (phone_visible_consent).
type PublicShop struct {
	ID          string  `json:"id"`
	ShopName    string  `json:"shop_name"`
	City        string  `json:"city"`
	State       string  `json:"state"`
	Pincode     string  `json:"pincode"`
	Description string  `json:"description,omitempty"`
	ImageURL    string  `json:"image_url,omitempty"`
	Latitude    float64 `json:"latitude,omitempty"`
	Longitude   float64 `json:"longitude,omitempty"`
	Phone       string  `json:"phone,omitempty"`
}
