package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// CustomerNudgeUseCase manages the editable customer onboarding templates +
// schedule and runs the sweep that sends them to recent signups.
type CustomerNudgeUseCase interface {
	GetTemplates(ctx context.Context) ([]domain.CustomerNudgeTemplate, error)
	SaveTemplate(ctx context.Context, tmpl domain.CustomerNudgeTemplate) error
	GetSettings(ctx context.Context) (domain.CustomerNudgeSettings, error)
	SaveSettings(ctx context.Context, settings domain.CustomerNudgeSettings) error
	RunSweep(ctx context.Context) (SweepResult, error)
	GetStats(ctx context.Context) (domain.CustomerNudgeStats, error)
}
