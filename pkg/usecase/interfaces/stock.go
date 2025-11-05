package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
)

type StockUseCase interface {
	GetAllStockDetails(ctx context.Context, pagination request.Pagination) (stocks []response.Stock, err error)
	UpdateStockBySKU(ctx context.Context, updateDetails request.UpdateStock) error
}
