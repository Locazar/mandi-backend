package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
)

// DealRepository provides read-only customer-facing deal queries.
// It unifies the legacy `offers` system with the new `promotions` system.
type DealRepository interface {
	// GetAllDeals returns a paginated, filtered list of active deals.
	GetAllDeals(ctx context.Context, filter request.DealFilter) ([]response.CustomerDeal, int64, error)

	// GetDealFilters returns aggregated filter metadata for the /deals/filters endpoint.
	GetDealFilters(ctx context.Context) (response.DealFilters, error)

	// GetDealByID returns a single deal by its ID.
	// The id may be a promotion id ("prm_...") or a legacy offer id ("ofr_...").
	GetDealByID(ctx context.Context, dealID string) (response.CustomerDeal, error)

	// GetProductsByDeal returns all products carrying the given deal.
	GetProductsByDeal(ctx context.Context, dealID string, pagination request.Pagination) ([]response.CustomerDealProduct, int64, error)

	// GetProductsWithDeals returns product items that have at least one active deal,
	// optionally filtered by deal type and discount range.
	GetProductsWithDeals(ctx context.Context, filter request.DealFilter, pagination request.Pagination) ([]response.CustomerDealProduct, int64, error)
}
