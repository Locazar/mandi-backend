package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type SubscriptionRepository interface {
	CreateSubscriptionOrder(ctx context.Context, order domain.SubscriptionOrder) (domain.SubscriptionOrder, error)
	FindSubscriptionOrderByRazorpayOrderID(ctx context.Context, razorpayOrderID string) (domain.SubscriptionOrder, error)
	FindSubscriptionOrderByRazorpayPaymentID(ctx context.Context, paymentID string) (domain.SubscriptionOrder, error)
	UpdateSubscriptionOrderToPaid(ctx context.Context, orderID uint, razorpayPaymentID string) error
	FindSubscriptionPlanByID(ctx context.Context, planID uint) (domain.SubscriptionPlan, error)
	FindActiveSubscriptionByUserID(ctx context.Context, userID uint) (domain.UserSubscription, error)
	ActivateSubscription(ctx context.Context, sub domain.UserSubscription) error
	Transaction(fn func(repo SubscriptionRepository) error) error
}
