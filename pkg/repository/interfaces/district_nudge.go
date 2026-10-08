package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// DistrictNudgeRepository persists the district-growth-nudge template, its
// single schedule settings row, and the idempotency ledger of pairings
// already sent.
type DistrictNudgeRepository interface {
	GetTemplate(ctx context.Context) (domain.DistrictNudgeTemplate, error)
	SaveTemplate(ctx context.Context, tmpl domain.DistrictNudgeTemplate) error

	GetSettings(ctx context.Context) (domain.DistrictNudgeSettings, error)
	SaveSettings(ctx context.Context, settings domain.DistrictNudgeSettings) error

	// FindCandidates returns, for every active shop in state (case-insensitive;
	// blank matches every state) that has a stronger active peer within
	// radiusKm, that single strongest peer it hasn't already been nudged
	// about — one row per target shop, never the same pairing twice.
	FindCandidates(ctx context.Context, state string, radiusKm float64) ([]domain.DistrictNudgeCandidate, error)

	// MarkSent records a (target, top) pairing as delivered. Safe to call
	// more than once for the same pair (ON CONFLICT DO NOTHING).
	MarkSent(ctx context.Context, targetShopID, topShopID string) error
}
