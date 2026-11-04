package usecase

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type shopProductEnquiryUseCase struct {
	repo repo.ShopProductEnquiryRepository
}

func NewShopProductEnquiryUseCase(r repo.ShopProductEnquiryRepository) interfaces.ShopProductEnquiryUseCase {
	return &shopProductEnquiryUseCase{repo: r}
}

func (uc *shopProductEnquiryUseCase) Submit(ctx context.Context, e domain.ShopProductEnquiry) error {
	return uc.repo.Create(ctx, e)
}

func (uc *shopProductEnquiryUseCase) List(ctx context.Context, shopID string, actioned *bool, enquiryType string) ([]domain.ShopProductEnquiry, error) {
	return uc.repo.List(ctx, shopID, actioned, enquiryType)
}

func (uc *shopProductEnquiryUseCase) MarkActioned(ctx context.Context, id string, note string) error {
	return uc.repo.MarkActioned(ctx, id, note)
}
