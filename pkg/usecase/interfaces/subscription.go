package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
)

type SubscriptionUseCase interface {
	GetSubscriptionStatus(ctx context.Context, userID uint) (response.SubscriptionStatusResponse, error)
	StartTrial(ctx context.Context, userID uint) (response.SubscriptionStatusResponse, error)
	GetPaidPlans(ctx context.Context) ([]response.SubscriptionPlanResponse, error)
}
