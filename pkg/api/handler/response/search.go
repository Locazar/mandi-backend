package response

import "time"

// GlobalSearchResult contains grouped search results across all entity types.
type GlobalSearchResult struct {
	Products    []ProductSearchItem    `json:"products"`
	Categories  []CategorySearchItem   `json:"categories"`
	Shops       []ShopSearchItem       `json:"shops"`
	Brands      []BrandSearchItem      `json:"brands"`
	Departments []DepartmentSearchItem `json:"departments"`
	Offers      []OfferSearchItem      `json:"offers"`
}

type ProductSearchItem struct {
	ID         string     `json:"id"`
	Name          string   `json:"name"`
	Description   string   `json:"description,omitempty"`
	CategoryID  string     `json:"category_id"`
	CategoryName  string   `json:"category_name"`
	ImageURL      string   `json:"image_url,omitempty"`
	ShopID        uint     `json:"shop_id"`
	ShopName      string   `json:"shop_name,omitempty"`
	DiscountRate  *uint    `json:"discount_rate,omitempty"`
	DynamicFields string   `json:"dynamic_fields,omitempty"`
	Images        []string `json:"images,omitempty"`
}

type CategorySearchItem struct {
	ID         string   `json:"id"`
	Name           string `json:"name"`
	DepartmentID string   `json:"department_id"`
	DepartmentName string `json:"department_name,omitempty"`
	ImageURL       string `json:"image_url,omitempty"`
}

type ShopSearchItem struct {
	ID         string    `json:"id"`
	ShopName     string  `json:"shop_name"`
	City         string  `json:"city,omitempty"`
	State        string  `json:"state,omitempty"`
	Pincode      string  `json:"pincode,omitempty"`
	ShopImageURL string  `json:"shop_image_url,omitempty"`
	Latitude     float64 `json:"latitude,omitempty"`
	Longitude    float64 `json:"longitude,omitempty"`
	ShopType     string  `json:"shop_type,omitempty"`
}

type BrandSearchItem struct {
	ID         string   `json:"id"`
	Name string `json:"name"`
}

type DepartmentSearchItem struct {
	ID         string   `json:"id"`
	Name     string `json:"name"`
	ImageURL string `json:"image_url,omitempty"`
}

type OfferSearchItem struct {
	ID         string      `json:"id"`
	Name         string    `json:"name"`
	Description  string    `json:"description,omitempty"`
	DiscountRate uint      `json:"discount_rate"`
	OfferType    string    `json:"offer_type,omitempty"`
	StartDate    time.Time `json:"start_date"`
	EndDate      time.Time `json:"end_date"`
	Image        string    `json:"image_url,omitempty"`
}

// AutocompleteSuggestion represents a single autocomplete suggestion with its source entity type.
type AutocompleteSuggestion struct {
	Text       string `json:"text"`
	EntityType string `json:"entity_type"`
	EntityID   uint   `json:"entity_id"`
	ImageURL   string `json:"image_url,omitempty"`
}

// --- Taxonomy search --------------------------------------------------------

// TaxonomySearchItem is one match in the department → category → sub-category
// tree. Every item carries its full ancestor path so a client can jump
// straight to the right place in a drill-down UI without extra lookups.
//
// For Type "category" the Category* fields are empty — the item *is* the
// category, and its parent is the department.
type TaxonomySearchItem struct {
	Type           string `json:"type"` // "category" | "subcategory"
	ID             string `json:"id"`
	Name           string `json:"name"`
	ImageURL       string `json:"image_url,omitempty"`
	DepartmentID   string `json:"department_id"`
	DepartmentName string `json:"department_name,omitempty"`
	CategoryID     string `json:"category_id,omitempty"`
	CategoryName   string `json:"category_name,omitempty"`
}

// TaxonomySearchResult is a single flat, ranked list rather than per-type
// buckets: the seller is looking for one thing and does not care whether it
// happens to be modelled as a category or a sub-category.
type TaxonomySearchResult struct {
	Results []TaxonomySearchItem `json:"results"`
}
