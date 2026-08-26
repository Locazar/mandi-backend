package request

// GlobalSearchRequest holds query parameters for the global search endpoint.
type GlobalSearchRequest struct {
	Query     string  `form:"q" binding:"required,min=1"`
	Types     string  `form:"types"`   // comma-separated entity types filter, e.g. "products,shops,categories"
	Latitude  float64 `form:"lat"`     // optional, for geo-boosting shops
	Longitude float64 `form:"lng"`     // optional
	RadiusKm  float64 `form:"radius"`  // optional, km
	Pincode   string  `form:"pincode"` // optional, alternative to lat/lng
	Limit     uint    `form:"limit"`   // per-entity limit, default 5
	Offset    uint    `form:"offset"`  // offset for pagination, default 0
}

// AutocompleteRequest holds query parameters for the autocomplete endpoint.
type AutocompleteRequest struct {
	Query string `form:"q" binding:"required,min=1"`
	Limit uint   `form:"limit"` // default 10, max 20
}

// TaxonomySearchRequest holds query parameters for the taxonomy search
// endpoint, which searches categories and sub-categories together.
type TaxonomySearchRequest struct {
	Query        string `form:"q" binding:"required,min=1"`
	DepartmentID string `form:"department_id"` // optional, restrict to one department
	Limit        uint   `form:"limit"`         // default 20, max 50
}
