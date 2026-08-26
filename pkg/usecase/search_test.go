package usecase

import (
	"context"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/mock/mockrepo"
	"github.com/stretchr/testify/assert"
)

func TestGlobalSearch_ValidQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepo.NewMockSearchRepository(ctrl)
	expectedResult := response.GlobalSearchResult{
		Products: []response.ProductSearchItem{
			{ID: "prd_test1", Name: "iPhone", CategoryID: "cat_test5", CategoryName: "Electronics"},
		},
		Categories:  []response.CategorySearchItem{},
		Shops:       []response.ShopSearchItem{},
		Brands:      []response.BrandSearchItem{},
		Departments: []response.DepartmentSearchItem{},
		Offers:      []response.OfferSearchItem{},
	}

	mockRepo.EXPECT().
		GlobalSearch(gomock.Any(), "iphone", gomock.Any(), float64(0), float64(0), float64(0), "", uint(5), uint(0)).
		Return(expectedResult, nil)

	useCase := NewSearchUseCase(mockRepo)
	req := request.GlobalSearchRequest{
		Query:  "iphone",
		Limit:  0, // Should default to 5
		Offset: 0,
	}

	result, err := useCase.GlobalSearch(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, len(expectedResult.Products), len(result.Products))
	assert.Equal(t, "iPhone", result.Products[0].Name)
}

func TestGlobalSearch_EmptyQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepo.NewMockSearchRepository(ctrl)
	useCase := NewSearchUseCase(mockRepo)

	req := request.GlobalSearchRequest{
		Query: "",
	}

	result, err := useCase.GlobalSearch(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result.Products))
	assert.Equal(t, 0, len(result.Categories))
}

func TestGlobalSearch_LimitCapping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepo.NewMockSearchRepository(ctrl)
	expectedResult := response.GlobalSearchResult{
		Products:    []response.ProductSearchItem{},
		Categories:  []response.CategorySearchItem{},
		Shops:       []response.ShopSearchItem{},
		Brands:      []response.BrandSearchItem{},
		Departments: []response.DepartmentSearchItem{},
		Offers:      []response.OfferSearchItem{},
	}

	// Should cap limit to maxSearchLimit (50)
	mockRepo.EXPECT().
		GlobalSearch(gomock.Any(), "iphone", gomock.Any(), float64(0), float64(0), float64(0), "", uint(50), uint(0)).
		Return(expectedResult, nil)

	useCase := NewSearchUseCase(mockRepo)
	req := request.GlobalSearchRequest{
		Query: "iphone",
		Limit: 999, // Requested limit exceeds max
	}

	_, err := useCase.GlobalSearch(context.Background(), req)

	assert.NoError(t, err)
}

func TestAutocomplete_ValidPrefix(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepo.NewMockSearchRepository(ctrl)
	expectedSuggestions := []response.AutocompleteSuggestion{
		{Text: "iPhone", EntityType: "product", EntityID: 1},
		{Text: "iPhone 14", EntityType: "product", EntityID: 2},
		{Text: "Apple", EntityType: "brand", EntityID: 10},
	}

	mockRepo.EXPECT().
		AutocompleteSuggestions(gomock.Any(), "iph", uint(10)).
		Return(expectedSuggestions, nil)

	useCase := NewSearchUseCase(mockRepo)
	req := request.AutocompleteRequest{
		Query: "iph",
		Limit: 0, // Should default to 10
	}

	result, err := useCase.Autocomplete(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, len(expectedSuggestions), len(result))
	assert.Equal(t, "iPhone", result[0].Text)
}

func TestAutocomplete_EmptyPrefix(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepo.NewMockSearchRepository(ctrl)
	useCase := NewSearchUseCase(mockRepo)

	req := request.AutocompleteRequest{
		Query: "",
	}

	result, err := useCase.Autocomplete(context.Background(), req)

	assert.NoError(t, err)
	assert.Equal(t, 0, len(result))
}

func TestAutocomplete_LimitCapping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepo.NewMockSearchRepository(ctrl)
	expectedSuggestions := []response.AutocompleteSuggestion{}

	// Should cap limit to maxAutocompleteLimit (20)
	mockRepo.EXPECT().
		AutocompleteSuggestions(gomock.Any(), "iph", uint(20)).
		Return(expectedSuggestions, nil)

	useCase := NewSearchUseCase(mockRepo)
	req := request.AutocompleteRequest{
		Query: "iph",
		Limit: 99, // Exceeds max
	}

	_, err := useCase.Autocomplete(context.Background(), req)

	assert.NoError(t, err)
}

func TestSearchTaxonomy_ValidQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepo.NewMockSearchRepository(ctrl)
	expected := []response.TaxonomySearchItem{
		{
			Type: "subcategory", ID: "scat_1", Name: "Shirts",
			DepartmentID: "dep_1", DepartmentName: "Clothing",
			CategoryID: "cat_1", CategoryName: "Menswear",
		},
	}

	mockRepo.EXPECT().
		SearchTaxonomy(gomock.Any(), "shirt", "", uint(20)).
		Return(expected, nil)

	useCase := NewSearchUseCase(mockRepo)
	result, err := useCase.SearchTaxonomy(context.Background(), request.TaxonomySearchRequest{
		Query: "  shirt  ", // trimmed before it reaches the repo
	})

	assert.NoError(t, err)
	assert.Len(t, result.Results, 1)
	assert.Equal(t, "Shirts", result.Results[0].Name)
	assert.Equal(t, "Clothing", result.Results[0].DepartmentName)
	assert.Equal(t, "Menswear", result.Results[0].CategoryName)
}

func TestSearchTaxonomy_EmptyQuery(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepo.NewMockSearchRepository(ctrl)
	useCase := NewSearchUseCase(mockRepo)

	result, err := useCase.SearchTaxonomy(context.Background(), request.TaxonomySearchRequest{Query: "   "})

	assert.NoError(t, err)
	assert.NotNil(t, result.Results)
	assert.Empty(t, result.Results)
}

func TestSearchTaxonomy_LimitCapping(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	mockRepo := mockrepo.NewMockSearchRepository(ctrl)
	mockRepo.EXPECT().
		SearchTaxonomy(gomock.Any(), "shirt", "dep_1", uint(50)).
		Return([]response.TaxonomySearchItem{}, nil)

	useCase := NewSearchUseCase(mockRepo)
	_, err := useCase.SearchTaxonomy(context.Background(), request.TaxonomySearchRequest{
		Query:        "shirt",
		DepartmentID: "dep_1",
		Limit:        500, // capped to maxTaxonomyLimit
	})

	assert.NoError(t, err)
}
