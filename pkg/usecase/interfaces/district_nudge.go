package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// DistrictNudgeUseCase manages the district-growth-nudge template + schedule
// settings, and runs the sweep that finds and sends them.
type DistrictNudgeUseCase interface {
	GetTemplate(ctx context.Context) (domain.DistrictNudgeTemplate, error)
	SaveTemplate(ctx context.Context, tmpl domain.DistrictNudgeTemplate) error

	GetSettings(ctx context.Context) (domain.DistrictNudgeSettings, error)
	SaveSettings(ctx context.Context, settings domain.DistrictNudgeSettings) error

	// RunSweep finds every (target shop, strongest nearby peer) pairing due
	// for a nudge and sends it. Meant to be called once a day by a scheduled
	// job — see cmd/api/main.go's ticker — but also safe to call manually
	// (e.g. admin-portal's "Run now"): the ledger makes it idempotent.
	RunSweep(ctx context.Context) (DistrictNudgeSweepResult, error)
}

// DistrictNudgeSweepResult summarizes one RunSweep call, for the ticker to
// log and for admin-portal's manual "Run now" trigger to display.
type DistrictNudgeSweepResult struct {
	Sent     int  `json:"sent"`
	Errors   int  `json:"errors"`
	Disabled bool `json:"disabled"`
}
