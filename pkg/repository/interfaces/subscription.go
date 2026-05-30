package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type SubscriptionRepository interface {
	CreateSubscriptionOrder(ctx context.Context, order domain.SubscriptionOrder) (domain.SubscriptionOrder, error)
	FindSubscriptionOrderByRazorpayOrderID(ctx context.Context, razorpayOrderID string) (domain.SubscriptionOrder, error)
	FindSubscriptionOrderByRazorpayPaymentID(ctx context.Context, paymentID string) (domain.SubscriptionOrder, error)
	UpdateSubscriptionOrderToPaid(ctx context.Context, orderID string, razorpayPaymentID string) error
	FindSubscriptionPlanByID(ctx context.Context, planID string) (domain.SubscriptionPlan, error)
	FindSubscriptionPlanByName(ctx context.Context, name string) (domain.SubscriptionPlan, error)
	FindPaidSubscriptionPlans(ctx context.Context) ([]domain.SubscriptionPlan, error)
	FindActiveSubscriptionByUserID(ctx context.Context, userID string) (domain.UserSubscription, error)
	ActivateSubscription(ctx context.Context, sub domain.UserSubscription) error
	DeactivateTrialSubscription(ctx context.Context, userID string) error
	Transaction(fn func(repo SubscriptionRepository) error) error
}
