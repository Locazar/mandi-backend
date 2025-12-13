package repository

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type stockDatabase struct {
	DB *gorm.DB
}

func NewStockRepository(db *gorm.DB) interfaces.StockRepository {
	return &stockDatabase{
		DB: db,
	}
}

func (c *stockDatabase) Update(ctx context.Context, valuesToUpdate request.UpdateStock) error {

	query := `UPDATE product_items SET qty_in_stock = qty_in_stock + $1 WHERE sku = $2`

	err := c.DB.Exec(query, valuesToUpdate.QtyToAdd, valuesToUpdate.SKU).Error

	return err
}

func (c *stockDatabase) FindAll(ctx context.Context, pagination request.Pagination) (stocks []response.Stock, err error) {

	limit := pagination.Count
	offset := (pagination.PageNumber - 1) * limit

	query := `SELECT pi.id AS product_item_id, p.name AS product_name
	FROM product_items pi 
	INNER JOIN products p ON p.id = pi.product_id
	ORDER BY product_item_id LIMIT $1 OFFSET $2`

	err = c.DB.Raw(query, limit, offset).Scan(&stocks).Error
	if err != nil {
		return nil, err
	}

	// insert each stocks variation full values
	query = `SELECT vo.id, vo.value FROM variation_options vo 
	INNER JOIN product_configurations pc ON vo.id = pc.variation_option_id 
	WHERE pc.product_item_id = $1`

	for i, stock := range stocks {

		var variationValue []response.VariationOption
		err = c.DB.Raw(query, stock.ProductItemID).Scan(&variationValue).Error
		if err != nil {
			return nil, err
		}
		stocks[i].VariationOptions = variationValue
	}

	return
}
