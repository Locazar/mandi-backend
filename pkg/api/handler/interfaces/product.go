package interfaces

import "github.com/gin-gonic/gin"

type ProductHandler interface {
	GetAllCategories(ctx *gin.Context)
	SaveCategory(ctx *gin.Context)
	SaveSubCategory(ctx *gin.Context)
	SaveVariation(ctx *gin.Context)
	SaveVariationOption(ctx *gin.Context)
	GetAllVariations(ctx *gin.Context)

	GetAllProductsAdmin() func(ctx *gin.Context)
	GetAllProductsUser() func(ctx *gin.Context)

	SaveProduct(ctx *gin.Context)
	UpdateProduct(ctx *gin.Context)

	SaveProductItem(ctx *gin.Context)
	GetAllProductItemsAdmin() func(ctx *gin.Context)
	GetAllProductItemsUser() func(ctx *gin.Context)
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
}
