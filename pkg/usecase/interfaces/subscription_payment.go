package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
)

type SubscriptionPaymentUseCase interface {
	CreateSubscriptionOrder(ctx context.Context, userID string, req request.CreateSubscriptionOrderRequest) (response.SubscriptionOrderResponse, error)
	VerifySubscriptionPayment(ctx context.Context, userID string, req request.VerifySubscriptionPaymentRequest) (response.SubscriptionVerificationResponse, error)
	HandlePaymentFailure(ctx context.Context, userID string, req request.PaymentFailureRequest) error
	HandleWebhook(ctx context.Context, signature string, rawBody []byte) error
}
