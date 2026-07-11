package interfaces

import (
	"context"
	"time"

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
	FindPaidOrdersByUserID(ctx context.Context, userID string) ([]domain.BillingOrder, error)
	FindActiveSubscriptionByUserID(ctx context.Context, userID string) (domain.UserSubscription, error)
	ActivateSubscription(ctx context.Context, sub domain.UserSubscription) error
	DeactivateTrialSubscription(ctx context.Context, userID string) error
	// SumGSTCollected returns the total GST collected from paid subscription
	// orders whose paid_at falls within [from, to].
	SumGSTCollected(ctx context.Context, from, to time.Time) (domain.Money, error)
	// GetGSTRateBasisPoints returns the admin-configured subscription GST
	// rate (1800 = 18.00%) read from the same subscription_gst_config
	// singleton row AdminRepository.GetSubscriptionGSTConfig manages. Falls
	// back to 1800 if the row is somehow missing, matching the previous
	// hardcoded env-var default, so a missing row never breaks checkout.
	GetGSTRateBasisPoints(ctx context.Context) (int, error)
	Transaction(fn func(repo SubscriptionRepository) error) error
}
