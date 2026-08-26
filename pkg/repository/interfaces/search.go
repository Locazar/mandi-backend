package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
)

type SearchRepository interface {
	GlobalSearch(ctx context.Context, query string, entityTypes []string,
		lat, lng, radiusKm float64, pincode string, limit, offset uint) (response.GlobalSearchResult, error)
	AutocompleteSuggestions(ctx context.Context, prefix string, limit uint) ([]response.AutocompleteSuggestion, error)
	SearchTaxonomy(ctx context.Context, query, departmentID string, limit uint) ([]response.TaxonomySearchItem, error)
}
