package interfaces

import (
	"context"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// OnboardingNudgeRepository persists the four editable nudge templates, the
// single schedule settings row, each shop's go-live anchor, and the
// idempotency ledger of nudges already sent.
type OnboardingNudgeRepository interface {
	GetTemplates(ctx context.Context) ([]domain.OnboardingNudgeTemplate, error)
	SaveTemplate(ctx context.Context, tmpl domain.OnboardingNudgeTemplate) error

	GetSettings(ctx context.Context) (domain.OnboardingNudgeSettings, error)
	SaveSettings(ctx context.Context, settings domain.OnboardingNudgeSettings) error

	// RecordGoLiveOnce inserts the shop's go-live anchor if it doesn't already
	// have one — a shop only ever gets one 7-day sequence, even if approved
	// more than once (e.g. suspended then re-approved). inserted reports
	// whether this call actually created the anchor (false if one already
	// existed).
	RecordGoLiveOnce(ctx context.Context, shopID string, at time.Time) (inserted bool, err error)

	// RecentlyLiveShops returns shops currently shop_status = 'active' whose
	// shop_details.updated_at (stamped by ApproveShop) is on/after since — the
	// backfill set for shops that went live before this feature existed, with
	// updated_at as the best available approximation of their go-live time.
	RecentlyLiveShops(ctx context.Context, since time.Time) ([]domain.OnboardingNudgeCandidate, error)

	// ActiveCandidates returns every shop still within its nudge window as of
	// now (go_live_at + durationDays >= now), joined with the shop_name/city
	// the "view your shop" nudge needs to build that shop's own link.
	ActiveCandidates(ctx context.Context, now time.Time, durationDays int) ([]domain.OnboardingNudgeCandidate, error)

	// SentSlotsByShop returns the (day, slot) pairs already sent for the
	// given shops, keyed by shop id, so the sweep can skip them in memory
	// instead of one query per candidate slot.
	SentSlotsByShop(ctx context.Context, shopIDs []string) (map[string]map[[2]int]bool, error)

	// MarkSent records a (shop, day, slot) as delivered. Safe to call more
	// than once for the same triple (ON CONFLICT DO NOTHING).
	MarkSent(ctx context.Context, shopID string, day, slot int) error

	// Stats aggregates the sent ledger — totals, per-slot, and per-day for
	// the last 7 days (dates in UTC, zero-filled). since is the start of
	// "today" used for the sent-today count.
	Stats(ctx context.Context, todayStart time.Time) (domain.OnboardingNudgeStats, error)
}
