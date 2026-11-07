package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// SubscriptionExpiryNudgeRepository persists settings, the notification
// template, and the per-user-per-day idempotency ledger for the automated
// subscription-expiry reminder sweep.
type SubscriptionExpiryNudgeRepository interface {
	GetSettings(ctx context.Context) (domain.SubscriptionExpiryNudgeSettings, error)
	SaveSettings(ctx context.Context, settings domain.SubscriptionExpiryNudgeSettings) error

	GetTemplate(ctx context.Context) (domain.SubscriptionExpiryNudgeTemplate, error)
	SaveTemplate(ctx context.Context, tmpl domain.SubscriptionExpiryNudgeTemplate) error

	// FindExpiringSoon returns sellers with active subscriptions expiring within
	// daysBeforeExpiry calendar days who have NOT already been nudged today.
	FindExpiringSoon(ctx context.Context, daysBeforeExpiry int) ([]domain.SubscriptionExpiryNudgeCandidate, error)

	// MarkSent records today's send for userID. Safe to call more than once
	// for the same (user, day) pair (ON CONFLICT DO NOTHING).
	MarkSent(ctx context.Context, userID string) error
}
