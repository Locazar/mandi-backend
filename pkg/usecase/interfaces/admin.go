package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type AdminUseCase interface {
	SignUp(ctx context.Context, admin domain.Admin) error

	FindAllUser(ctx context.Context, pagination request.Pagination) (users []response.User, err error)
	BlockOrUnBlockUser(ctx context.Context, blockDetails request.BlockUser) error

	GetFullSalesReport(ctx context.Context, requestData request.SalesReport) (salesReport []response.SalesReport, err error)
}

// GetCategory(ctx context.Context) (helper.Category, any)
// 	SetCategory(ctx context.Context, body helper.Category)
