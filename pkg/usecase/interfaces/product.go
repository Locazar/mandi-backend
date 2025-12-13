package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type ProductUseCase interface {
	FindAllCategories(ctx context.Context, pagination request.Pagination) ([]response.Category, error)
	SaveCategory(ctx context.Context, body request.Category, departmentId string) error
	SaveSubCategory(ctx context.Context, body request.SubCategory, departmentId string, category_id string) error

	// variations
	SaveVariation(ctx context.Context, categoryID uint, variationNames []string) error
	SaveVariationOption(ctx context.Context, variationID uint, variationOptionValues []string) error

	FindAllVariationsAndItsValues(ctx context.Context, categoryID uint) ([]response.Variation, error)

	// products
	FindAllProducts(ctx context.Context, pagination request.Pagination) (products []response.Product, err error)
	FindProductByID(ctx context.Context, productID uint) (product domain.Product, err error)
	SaveProduct(ctx context.Context, product request.Product, adminID string) (productID uint, err error)
	UpdateProduct(ctx context.Context, product domain.Product) error

	SaveProductItem(ctx context.Context, productItem request.ProductItem, productID uint) error
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
	GetProductsByRadius(ctx context.Context, latitude int, longitude, radius int, limit, offset int) ([]response.Product, error)

	// department
	SaveDepartment(ctx context.Context, departmentName string) error
	GetAllDepartments(ctx context.Context) ([]response.Department, error)
	GetDepartmentByID(ctx context.Context, departmentID uint) (response.Department, error)

	GetAllSubCategories(ctx context.Context) ([]response.SubCategory, error)
	GetAllCategoriesByDepartmentID(ctx context.Context, departmentID uint) ([]response.Category, error)
	GetAllSubCategoriesByCategoryID(ctx context.Context, categoryID uint) ([]response.SubCategory, error)

	// sub type attributes
	SaveSubTypeAttribute(ctx context.Context, subCategoryID uint, attribute request.SubTypeAttribute) error
	GetAllSubTypeAttributes(ctx context.Context, subCategoryID uint) ([]response.SubTypeAttribute, error)
	GetSubTypeAttributeByID(ctx context.Context, attributeID uint) (response.SubTypeAttribute, error)

	// sub type attribute options
	SaveSubTypeAttributeOption(ctx context.Context, attributeID uint, option request.SubTypeAttributeOption) error
	GetAllSubTypeAttributeOptions(ctx context.Context, attributeID uint) ([]response.SubTypeAttributeOption, error)
	GetSubTypeAttributeOptionByID(ctx context.Context, optionID uint) (response.SubTypeAttributeOption, error)

	// category images
	SaveCategoryImage(ctx context.Context, categoryID uint, image request.CategoryImage) error
	GetAllCategoryImages(ctx context.Context, categoryID uint) ([]response.CategoryImage, error)
	GetCategoryImageByID(ctx context.Context, imageID uint) (response.CategoryImage, error)
	UpdateCategoryImage(ctx context.Context, imageID uint, image request.CategoryImage) error
	DeleteCategoryImage(ctx context.Context, imageID uint) error
	GetProductItemByID(ctx context.Context, productItemID uint) (response.ProductItems, error)
}
