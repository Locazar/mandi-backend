package request

// BannerFilterRequest represents the request for filtering banners by location and category
type BannerFilterRequest struct {
	DepartmentID uint    `json:"department_id"`
	CategoryID   uint    `json:"category_id"`
	AppType      string  `json:"app_type"`
	Latitude     float64 `json:"lat"`
	Longitude    float64 `json:"lng"`
	Radius       float64 `json:"radius"`
	Pincode      string  `json:"pincode"` // optional: for pincode-based matching
}
