package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// OnboardingNudgeUseCase manages the 4 editable onboarding-reminder
// templates + schedule settings, and runs the sweep that sends them.
type OnboardingNudgeUseCase interface {
	GetTemplates(ctx context.Context) ([]domain.OnboardingNudgeTemplate, error)
	SaveTemplate(ctx context.Context, tmpl domain.OnboardingNudgeTemplate) error

	GetSettings(ctx context.Context) (domain.OnboardingNudgeSettings, error)
	SaveSettings(ctx context.Context, settings domain.OnboardingNudgeSettings) error

	// RecordGoLive registers a shop's day-0 anchor. Called once, when a shop
	// first goes live (shop_status → active); a no-op on later re-approvals.
	RecordGoLive(ctx context.Context, shopID string) error

	// BackfillActiveShops anchors active shops that went live inside the
	// nudge window before this feature existed, at their approval time
	// (shop_details.updated_at), and marks already-elapsed slots as sent so
	// they resume at their next scheduled nudge instead of receiving a
	// catch-up burst. Idempotent: already-anchored shops are skipped. Returns
	// how many were newly anchored.
	BackfillActiveShops(ctx context.Context) (anchored int, err error)

	// RunSweep sends every due, not-yet-sent nudge across all shops still
	// within their window. Meant to be called repeatedly (e.g. every 15-30
	// minutes) by a scheduled job — see cmd/onboarding-nudges.
	RunSweep(ctx context.Context) (SweepResult, error)

	// GetStats reports how many nudges have actually been delivered (from
	// the sent ledger), independent of any single sweep run.
	GetStats(ctx context.Context) (domain.OnboardingNudgeStats, error)

	// GetShopCounts reports, per shop, how many onboarding nudges were
	// delivered and which type — for the shops list in admin-portal.
	GetShopCounts(ctx context.Context, shopIDs []string) (map[string]domain.ShopOnboardingNudgeCount, error)
}

// SweepResult summarizes one RunSweep call, for the cron binary to log and
// for admin-portal's manual "Run now" trigger to display.
type SweepResult struct {
	Sent     int  `json:"sent"`
	Skipped  int  `json:"skipped"`
	Errors   int  `json:"errors"`
	Disabled bool `json:"disabled"`
}
