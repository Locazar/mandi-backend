package interfaces

import (
	"github.com/gin-gonic/gin"
)

type UserHandler interface {
	GetProfile(ctx *gin.Context)
	UpdateProfile(ctx *gin.Context)

	SaveAddress(ctx *gin.Context)
	GetAllAddresses(ctx *gin.Context)
	UpdateAddress(ctx *gin.Context)
	SaveToWishList(ctx *gin.Context)
	RemoveFromWishList(ctx *gin.Context)
	GetWishList(ctx *gin.Context)
	UploadProfileImage(ctx *gin.Context)
	GetSellerByRadius(ctx *gin.Context)
	GetSellerByPincode(ctx *gin.Context)
	SearchShopList(ctx *gin.Context)
	GetProductItemsByDepartment(ctx *gin.Context)
	GetProductItemsByCategory(ctx *gin.Context)
	GetProductItemsBySubCategory(ctx *gin.Context)
	GetProductItemsByShop(ctx *gin.Context)
	// GetAllJobs(ctx *gin.Context)
	// GetUserJobApplications(ctx *gin.Context)
	// DeleteJobApplication(ctx *gin.Context)
	// GetJobSearchSuggestions(ctx *gin.Context)
	// GetJobSearchFilters(ctx *gin.Context)
	// SearchJobs(ctx *gin.Context)
	// GetJobSearchLocations(ctx *gin.Context)

	// GetAllJobCategories(ctx *gin.Context)
	// GetJobsByCategory(ctx *gin.Context)
	// GetJobSubCategories(ctx *gin.Context)
	// GetAllCompanies(ctx *gin.Context)
	// GetJobsByCompany(ctx *gin.Context)
	// GetAllJobLocations(ctx *gin.Context)
	// GetJobsByLocation(ctx *gin.Context)
	// GetJobCategoryFilters(ctx *gin.Context)
	// GetCompanyFilters(ctx *gin.Context)
	// GetLocationFilters(ctx *gin.Context)
	// GetJobCategoryLocations(ctx *gin.Context)
	// SearchJobsInCategory(ctx *gin.Context)
	// GetJobsBySubCategory(ctx *gin.Context)
}
