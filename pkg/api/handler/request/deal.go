package request

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// DealFilter holds all query-param filters for GET /api/deals/.
type DealFilter struct {
	// Offer type: "percentage", "fixed", "bogo", "bundle", "flash",
	// "free_gift", "free_shipping", "loyalty", or "" for all.
	DealType string `form:"deal_type"`

	// Discount band
	MinDiscount float64 `form:"min_discount"` // e.g. 10  → ≥10% / ≥₹10
	MaxDiscount float64 `form:"max_discount"` // e.g. 50  → ≤50%; 0 = no cap

	// Scope filters
	CategoryID   string `form:"category_id"`
	DepartmentID string `form:"department_id"`
	ShopID       string `form:"shop_id"`

	// Location (optional — shows deals near the customer)
	Lat      float64 `form:"lat"`
	Lng      float64 `form:"lng"`
	RadiusKM float64 `form:"radius_km"`

	// Source toggle
	Source string `form:"source"` // "promotion" | "offer" | "" (both)

	// Sort
	SortBy string `form:"sort_by"` // "discount_desc" | "discount_asc" | "ending_soon" | "newest"

	Pagination
}

// ProductDealFilter extends the existing product-listing filter with deal params.
type ProductDealFilter struct {
	HasDeal     bool    `form:"has_deal"`      // true → only products with an active deal
	DealType    string  `form:"deal_type"`
	MinDiscount float64 `form:"min_discount"`
	MaxDiscount float64 `form:"max_discount"`
}

// ParseDealFilter builds a DealFilter from gin query params.
func ParseDealFilter(ctx *gin.Context) DealFilter {
	var f DealFilter
	f.DealType = ctx.DefaultQuery("deal_type", "")
	f.Source = ctx.DefaultQuery("source", "")
	f.SortBy = ctx.DefaultQuery("sort_by", "discount_desc")
	f.CategoryID = ctx.DefaultQuery("category_id", "")
	f.DepartmentID = ctx.DefaultQuery("department_id", "")
	f.ShopID = ctx.DefaultQuery("shop_id", "")
	f.MinDiscount, _ = strconv.ParseFloat(ctx.DefaultQuery("min_discount", "0"), 64)
	f.MaxDiscount, _ = strconv.ParseFloat(ctx.DefaultQuery("max_discount", "0"), 64)
	f.Lat, _ = strconv.ParseFloat(ctx.DefaultQuery("lat", "0"), 64)
	f.Lng, _ = strconv.ParseFloat(ctx.DefaultQuery("lng", "0"), 64)
	f.RadiusKM, _ = strconv.ParseFloat(ctx.DefaultQuery("radius_km", "10"), 64)
	f.Pagination = GetPagination(ctx)
	return f
}
