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

	// RunSweep sends every due, not-yet-sent nudge across all shops still
	// within their window. Meant to be called repeatedly (e.g. every 15-30
	// minutes) by a scheduled job — see cmd/onboarding-nudges.
	RunSweep(ctx context.Context) (SweepResult, error)
}

// SweepResult summarizes one RunSweep call, for the cron binary to log.
type SweepResult struct {
	Sent     int
	Skipped  int
	Errors   int
	Disabled bool
}
