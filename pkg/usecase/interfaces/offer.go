package interfaces

import (
	"context"

<<<<<<< HEAD
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
=======
	"github.com/nikhilnarayanan623/ecommerce-gin-clean-arch/pkg/api/handler/request"
	"github.com/nikhilnarayanan623/ecommerce-gin-clean-arch/pkg/api/handler/response"
	"github.com/nikhilnarayanan623/ecommerce-gin-clean-arch/pkg/domain"
>>>>>>> b9ab446 (Initial commit)
)

type OfferUseCase interface {

	// offer
	SaveOffer(ctx context.Context, offer request.Offer) error
	RemoveOffer(ctx context.Context, offerID uint) error
	FindAllOffers(ctx context.Context, pagination request.Pagination) ([]domain.Offer, error)

	// offer category
	SaveCategoryOffer(ctx context.Context, offerCategory request.OfferCategory) error
	FindAllCategoryOffers(ctx context.Context, pagination request.Pagination) ([]response.OfferCategory, error)
	RemoveCategoryOffer(ctx context.Context, categoryOfferID uint) error
	ChangeCategoryOffer(ctx context.Context, categoryOfferID, offerID uint) error

	// offer product
	SaveProductOffer(ctx context.Context, offerProduct domain.OfferProduct) error
	FindAllProductOffers(ctx context.Context, pagination request.Pagination) ([]response.OfferProduct, error)
	RemoveProductOffer(ctx context.Context, productOfferID uint) error
	ChangeProductOffer(ctx context.Context, productOfferID, offerID uint) error
}
