package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
)

type DealUseCase interface {
	GetAllDeals(ctx context.Context, filter request.DealFilter) ([]response.CustomerDeal, int64, error)
	GetDealFilters(ctx context.Context) (response.DealFilters, error)
	GetDealByID(ctx context.Context, dealID string) (response.CustomerDeal, error)
	GetProductsByDeal(ctx context.Context, dealID string, pagination request.Pagination) ([]response.CustomerDealProduct, int64, error)
	GetProductsWithDeals(ctx context.Context, filter request.DealFilter) ([]response.CustomerDealProduct, int64, error)
}
