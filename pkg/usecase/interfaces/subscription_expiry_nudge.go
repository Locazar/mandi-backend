package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// SubscriptionExpiryNudgeUseCase manages the daily automated reminder sent to
// sellers whose subscription is about to expire.
type SubscriptionExpiryNudgeUseCase interface {
	GetSettings(ctx context.Context) (domain.SubscriptionExpiryNudgeSettings, error)
	SaveSettings(ctx context.Context, settings domain.SubscriptionExpiryNudgeSettings) error

	GetTemplate(ctx context.Context) (domain.SubscriptionExpiryNudgeTemplate, error)
	SaveTemplate(ctx context.Context, tmpl domain.SubscriptionExpiryNudgeTemplate) error

	// RunSweep finds every seller expiring within the configured window who
	// hasn't been nudged today and sends them the template. Called once a day
	// by the in-process ticker in cmd/api/main.go, and also triggerable on
	// demand via the admin-portal "Run now" button. Idempotent: the sent
	// ledger prevents double-sends within the same calendar day.
	RunSweep(ctx context.Context) (SubscriptionExpiryNudgeSweepResult, error)
}

type SubscriptionExpiryNudgeSweepResult struct {
	Sent     int  `json:"sent"`
	Errors   int  `json:"errors"`
	Disabled bool `json:"disabled"`
}
