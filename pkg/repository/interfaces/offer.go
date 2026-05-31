package interfaces

import (
	"context"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type OfferRepository interface {
	Transactions(ctx context.Context, trxFn func(repo OfferRepository) error) error

	// offer
	FindOfferByID(ctx context.Context, offerID string) (domain.Offer, error)
	FindOfferByName(ctx context.Context, offerName string) (domain.Offer, error)
	FindAllOffers(ctx context.Context, pagination request.Pagination) (response.OffersAndPromotions, error)
	SaveOffer(ctx context.Context, offer request.Offer) error
	SaveOfferWithImages(ctx context.Context, offer request.Offer, imageURL, thumbnailURL string) error
	DeleteOffer(ctx context.Context, offerID string) error

	// Get active offers based on start and end date
	FindActiveOffers(ctx context.Context) ([]domain.Offer, error)

	// to calculate the discount price and update
	UpdateProductsDiscountByCategoryOfferID(ctx context.Context, categoryOfferID string) error
	UpdateProductItemsDiscountByCategoryOfferID(ctx context.Context, categoryOfferID string) error
	ApplyOfferToProductItem(ctx context.Context, productOfferID string) error
	UpdateProductItemsDiscountByProductOfferID(ctx context.Context, productOfferID string) error

	// to remove the discount product price
	RemoveProductsDiscountByCategoryOfferID(ctx context.Context, categoryOfferID string) error
	RemoveProductItemsDiscountByCategoryOfferID(ctx context.Context, categoryOfferID string) error
	RemoveProductsDiscountByProductOfferID(ctx context.Context, productOfferID string) error
	RemoveProductItemsDiscountByProductOfferID(ctx context.Context, productOfferID string) error

	// offer category
	FindOfferCategoryCategoryID(ctx context.Context, categoryID string) (domain.OfferCategory, error)
	FindAllOfferCategories(ctx context.Context, pagination request.Pagination) ([]response.OfferCategory, error)

	SaveCategoryOffer(ctx context.Context, categoryOffer request.OfferCategory) (categoryOfferID string, err error)
	DeleteCategoryOffer(ctx context.Context, categoryOfferID string) error
	UpdateCategoryOffer(ctx context.Context, categoryOfferID, offerID string) error

	// offer products
	FindOfferProductByProductID(ctx context.Context, productID string) (domain.OfferProduct, error)
	FindAllOfferProducts(ctx context.Context, pagination request.Pagination) ([]response.OfferProduct, error)

	SaveOfferProduct(ctx context.Context, offerProduct domain.OfferProduct) (productOfferId string, err error)
	DeleteOfferProduct(ctx context.Context, productOfferID string) error
	UpdateOfferProduct(ctx context.Context, productOfferID, offerID string) error

	DeleteAllProductOffersByOfferID(ctx context.Context, offerID string) error
	DeleteAllCategoryOffersByOfferID(ctx context.Context, offerID string) error
	ApplyOfferToShop(ctx context.Context, adminID string, shopID string, body request.ApplyOfferToShop) error
	FindShopOffersByShopIDAndDateRange(ctx context.Context, shopID string, startDate, endDate time.Time) ([]domain.ShopOffer, error)
	FindShopOffersByShopID(ctx context.Context, shopID string, adminID string) ([]domain.ShopOffer, error)
}
