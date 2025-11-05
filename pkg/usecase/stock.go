package usecase

import (
	"context"
	"log"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	service "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type stockUseCase struct {
	stockRepo interfaces.StockRepository
}

func NewStockUseCase(stockRepo interfaces.StockRepository) service.StockUseCase {

	return &stockUseCase{
		stockRepo: stockRepo,
	}
}

func (c *stockUseCase) GetAllStockDetails(ctx context.Context, pagination request.Pagination) (stocks []response.Stock, err error) {
	stocks, err = c.stockRepo.FindAll(ctx, pagination)

	if err != nil {
		return stocks, err
	}

	log.Printf("successfully got stock details")
	return stocks, nil
}

func (c *stockUseCase) UpdateStockBySKU(ctx context.Context, updateDetails request.UpdateStock) error {

	err := c.stockRepo.Update(ctx, updateDetails)

	if err != nil {
		return err
	}

	log.Printf("successfully updated of stock details of stock with sku %v", updateDetails.SKU)
	return nil
}
