package interfaces

import "github.com/gin-gonic/gin"

type ProductHandler interface {
	GetAllCategories(ctx *gin.Context)
	SaveCategory(ctx *gin.Context)
	SaveSubCategory(ctx *gin.Context)
	SaveVariation(ctx *gin.Context)
	SaveVariationOption(ctx *gin.Context)
	GetAllVariations(ctx *gin.Context)
	GetAllSubCategories(ctx *gin.Context)
	GetAllCategoriesByDepartmentID(ctx *gin.Context)
	GetAllSubCategoriesByCategoryID(ctx *gin.Context)
	FindLowViewProductItems(ctx *gin.Context)
	GetDepartmentByID(ctx *gin.Context)

	GetAllProductsAdmin(ctx *gin.Context)
	GetAllProductsUser(ctx *gin.Context)
	GetProductByID(ctx *gin.Context)

	SaveProduct(ctx *gin.Context)
	UpdateProduct(ctx *gin.Context)

	SaveProductItem(ctx *gin.Context)
	GetAllProductItemsAdmin() func(ctx *gin.Context)
	GetAllProductItemsUser() func(ctx *gin.Context)
	GetProductItemsByShopID() func(ctx *gin.Context)
	FindProductItemFilters(ctx *gin.Context)
	SearchProducts(ctx *gin.Context)
	GetProductsByCategory(ctx *gin.Context)
	GetAllBrands(ctx *gin.Context)
	GetProductsByBrand(ctx *gin.Context)
	GetCategoryFilters(ctx *gin.Context)
	GetBrandFilters(ctx *gin.Context)
	GetLocationFilter(ctx *gin.Context)
	GetProductsByLocation(ctx *gin.Context)
	GetAllAreas(ctx *gin.Context)
	GetAllCities(ctx *gin.Context)
	GetAllStates(ctx *gin.Context)
	GetAllCountries(ctx *gin.Context)
	GetAllPincodes(ctx *gin.Context)
	GetCitiesByState(ctx *gin.Context)
	GetAreasByCity(ctx *gin.Context)
	GetPincodesByArea(ctx *gin.Context)
	GetLocationByPincode(ctx *gin.Context)
	GetNearbyProductsByPincode(ctx *gin.Context)
	GetProductSearchSuggestions(ctx *gin.Context)
	GetProductSearchFilters(ctx *gin.Context)
	GetProductSearchLocations(ctx *gin.Context)
	GetProductItemByID(ctx *gin.Context)
	DeleteProductItem(ctx *gin.Context)
	UpdateProductItem(ctx *gin.Context)
	GetProductsByRadius(ctx *gin.Context)

	// department
	SaveDepartment(ctx *gin.Context)
	GetAllDepartments(ctx *gin.Context)

	// sub type attributes
	SaveSubTypeAttribute(ctx *gin.Context)
	GetAllSubTypeAttributes(ctx *gin.Context)
	GetSubTypeAttributeByID(ctx *gin.Context)

	// sub type attribute options
	SaveSubTypeAttributeOption(ctx *gin.Context)
	GetAllSubTypeAttributeOptions(ctx *gin.Context)
	GetSubTypeAttributeOptionByID(ctx *gin.Context)

	// category images
	SaveCategoryImage(ctx *gin.Context)
	GetAllCategoryImages(ctx *gin.Context)
	GetCategoryImageByID(ctx *gin.Context)
	UpdateCategoryImage(ctx *gin.Context)
	DeleteCategoryImage(ctx *gin.Context)
	IncrementProductItemViewCount(ctx *gin.Context)
	GetProductItemViewCount(ctx *gin.Context)

	GetProductItemsByOfferID(ctx *gin.Context)
}
