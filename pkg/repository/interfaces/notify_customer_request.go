package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type NotifyCustomerRequestRepository interface {
	Create(ctx context.Context, req domain.NotifyCustomerRequest) error
	ExistsByEnquiryID(ctx context.Context, enquiryID string) (bool, error)
	List(ctx context.Context, actioned *bool) ([]domain.NotifyCustomerRequest, error)
	MarkActioned(ctx context.Context, id string, note string) error
}
