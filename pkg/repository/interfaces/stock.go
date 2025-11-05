package interfaces

import (
	"context"

<<<<<<< HEAD
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
=======
	"github.com/nikhilnarayanan623/ecommerce-gin-clean-arch/pkg/api/handler/request"
	"github.com/nikhilnarayanan623/ecommerce-gin-clean-arch/pkg/api/handler/response"
>>>>>>> b9ab446 (Initial commit)
)

type StockRepository interface {
	FindAll(ctx context.Context, pagination request.Pagination) (stocks []response.Stock, err error)
	Update(ctx context.Context, updateValues request.UpdateStock) error
}
