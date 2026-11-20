package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// SubscriptionExpiredNudgeRepository persists settings, the notification
// template, and the per-user-per-day idempotency ledger for the automated
// already-expired subscription reminder sweep.
type SubscriptionExpiredNudgeRepository interface {
	GetSettings(ctx context.Context) (domain.SubscriptionExpiredNudgeSettings, error)
	SaveSettings(ctx context.Context, settings domain.SubscriptionExpiredNudgeSettings) error

	GetTemplate(ctx context.Context) (domain.SubscriptionExpiredNudgeTemplate, error)
	SaveTemplate(ctx context.Context, tmpl domain.SubscriptionExpiredNudgeTemplate) error

	// FindAlreadyExpired returns sellers whose subscription has expired and
	// who have NOT already been nudged today. Sellers who have since renewed
	// (have a currently active subscription with end_date > NOW()) are excluded.
	FindAlreadyExpired(ctx context.Context) ([]domain.SubscriptionExpiredNudgeCandidate, error)

	// MarkSent records today's send for userID. Safe to call more than once
	// for the same (user, day) pair (ON CONFLICT DO NOTHING).
	MarkSent(ctx context.Context, userID string) error
}
