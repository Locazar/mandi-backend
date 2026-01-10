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

	// sub category
	IsSubCategoryNameExist(ctx context.Context, categoryName string, categoryID uint) (bool, error)
	FindAllSubCategories(ctx context.Context, categoryID uint) ([]response.SubCategory, error)
	SaveSubCategory(ctx context.Context, body request.SubCategory, departmentID string, categoryID string) error

	// variation
	IsVariationNameExistForCategory(ctx context.Context, name string, categoryID uint) (bool, error)
	SaveVariation(ctx context.Context, categoryID uint, variationName string) error
	FindAllVariationsByCategoryID(ctx context.Context, categoryID uint) ([]response.Variation, error)

	// variation values
	IsVariationValueExistForVariation(ctx context.Context, value string, variationID uint) (exist bool, err error)
	SaveVariationOption(ctx context.Context, variationID uint, variationValue string) error
	FindAllVariationOptionsByVariationID(ctx context.Context, variationID uint) ([]response.VariationOption, error)

	FindAllVariationValuesOfProductItem(ctx context.Context, productItemID uint) ([]response.ProductVariationValue, error)
	//product
	FindProductByID(ctx context.Context, productID uint) (product domain.Product, err error)
	IsProductNameExistForOtherProduct(ctx context.Context, name string, productID uint) (bool, error)
	IsProductNameExist(ctx context.Context, productName string) (exist bool, err error)

	FindAllProducts(ctx context.Context, pagination request.Pagination, search string) ([]response.Product, error)
	SaveProduct(ctx context.Context, product domain.Product, adminID string) (productID uint, err error)
	UpdateProduct(ctx context.Context, product domain.Product) error

	// product items
	FindProductItemByID(ctx context.Context, productItemID uint) (domain.ProductItem, error)
	FindAllProductItems(ctx context.Context, adminID string, keyword string, categoryID, brandID, locationID *string, offer string, sortby string, pagination *request.Pagination) ([]response.ProductItems, error)
	DeleteProductItem(ctx context.Context, productItemID uint) error
	FindProductItemFilters(ctx context.Context, adminID string) ([]domain.ProductItemFilterType, error)
	// GetProductItemsByDepartment returns product items associated with a department id
	GetProductItemsByDepartment(ctx context.Context, departmentID uint) ([]response.ProductItems, error)
	GetProductItemsByCategory(ctx context.Context, categoryID uint) ([]response.ProductItems, error)
	GetProductItemsBySubCategory(ctx context.Context, subCategoryID uint) ([]response.ProductItems, error)
	GetProductItemsByShop(ctx context.Context, adminID uint) ([]response.ProductItems, error)
	FindVariationCountForProduct(ctx context.Context, productID uint) (variationCount uint, err error) // to check the product config already exist
	FindAllProductItemIDsByProductIDAndVariationOptionID(ctx context.Context, productID, variationOptionID uint) ([]uint, error)
	SaveProductConfiguration(ctx context.Context, productItemID, variationOptionID uint) error
	SaveProductItem(ctx context.Context, productItem request.ProductItem, adminID string) (productItemID uint, err error)
	// product item image
	FindAllProductItemImages(ctx context.Context, productItemID uint) (images []string, err error)
	SaveProductItemImage(ctx context.Context, productItemId uint, image domain.ProductItemImage) error

	SearchProducts(ctx context.Context, keyword string, categoryID, brandID, locationID *string, pagination request.Pagination) (products []response.ProductItems, err error)

	// department
	SaveDepartment(ctx context.Context, departmentName string) error
	GetAllDepartments(ctx context.Context) ([]response.Department, error)
	GetDepartmentByID(ctx context.Context, departmentID uint) (response.Department, error)

	GetAllSubCategories(ctx context.Context) ([]response.SubCategory, error)

	GetAllCategoriesByDepartmentID(ctx context.Context, departmentID uint) ([]response.Category, error)

	GetAllSubCategoriesByCategoryID(ctx context.Context, categoryID uint) ([]response.SubCategory, error)

	// sub type attributes
	SaveSubTypeAttribute(ctx context.Context, subCategoryID uint, attribute domain.SubTypeAttributes) error
	GetAllSubTypeAttributes(ctx context.Context, subCategoryID uint) ([]response.SubTypeAttribute, error)
	GetSubTypeAttributeByID(ctx context.Context, attributeID uint) (response.SubTypeAttribute, error)

	// sub type attribute options
	SaveSubTypeAttributeOption(ctx context.Context, attributeID uint, option domain.SubTypeAttributeOptions) error
	GetAllSubTypeAttributeOptions(ctx context.Context, attributeID uint) ([]response.SubTypeAttributeOption, error)
	GetSubTypeAttributeOptionByID(ctx context.Context, optionID uint) (response.SubTypeAttributeOption, error)

	// category images
	SaveCategoryImage(ctx context.Context, categoryID uint, image domain.CategoryImage) error
	GetAllCategoryImages(ctx context.Context, categoryID uint) ([]response.CategoryImage, error)
	GetCategoryImageByID(ctx context.Context, imageID uint) (response.CategoryImage, error)
	UpdateCategoryImage(ctx context.Context, image domain.CategoryImage) error
	DeleteCategoryImage(ctx context.Context, imageID uint) error
	GetProductItemByID(ctx context.Context, productItemID uint) (response.ProductItems, error)
	IncrementProductItemViewCount(ctx context.Context, productItemID uint, adminID string) error
	GetProductItemViewCount(ctx context.Context, productItemID uint, adminID string) (uint, error)
}
