package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type ShopProductEnquiryUseCase interface {
	Submit(ctx context.Context, e domain.ShopProductEnquiry) error
	List(ctx context.Context, shopID string, actioned *bool, enquiryType string) ([]domain.ShopProductEnquiry, error)
	MarkActioned(ctx context.Context, id string, note string) error
	// NotifyShop sends a push notification to the shopkeeper about an enquiry.
	// Returns delivered=true only when at least one device was actually reached.
	NotifyShop(ctx context.Context, enquiryID string) (delivered bool, err error)
}
