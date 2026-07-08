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
	CreateCategory(ctx context.Context, departmentID, name, imageURL string, sortOrder int, isActive bool) error
	UpdateCategory(ctx context.Context, categoryID, name, imageURL string, sortOrder int, isActive bool) error
	DeleteCategory(ctx context.Context, categoryID string) error
	SaveSubCategory(ctx context.Context, body request.SubCategory, departmentId string, category_id string) error
	CreateSubCategory(ctx context.Context, departmentID, categoryID, name, imageURL string, sortOrder int, isActive bool) error
	UpdateSubCategory(ctx context.Context, subCategoryID, name, imageURL string, sortOrder int, isActive bool) error
	DeleteSubCategory(ctx context.Context, subCategoryID string) error

	// variations
	SaveVariation(ctx context.Context, categoryID string, variationNames []string) error
	SaveVariationOption(ctx context.Context, variationID string, variationOptionValues []string) error

	FindAllVariationsAndItsValues(ctx context.Context, categoryID string) ([]response.Variation, error)

	// products
	FindAllProducts(ctx context.Context, pagination request.Pagination, search string) (products []response.Product, err error)
	FindProductByID(ctx context.Context, productID string) (product domain.Product, err error)
	SaveProduct(ctx context.Context, product request.Product, adminID string) (productID string, err error)
	UpdateProduct(ctx context.Context, product domain.Product) error

	SaveProductItem(ctx context.Context, productItem request.ProductItem, adminID string, shopID string) error
	FindAllProductItems(ctx context.Context, shopID string, keyword string, categoryID, brandID, locationID *string, offer string, sortby string, pagination *request.Pagination, filterByShopID string) ([]response.ProductItems, error)
	FindLowViewProductItems(ctx context.Context, shopID string, keyword string, categoryID, brandID, locationID *string, sortby string, pagination *request.Pagination, filterByShopID *string) ([]response.ProductItems, error)
	UpdateProductItem(ctx context.Context, productItemID string, productItem request.ProductItem) error
	UpdateProductItemStock(ctx context.Context, productItemID string, inStock bool) error
	DeleteProductItem(ctx context.Context, productItemID string) error
	FindProductItemFilters(ctx context.Context, adminID string, shopID string) ([]domain.ProductItemFilterType, error)
	SearchProducts(ctx context.Context, keyword string, categoryID, departmentID, brandID, locationID *string, shopID *string, latitude, longitude, radius float64, pincode *uint, limit, offset int) (products []response.ProductItems, err error)
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
	GetNearbyProductsByPincode(ctx context.Context, pincode string, limit, offset int) (products []response.ProductItems, err error)
	GetProductsByRadius(ctx context.Context, latitude float64, longitude float64, radiusKm float64, limit, offset int) ([]response.ProductItems, error)

	// department
	SaveDepartment(ctx context.Context, department request.Department) error
	GetAllDepartments(ctx context.Context) ([]response.Department, error)
	GetDepartmentByID(ctx context.Context, departmentID string) (response.Department, error)
	CreateDepartment(ctx context.Context, name, imageURL string, sortOrder int, isActive bool) error
	UpdateDepartment(ctx context.Context, departmentID, name, imageURL string, sortOrder int, isActive bool) error
	DeleteDepartment(ctx context.Context, departmentID string) error

	GetAllSubCategories(ctx context.Context) ([]response.SubCategory, error)
	GetAllCategoriesByDepartmentID(ctx context.Context, departmentID string) ([]response.Category, error)
	GetAllSubCategoriesByCategoryID(ctx context.Context, categoryID string) ([]response.SubCategory, error)

	// sub type attributes
	SaveSubTypeAttribute(ctx context.Context, subCategoryID string, attribute request.SubTypeAttribute) error
	GetAllSubTypeAttributes(ctx context.Context, subCategoryID string) ([]response.SubTypeAttribute, error)
	GetSubTypeAttributeByID(ctx context.Context, attributeID string) (response.SubTypeAttribute, error)
	UpdateSubTypeAttribute(ctx context.Context, attributeID string, attribute request.SubTypeAttribute) error
	DeleteSubTypeAttribute(ctx context.Context, attributeID string) error

	// sub type attribute options
	SaveSubTypeAttributeOption(ctx context.Context, attributeID string, option request.SubTypeAttributeOption) error
	GetAllSubTypeAttributeOptions(ctx context.Context, attributeID string) ([]response.SubTypeAttributeOption, error)
	GetSubTypeAttributeOptionByID(ctx context.Context, optionID string) (response.SubTypeAttributeOption, error)
	UpdateSubTypeAttributeOption(ctx context.Context, optionID string, option request.SubTypeAttributeOption) error
	DeleteSubTypeAttributeOption(ctx context.Context, optionID string) error

	// category images
	SaveCategoryImage(ctx context.Context, categoryID string, image request.CategoryImage) error
	GetAllCategoryImages(ctx context.Context, categoryID string) ([]response.CategoryImage, error)
	GetCategoryImageByID(ctx context.Context, imageID string) (response.CategoryImage, error)
	UpdateCategoryImage(ctx context.Context, imageID string, image request.CategoryImage) error
	DeleteCategoryImage(ctx context.Context, imageID string) error
	GetProductItemByID(ctx context.Context, productItemID string) (response.ProductItems, error)
	IncrementProductItemViewCount(ctx context.Context, productItemID string, adminID string) error
	GetProductItemViewCount(ctx context.Context, productItemID string, adminID string) (uint, error)

	// Offer
	GetProductItemsByOfferID(ctx context.Context, offerID string, categoryID int, departmentID int, subCategoryID int, latStr string, lngStr string, pincode string, radiusKm float64, limit int, offset int) ([]response.ProductItems, error)
}
