package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type LeadRequestRepository interface {
	Create(ctx context.Context, lead *domain.LeadRequest) error
	List(ctx context.Context, city, status string, limit, offset int) ([]domain.LeadRequest, int64, error)
	UpdateStatus(ctx context.Context, id uint, status, adminNote string) error
}
