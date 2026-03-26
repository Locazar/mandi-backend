package interfaces

import "github.com/gin-gonic/gin"

type OfferHandler interface {

	// offer
	SaveOffer(ctx *gin.Context)
	RemoveOffer(ctx *gin.Context)
	GetAllOffers(ctx *gin.Context)
	GetActiveOffers(ctx *gin.Context)

	// category offer
	GetAllCategoryOffers(ctx *gin.Context)
	SaveCategoryOffer(ctx *gin.Context)
	RemoveCategoryOffer(ctx *gin.Context)
	ChangeCategoryOffer(ctx *gin.Context)

	// product offer
	GetAllProductsOffers(ctx *gin.Context)
	SaveProductItemOffer(ctx *gin.Context)
	RemoveProductOffer(ctx *gin.Context)
	ChangeProductOffer(ctx *gin.Context)

	//Shop offer
	ApplyOfferToShop(ctx *gin.Context)
	GetShopOffers(ctx *gin.Context)

	// Post login offer decision
	PostLoginOffer(ctx *gin.Context)

	// Banner
	GetBanners(ctx *gin.Context)

	// Deprecated: use GetShopOffers instead
	GetShopOffersByShopID(ctx *gin.Context)
}
