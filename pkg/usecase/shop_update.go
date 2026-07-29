package usecase

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
)

// ShopUpdateUseCase is a thin passthrough over the repository for the per-shop
// "What's New" advertisement feature.
type ShopUpdateUseCase struct {
	repo repo.ShopUpdateRepository
}

func NewShopUpdateUseCase(repo repo.ShopUpdateRepository) *ShopUpdateUseCase {
	return &ShopUpdateUseCase{repo: repo}
}

func (u *ShopUpdateUseCase) Create(ctx context.Context, su domain.ShopUpdate) (domain.ShopUpdate, error) {
	return u.repo.Create(ctx, su)
}

func (u *ShopUpdateUseCase) List(ctx context.Context) ([]domain.ShopUpdate, error) {
	return u.repo.List(ctx)
}

func (u *ShopUpdateUseCase) GetByID(ctx context.Context, id string) (domain.ShopUpdate, error) {
	return u.repo.GetByID(ctx, id)
}

func (u *ShopUpdateUseCase) Update(ctx context.Context, su domain.ShopUpdate) (domain.ShopUpdate, error) {
	return u.repo.Update(ctx, su)
}

func (u *ShopUpdateUseCase) Delete(ctx context.Context, id string) error {
	return u.repo.Delete(ctx, id)
}

func (u *ShopUpdateUseCase) AddProduct(ctx context.Context, p domain.ShopUpdateProduct) (domain.ShopUpdateProduct, error) {
	return u.repo.AddProduct(ctx, p)
}

func (u *ShopUpdateUseCase) UpdateProduct(ctx context.Context, p domain.ShopUpdateProduct) (domain.ShopUpdateProduct, error) {
	return u.repo.UpdateProduct(ctx, p)
}

func (u *ShopUpdateUseCase) DeleteProduct(ctx context.Context, id string) error {
	return u.repo.DeleteProduct(ctx, id)
}

func (u *ShopUpdateUseCase) GetForUser(ctx context.Context, lat, lng, radius float64, pincode string) ([]response.ShopUpdateCard, error) {
	return u.repo.GetForUser(ctx, lat, lng, radius, pincode)
}
