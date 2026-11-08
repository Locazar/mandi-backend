package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type LeadRequestUseCase interface {
	Submit(ctx context.Context, phone, requirement, city, source string) (*domain.LeadRequest, error)
	List(ctx context.Context, city, status string, limit, offset int) ([]domain.LeadRequest, int64, error)
	UpdateStatus(ctx context.Context, id uint, status, adminNote string) error
}
