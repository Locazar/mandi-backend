package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type ShopProductEnquiryRepository interface {
	Create(ctx context.Context, e domain.ShopProductEnquiry) error
	List(ctx context.Context, shopID string, actioned *bool, enquiryType string) ([]domain.ShopProductEnquiry, error)
	MarkActioned(ctx context.Context, id string, note string) error
}
