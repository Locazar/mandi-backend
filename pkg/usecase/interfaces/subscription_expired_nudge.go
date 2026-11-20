package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// SubscriptionExpiredNudgeUseCase manages the daily automated reminder sent to
// sellers whose subscription has ALREADY expired (post-expiry nudge).
type SubscriptionExpiredNudgeUseCase interface {
	GetSettings(ctx context.Context) (domain.SubscriptionExpiredNudgeSettings, error)
	SaveSettings(ctx context.Context, settings domain.SubscriptionExpiredNudgeSettings) error

	GetTemplate(ctx context.Context) (domain.SubscriptionExpiredNudgeTemplate, error)
	SaveTemplate(ctx context.Context, tmpl domain.SubscriptionExpiredNudgeTemplate) error

	// RunSweep finds every seller with an expired subscription who hasn't been
	// nudged today, and sends them the template. Called once a day by the
	// in-process ticker and also triggerable on demand via the admin portal
	// "Run now" button. Idempotent: the sent ledger prevents double-sends.
	RunSweep(ctx context.Context) (SubscriptionExpiredNudgeSweepResult, error)
}

type SubscriptionExpiredNudgeSweepResult struct {
	Sent     int  `json:"sent"`
	Errors   int  `json:"errors"`
	Disabled bool `json:"disabled"`
}
