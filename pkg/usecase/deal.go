package usecase

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	repoInterfaces "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	usecaseInterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type dealUseCase struct {
	dealRepo repoInterfaces.DealRepository
}

func NewDealUseCase(dealRepo repoInterfaces.DealRepository) usecaseInterfaces.DealUseCase {
	return &dealUseCase{dealRepo: dealRepo}
}

func (u *dealUseCase) GetAllDeals(ctx context.Context, filter request.DealFilter) ([]response.CustomerDeal, int64, error) {
	return u.dealRepo.GetAllDeals(ctx, filter)
}

func (u *dealUseCase) GetDealFilters(ctx context.Context) (response.DealFilters, error) {
	return u.dealRepo.GetDealFilters(ctx)
}

func (u *dealUseCase) GetDealByID(ctx context.Context, dealID string) (response.CustomerDeal, error) {
	return u.dealRepo.GetDealByID(ctx, dealID)
}

func (u *dealUseCase) GetProductsByDeal(ctx context.Context, dealID string, pagination request.Pagination) ([]response.CustomerDealProduct, int64, error) {
	return u.dealRepo.GetProductsByDeal(ctx, dealID, pagination)
}

func (u *dealUseCase) GetProductsWithDeals(ctx context.Context, filter request.DealFilter) ([]response.CustomerDealProduct, int64, error) {
	return u.dealRepo.GetProductsWithDeals(ctx, filter, filter.Pagination)
}
