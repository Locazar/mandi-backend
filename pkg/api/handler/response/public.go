package response

// PublicShop is the SEO-safe subset of a shop exposed on public directory pages
// (e.g. locazar.com/{city}/{category}). It deliberately excludes owner PII
// (owner name, email), full address lines, all financial/KYC fields and the
// seller's phone number (customers reach sellers via the WhatsApp relay).
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
	// Phone is never serialised — seller numbers are not exposed publicly.
	Phone       string  `json:"-"`
}

// PublicProduct is the SEO-safe subset of a product for public directory pages.
// It deliberately excludes PRICE and any seller PII/internal fields — only what a
// public marketplace listing shows: the product, its category, the shop it
// belongs to and where. Used by the public /api/public/products endpoint.
type PublicProduct struct {
	ID          string `json:"id"`
	Name        string `json:"name"`     // product display name (sub_category_name)
	Category    string `json:"category"` // category name
	Department  string `json:"department"`
	Description string `json:"description,omitempty"`
	ImageURL    string `json:"image_url,omitempty"`
	ShopID      string `json:"shop_id"`
	ShopName    string `json:"shop_name"`
	City        string `json:"city"`
	State       string `json:"state"`
	InStock     bool   `json:"in_stock"`
}
