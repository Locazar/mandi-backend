package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type NotifyCustomerRequestUseCase interface {
	Submit(ctx context.Context, req domain.NotifyCustomerRequest) error
	List(ctx context.Context, actioned *bool) ([]domain.NotifyCustomerRequest, error)
	MarkActioned(ctx context.Context, id string, note string) error
}
