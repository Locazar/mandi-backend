package usecase

import (
	"context"
	"fmt"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type notifyCustomerRequestUseCase struct {
	repo repo.NotifyCustomerRequestRepository
}

func NewNotifyCustomerRequestUseCase(r repo.NotifyCustomerRequestRepository) interfaces.NotifyCustomerRequestUseCase {
	return &notifyCustomerRequestUseCase{repo: r}
}

func (uc *notifyCustomerRequestUseCase) Submit(ctx context.Context, req domain.NotifyCustomerRequest) error {
	exists, err := uc.repo.ExistsByEnquiryID(ctx, req.EnquiryID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("notification request already submitted for this deal")
	}
	return uc.repo.Create(ctx, req)
}

func (uc *notifyCustomerRequestUseCase) List(ctx context.Context, actioned *bool) ([]domain.NotifyCustomerRequest, error) {
	return uc.repo.List(ctx, actioned)
}

func (uc *notifyCustomerRequestUseCase) MarkActioned(ctx context.Context, id string, note string) error {
	return uc.repo.MarkActioned(ctx, id, note)
}
