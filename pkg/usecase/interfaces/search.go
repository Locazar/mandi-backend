package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
)

type SearchUseCase interface {
	GlobalSearch(ctx context.Context, req request.GlobalSearchRequest) (response.GlobalSearchResult, error)
	Autocomplete(ctx context.Context, req request.AutocompleteRequest) ([]response.AutocompleteSuggestion, error)
	SearchTaxonomy(ctx context.Context, req request.TaxonomySearchRequest) (response.TaxonomySearchResult, error)
}
