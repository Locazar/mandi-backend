package usecase

import (
	"context"
	"strings"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	repoInterface "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	ucinterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

const (
	defaultSearchLimit       uint = 5
	maxSearchLimit           uint = 50
	defaultAutocompleteLimit uint = 10
	maxAutocompleteLimit     uint = 20
)

type searchUseCase struct {
	searchRepo repoInterface.SearchRepository
}

func NewSearchUseCase(searchRepo repoInterface.SearchRepository) ucinterfaces.SearchUseCase {
	return &searchUseCase{searchRepo: searchRepo}
}

func (s *searchUseCase) GlobalSearch(ctx context.Context, req request.GlobalSearchRequest) (response.GlobalSearchResult, error) {
	query := strings.TrimSpace(req.Query)
	if query == "" {
		return response.GlobalSearchResult{
			Products:    []response.ProductSearchItem{},
			Categories:  []response.CategorySearchItem{},
			Shops:       []response.ShopSearchItem{},
			Brands:      []response.BrandSearchItem{},
			Departments: []response.DepartmentSearchItem{},
			Offers:      []response.OfferSearchItem{},
		}, nil
	}

	limit := req.Limit
	if limit == 0 {
		limit = defaultSearchLimit
	}
	if limit > maxSearchLimit {
		limit = maxSearchLimit
	}

	var entityTypes []string
	if req.Types != "" {
		for _, t := range strings.Split(req.Types, ",") {
			t = strings.TrimSpace(strings.ToLower(t))
			if t != "" {
				entityTypes = append(entityTypes, t)
			}
		}
	}

	return s.searchRepo.GlobalSearch(ctx, query, entityTypes,
		req.Latitude, req.Longitude, req.RadiusKm, req.Pincode, limit, req.Offset)
}

func (s *searchUseCase) Autocomplete(ctx context.Context, req request.AutocompleteRequest) ([]response.AutocompleteSuggestion, error) {
	prefix := strings.TrimSpace(req.Query)
	if prefix == "" {
		return []response.AutocompleteSuggestion{}, nil
	}

	limit := req.Limit
	if limit == 0 {
		limit = defaultAutocompleteLimit
	}
	if limit > maxAutocompleteLimit {
		limit = maxAutocompleteLimit
	}

	return s.searchRepo.AutocompleteSuggestions(ctx, prefix, limit)
}
