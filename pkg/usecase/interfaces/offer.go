package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type OfferUseCase interface {

	// offer
	SaveOffer(ctx context.Context, offer request.Offer) error
	RemoveOffer(ctx context.Context, offerID string) error
	FindAllOffers(ctx context.Context, pagination request.Pagination) (response.OffersAndPromotions, error)
	FindActiveOffers(ctx context.Context) ([]domain.Offer, error)

	// offer category
	SaveCategoryOffer(ctx context.Context, offerCategory request.OfferCategory) error
	FindAllCategoryOffers(ctx context.Context, pagination request.Pagination) ([]response.OfferCategory, error)
	RemoveCategoryOffer(ctx context.Context, categoryOfferID string) error
	ChangeCategoryOffer(ctx context.Context, categoryOfferID, offerID string) error

	// offer product
	SaveProductItemOffer(ctx context.Context, offerProduct domain.OfferProduct) error
	FindAllProductOffers(ctx context.Context, pagination request.Pagination) ([]response.OfferProduct, error)
	RemoveProductOffer(ctx context.Context, productOfferID string) error
	ChangeProductOffer(ctx context.Context, productOfferID, offerID string) error
	ApplyOfferToShop(ctx context.Context, adminId string, shopID string, body request.ApplyOfferToShop) error
	GetShopOffersByShopIDAndDateRange(ctx context.Context, shopID string, startDate, endDate string) ([]domain.ShopOffer, error)

	// Post login offer decision
	GetPostLoginOffer(ctx context.Context, userID string) (response.PostLoginOfferResponse, error)

	// Banner
	GetBanners(ctx context.Context, departmentID, categoryID *string) ([]response.Banner, error)

	// Deprecated: use GetShopOffersByShopIDAndDateRange instead
	GetShopOffersByShopID(ctx context.Context, shopID string, adminID string) ([]domain.ShopOffer, error)
}
