package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type ProductUseCase interface {
	FindAllCategories(ctx context.Context, pagination request.Pagination) ([]response.Category, error)
	SaveCategory(ctx context.Context, categoryName string) error
	SaveSubCategory(ctx context.Context, subCategory request.SubCategory) error

	// variations
	SaveVariation(ctx context.Context, categoryID uint, variationNames []string) error
	SaveVariationOption(ctx context.Context, variationID uint, variationOptionValues []string) error

	FindAllVariationsAndItsValues(ctx context.Context, categoryID uint) ([]response.Variation, error)

	// products
	FindAllProducts(ctx context.Context, pagination request.Pagination) (products []response.Product, err error)
	SaveProduct(ctx context.Context, product request.Product) error
	UpdateProduct(ctx context.Context, product domain.Product) error

	SaveProductItem(ctx context.Context, productID uint, productItem request.ProductItem) error
	FindAllProductItems(ctx context.Context, productID uint) ([]response.ProductItems, error)
	SearchProducts(ctx context.Context, keyword string, categoryID, brandID, locationID *string, limit, offset int) (products []response.Product, err error)
	GetProductNameSuggestions(ctx context.Context, prefix string) (suggestions []string, err error)
	GetProductFilters(ctx context.Context) (filters response.ProductFilters, err error)
	GetProductLocations(ctx context.Context) (locations []response.Location, err error)
	GetProductsByCategory(ctx context.Context, categoryID, limit, offset int) (products []response.Product, err error)
	GetAllBrands(ctx context.Context) ([]response.Brand, error)
	GetProductsByBrand(ctx context.Context, brandID, limit, offset int) (products []response.Product, err error)
	GetCategoryFilters(ctx context.Context) (categories []response.Category, err error)
	GetBrandFilters(ctx context.Context) (brands []response.Brand, err error)
	GetLocationFilter(ctx context.Context) (locations []response.Location, err error)
	GetProductsByLocation(ctx context.Context, locationID, limit, offset int) (products []response.Product, err error)
	GetAllAreas(ctx context.Context) (areas []response.Area, err error)
	GetAllCities(ctx context.Context) (cities []string, err error)
	GetAllStates(ctx context.Context) (states []string, err error)
	GetAllCountries(ctx context.Context) (countries []string, err error)
	GetAllPincodes(ctx context.Context) (pincodes []string, err error)
	GetCitiesByState(ctx context.Context, stateID string) (cities []string, err error)
	GetAreasByCity(ctx context.Context, cityID string) (areas []string, err error)
	GetPincodesByArea(ctx context.Context, areaID string) (pincodes []string, err error)
	GetLocationByPincode(ctx context.Context, pincodeID string) (location response.Location, err error)
	GetNearbyProductsByPincode(ctx context.Context, pincode string, limit, offset int) (products []response.Product, err error)
}
