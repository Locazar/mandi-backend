package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// ShopUpdateRepository is the persistence contract for the per-shop
// "What's New" advertisement feature (admin CRUD + customer geo/pincode read).
type ShopUpdateRepository interface {
	// Admin CRUD
	Create(ctx context.Context, su domain.ShopUpdate) (domain.ShopUpdate, error)
	List(ctx context.Context) ([]domain.ShopUpdate, error)
	GetByID(ctx context.Context, id string) (domain.ShopUpdate, error)
	Update(ctx context.Context, su domain.ShopUpdate) (domain.ShopUpdate, error)
	Delete(ctx context.Context, id string) error

	// Product items
	AddProduct(ctx context.Context, p domain.ShopUpdateProduct) (domain.ShopUpdateProduct, error)
	UpdateProduct(ctx context.Context, p domain.ShopUpdateProduct) (domain.ShopUpdateProduct, error)
	DeleteProduct(ctx context.Context, id string) error
	ListProducts(ctx context.Context, shopUpdateID string) ([]domain.ShopUpdateProduct, error)

	// Customer read: only enabled entries, sorted by proximity (lat/lng+radius)
	// else pincode, else admin order. Empty slice is a valid result.
	GetForUser(ctx context.Context, lat, lng, radius float64, pincode string) ([]response.ShopUpdateCard, error)
}
