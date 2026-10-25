package response

import "time"

// CustomerDeal is the unified deal object exposed to the customer app.
// It merges the legacy `offers` system (admin-created, category/product-linked)
// and the new `promotions` system (seller-created, product-level).
type CustomerDeal struct {
	ID          string  `json:"id"`
	Source      string  `json:"source"`       // "promotion" | "offer"
	DealType    string  `json:"deal_type"`    // "percentage" | "fixed" | "bogo" | "bundle" | "flash" | "free_gift" | "free_shipping" | "loyalty"
	Title       string  `json:"title"`        // offer_name / offer name
	Description string  `json:"description"`
	DiscountRate float64 `json:"discount_rate"` // 0 for non-numeric types (bogo, gift, shipping)
	DiscountLabel string `json:"discount_label"` // e.g. "30% OFF" | "₹100 OFF" | "BOGO" | "Free Shipping"

	StartDate time.Time `json:"start_date"`
	EndDate   time.Time `json:"end_date"`
	IsExpired bool      `json:"is_expired"`
	DaysLeft  int       `json:"days_left"` // negative once expired

	ShopID   string `json:"shop_id,omitempty"`
	ShopName string `json:"shop_name,omitempty"`
	ShopLogo string `json:"shop_logo,omitempty"`

	IconPath  string `json:"icon_path,omitempty"`
	ImageURL  string `json:"image_url,omitempty"`

	ProductCount int `json:"product_count"` // how many products carry this deal

	// Promotion-specific extras (nil for legacy offers)
	MinimumPurchaseAmount  *float64 `json:"minimum_purchase_amount,omitempty"`
	BogoBuyQuantity        *int     `json:"bogo_buy_quantity,omitempty"`
	BogoGetQuantity        *int     `json:"bogo_get_quantity,omitempty"`
	BogoCombinationEnabled *bool    `json:"bogo_combination_enabled,omitempty"`
	GiftDescription        *string  `json:"gift_description,omitempty"`
}

// CustomerDealProduct is a product item decorated with its active deal.
type CustomerDealProduct struct {
	ProductItemID   string   `json:"product_item_id"`
	ProductName     string   `json:"product_name"`
	SubCategoryName string   `json:"sub_category_name"`
	ShopID          string   `json:"shop_id"`
	ShopName        string   `json:"shop_name"`
	ShopLogo        string   `json:"shop_logo"`
	Images          []string `json:"images"`
	InStock         bool     `json:"in_stock"`

	DealID       string  `json:"deal_id"`
	DealType     string  `json:"deal_type"`
	DiscountRate float64 `json:"discount_rate"`
	DiscountLabel string `json:"discount_label"`
	DealEndsAt   time.Time `json:"deal_ends_at"`
}

// DealFilters describes what filter options are available on the /deals/ listing.
// The customer app uses this to build the filter chip bar.
type DealFilters struct {
	DealTypes      []DealTypeFilter      `json:"deal_types"`
	DiscountRanges []DiscountRangeFilter  `json:"discount_ranges"`
	Categories     []DealCategoryFilter   `json:"categories"`
	TotalDeals     int                    `json:"total_deals"`
}

type DealTypeFilter struct {
	Type      string `json:"type"`
	Label     string `json:"label"`
	IconPath  string `json:"icon_path,omitempty"`
	Count     int    `json:"count"`
}

type DiscountRangeFilter struct {
	Label    string  `json:"label"`    // "Under 10%", "10–30%", "30–50%", "50%+"
	Min      float64 `json:"min"`
	Max      float64 `json:"max"`      // 0 means no upper bound
	Count    int     `json:"count"`
}

type DealCategoryFilter struct {
	CategoryID   string `json:"category_id"`
	CategoryName string `json:"category_name"`
	Count        int    `json:"count"`
}
