package interfaces

import (
	"context"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type CustomerNudgeRepository interface {
	GetTemplates(ctx context.Context) ([]domain.CustomerNudgeTemplate, error)
	SaveTemplate(ctx context.Context, tmpl domain.CustomerNudgeTemplate) error
	GetSettings(ctx context.Context) (domain.CustomerNudgeSettings, error)
	SaveSettings(ctx context.Context, settings domain.CustomerNudgeSettings) error

	// Candidates returns non-blocked customers who signed up in [since, now].
	Candidates(ctx context.Context, since, now time.Time) ([]domain.CustomerNudgeCandidate, error)
	SentSlotsByUser(ctx context.Context, userIDs []string) (map[string]map[[2]int]bool, error)
	MarkSent(ctx context.Context, userID string, day, slot int, delivered bool) error
	Stats(ctx context.Context, todayStart time.Time) (domain.CustomerNudgeStats, error)
}
