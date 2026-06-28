package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type ProductRepository interface {
	Transactions(ctx context.Context, trxFn func(repo ProductRepository) error) error

	// category
	FindAllMainCategories(ctx context.Context, pagination request.Pagination) ([]response.Category, error)
	SaveCategory(ctx context.Context, category request.Category, departmentId string) error
	CreateCategory(ctx context.Context, departmentID, name, imageURL string, sortOrder int, isActive bool) error
	UpdateCategory(ctx context.Context, categoryID, name, imageURL string, sortOrder int, isActive bool) error
	DeleteCategory(ctx context.Context, categoryID string) error

	// sub category
	IsSubCategoryNameExist(ctx context.Context, categoryName string, categoryID string) (bool, error)
	FindAllSubCategories(ctx context.Context, categoryID string) ([]response.SubCategory, error)
	SaveSubCategory(ctx context.Context, body request.SubCategory, departmentID string, categoryID string) error
	CreateSubCategory(ctx context.Context, departmentID, categoryID, name, imageURL string, sortOrder int, isActive bool) error
	UpdateSubCategory(ctx context.Context, subCategoryID, name, imageURL string, sortOrder int, isActive bool) error
	DeleteSubCategory(ctx context.Context, subCategoryID string) error

	// variation
	IsVariationNameExistForCategory(ctx context.Context, name string, categoryID string) (bool, error)
	SaveVariation(ctx context.Context, categoryID string, variationName string) error
	FindAllVariationsByCategoryID(ctx context.Context, categoryID string) ([]response.Variation, error)

	// variation values
	IsVariationValueExistForVariation(ctx context.Context, value string, variationID string) (exist bool, err error)
	SaveVariationOption(ctx context.Context, variationID string, variationValue string) error
	FindAllVariationOptionsByVariationID(ctx context.Context, variationID string) ([]response.VariationOption, error)

	FindAllVariationValuesOfProductItem(ctx context.Context, productItemID string) ([]response.ProductVariationValue, error)
	// product
	FindProductByID(ctx context.Context, productID string) (product domain.Product, err error)
	IsProductNameExistForOtherProduct(ctx context.Context, name string, productID string) (bool, error)
	IsProductNameExist(ctx context.Context, productName string) (exist bool, err error)

	FindAllProducts(ctx context.Context, pagination request.Pagination, search string) ([]response.Product, error)
	SaveProduct(ctx context.Context, product domain.Product, adminID string) (productID string, err error)
	UpdateProduct(ctx context.Context, product domain.Product) error

	// product items
	FindProductItemByID(ctx context.Context, productItemID string) (domain.ProductItem, error)
	FindAllProductItems(ctx context.Context, adminID string, keyword string, categoryID, brandID, locationID *string, offer string, sortby string, pagination *request.Pagination, filterByShopID string) ([]response.ProductItems, error)
	FindLowViewProductItems(ctx context.Context, adminID string, keyword string, categoryID, brandID, locationID *string, sortby string, pagination *request.Pagination, filterByShopID *string) ([]response.ProductItems, error)
	DeleteProductItem(ctx context.Context, productItemID string) error
	FindProductItemFilters(ctx context.Context, adminID string, shopID string) ([]domain.ProductItemFilterType, error)
	// GetProductItemsByDepartment returns product items associated with a department id
	GetProductItemsByDepartment(ctx context.Context, departmentID string) ([]response.ProductItems, error)
	GetProductItemsByCategory(ctx context.Context, categoryID string) ([]response.ProductItems, error)
	GetProductItemsBySubCategory(ctx context.Context, subCategoryID string) ([]response.ProductItems, error)
	GetProductItemsByShop(ctx context.Context, adminID string) ([]response.ProductItems, error)
	FindVariationCountForProduct(ctx context.Context, productID string) (variationCount uint, err error) // to check the product config already exist
	FindAllProductItemIDsByProductIDAndVariationOptionID(ctx context.Context, productID, variationOptionID string) ([]string, error)
	SaveProductConfiguration(ctx context.Context, productItemID, variationOptionID string) error
	SaveProductItem(ctx context.Context, productItem request.ProductItem, adminID string, shopID string) (productItemID string, err error)
	UpdateProductItem(ctx context.Context, productItemID string, productItem request.ProductItem) error
	// product item image
	FindAllProductItemImages(ctx context.Context, productItemID string) (images []string, err error)
	SaveProductItemImage(ctx context.Context, productItemId string, image domain.ProductItemImage) error

	SearchProducts(ctx context.Context, keyword string, categoryID, departmentID, brandID, locationID *string, shopID *string, latitude, longitude, radius float64, pincode *uint, pagination request.Pagination) (products []response.ProductItems, err error)

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
	SaveSubTypeAttribute(ctx context.Context, subCategoryID string, attribute domain.SubTypeAttributes) error
	GetAllSubTypeAttributes(ctx context.Context, subCategoryID string) ([]response.SubTypeAttribute, error)
	GetSubTypeAttributeByID(ctx context.Context, attributeID string) (response.SubTypeAttribute, error)
	UpdateSubTypeAttribute(ctx context.Context, attributeID string, attribute domain.SubTypeAttributes) error
	DeleteSubTypeAttribute(ctx context.Context, attributeID string) error

	// sub type attribute options
	SaveSubTypeAttributeOption(ctx context.Context, attributeID string, option domain.SubTypeAttributeOptions) error
	GetAllSubTypeAttributeOptions(ctx context.Context, attributeID string) ([]response.SubTypeAttributeOption, error)
	GetSubTypeAttributeOptionByID(ctx context.Context, optionID string) (response.SubTypeAttributeOption, error)
	UpdateSubTypeAttributeOption(ctx context.Context, optionID string, option domain.SubTypeAttributeOptions) error
	DeleteSubTypeAttributeOption(ctx context.Context, optionID string) error

	// category images
	SaveCategoryImage(ctx context.Context, categoryID string, image domain.CategoryImage) error
	GetAllCategoryImages(ctx context.Context, categoryID string) ([]response.CategoryImage, error)
	GetCategoryImageByID(ctx context.Context, imageID string) (response.CategoryImage, error)
	UpdateCategoryImage(ctx context.Context, image domain.CategoryImage) error
	DeleteCategoryImage(ctx context.Context, imageID string) error
	GetProductItemByID(ctx context.Context, productItemID string) (response.ProductItems, error)
	IncrementProductItemViewCount(ctx context.Context, productItemID string, adminID string) error
	GetProductItemViewCount(ctx context.Context, productItemID string, adminID string) (uint, error)

	GetProductItemsByOfferID(ctx context.Context, offerID string, categoryID int, departmentID int, subCategoryID int, latStr string, lngStr string, pincode string, radiusKm float64, limit int, offset int) ([]response.ProductItems, error)
}
