package usecase

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoInterfaces "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	usecaseInterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

type onboardingNudgeUseCase struct {
	repo           repoInterfaces.OnboardingNudgeRepository
	notificationUC usecaseInterfaces.NotificationUseCase
}

func NewOnboardingNudgeUseCase(repo repoInterfaces.OnboardingNudgeRepository, notificationUC usecaseInterfaces.NotificationUseCase) usecaseInterfaces.OnboardingNudgeUseCase {
	return &onboardingNudgeUseCase{repo: repo, notificationUC: notificationUC}
}

// istLocation is loaded once at startup. Falls back to a fixed offset if the
// timezone database is unavailable (e.g. a minimal Docker image without tzdata).
var istLocation = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		loc = time.FixedZone("IST", 5*60*60+30*60)
	}
	return loc
}()

// nudgeDueAt shifts a nominal due instant to the configured send-hour in IST.
// This keeps the DUE DATE (IST calendar date of the nominal time) but replaces
// the time-of-day with sendHour:00:00 IST, so every nudge fires at a consistent
// hour regardless of when the shop went live.
func nudgeDueAt(nominal time.Time, sendHour int) time.Time {
	t := nominal.In(istLocation)
	return time.Date(t.Year(), t.Month(), t.Day(), sendHour, 0, 0, 0, istLocation)
}

var validTemplateKeys = map[string]bool{
	domain.NudgeAddProducts:   true,
	domain.NudgeUpdatePhoto:   true,
	domain.NudgeUpdateAddress: true,
	domain.NudgeViewShop:      true,
}

func (uc *onboardingNudgeUseCase) GetTemplates(ctx context.Context) ([]domain.OnboardingNudgeTemplate, error) {
	return uc.repo.GetTemplates(ctx)
}

func (uc *onboardingNudgeUseCase) SaveTemplate(ctx context.Context, tmpl domain.OnboardingNudgeTemplate) error {
	if !validTemplateKeys[tmpl.Key] {
		return fmt.Errorf("unknown template key %q", tmpl.Key)
	}
	if strings.TrimSpace(tmpl.Title) == "" || strings.TrimSpace(tmpl.Body) == "" {
		return fmt.Errorf("title and body are required")
	}
	return uc.repo.SaveTemplate(ctx, tmpl)
}

func (uc *onboardingNudgeUseCase) GetSettings(ctx context.Context) (domain.OnboardingNudgeSettings, error) {
	return uc.repo.GetSettings(ctx)
}

func (uc *onboardingNudgeUseCase) SaveSettings(ctx context.Context, settings domain.OnboardingNudgeSettings) error {
	if settings.GapHours < 1 {
		return fmt.Errorf("gap_hours must be at least 1")
	}
	if settings.DurationDays < 1 {
		return fmt.Errorf("duration_days must be at least 1")
	}
	if settings.WeekGapDays < 0 {
		return fmt.Errorf("week_gap_days must be 0 or greater")
	}
	if settings.SendHour < 0 || settings.SendHour > 23 {
		return fmt.Errorf("send_hour must be between 0 and 23")
	}
	return uc.repo.SaveSettings(ctx, settings)
}

func (uc *onboardingNudgeUseCase) RecordGoLive(ctx context.Context, shopID string) error {
	_, err := uc.repo.RecordGoLiveOnce(ctx, shopID, time.Now())
	return err
}

// BackfillActiveShops anchors shops that went live inside the nudge window
// before this feature existed, at their real approval time (updated_at), so
// they resume the sequence where it would be today. Slots already past due
// are marked sent rather than delivered — otherwise a shop live for 5 days
// would receive every missed nudge at once.
func (uc *onboardingNudgeUseCase) BackfillActiveShops(ctx context.Context) (int, error) {
	settings, err := uc.repo.GetSettings(ctx)
	if err != nil {
		return 0, fmt.Errorf("load settings: %w", err)
	}
	now := time.Now()
	shops, err := uc.repo.RecentlyLiveShops(ctx, now.AddDate(0, 0, -settings.DurationDays))
	if err != nil {
		return 0, fmt.Errorf("load recently live shops: %w", err)
	}

	numSlots := len(domain.NudgeTemplateOrder)
	slotGap := time.Duration(settings.GapHours) * time.Hour
	weekGap := time.Duration(settings.WeekGapDays) * 24 * time.Hour
	// Round duration = gap from one round's START to the next. Using (numSlots-1)
	// means the break between the LAST slot of round N and the FIRST slot of
	// round N+1 equals exactly weekGap — matching the user expectation of "1 week
	// gap after all 4 fire". Using numSlots would add one extra slotGap to that
	// break, giving 9 days instead of 7 for the default 2-day/7-day settings.
	roundDur := time.Duration(numSlots-1)*slotGap + weekGap
	totalDur := time.Duration(settings.DurationDays) * 24 * time.Hour

	anchored := 0
	for _, shop := range shops {
		inserted, err := uc.repo.RecordGoLiveOnce(ctx, shop.ShopID, shop.GoLiveAt)
		if err != nil {
			log.Printf("WARN [OnboardingNudge backfill]: anchor failed for %s: %v", shop.ShopID, err)
			continue
		}
		if !inserted {
			continue
		}
		anchored++
		deadline := shop.GoLiveAt.Add(totalDur)
		maxRounds := settings.DurationDays + 1
		for round := 0; round <= maxRounds; round++ {
			roundStart := shop.GoLiveAt.Add(time.Duration(round) * roundDur)
			if roundStart.After(deadline) {
				break
			}
			for slot := 0; slot < numSlots; slot++ {
				nominal := roundStart.Add(time.Duration(slot) * slotGap)
				due := nudgeDueAt(nominal, settings.SendHour)
				if due.After(now) || nominal.After(deadline) {
					continue
				}
				day := round*numSlots + slot + 1
				if err := uc.repo.MarkSent(ctx, shop.ShopID, day, slot); err != nil {
					log.Printf("WARN [OnboardingNudge backfill]: pre-mark failed shop=%s day=%d slot=%d: %v", shop.ShopID, day, slot, err)
				}
			}
		}
	}
	return anchored, nil
}

// RunSweep checks every (day, slot) combination for every shop still inside
// its window and sends whatever is due and not yet sent. Checking the whole
// window rather than just "today" makes it safe to run on any cadence —
// downtime or a slow sweep just catches up on the next run, nothing is lost
// or double-sent (MarkSent is the idempotency guard).
func (uc *onboardingNudgeUseCase) RunSweep(ctx context.Context) (usecaseInterfaces.SweepResult, error) {
	settings, err := uc.repo.GetSettings(ctx)
	if err != nil {
		return usecaseInterfaces.SweepResult{}, fmt.Errorf("load settings: %w", err)
	}
	if !settings.Enabled {
		return usecaseInterfaces.SweepResult{Disabled: true}, nil
	}

	templates, err := uc.repo.GetTemplates(ctx)
	if err != nil {
		return usecaseInterfaces.SweepResult{}, fmt.Errorf("load templates: %w", err)
	}
	templateByKey := make(map[string]domain.OnboardingNudgeTemplate, len(templates))
	for _, t := range templates {
		templateByKey[t.Key] = t
	}

	now := time.Now()
	candidates, err := uc.repo.ActiveCandidates(ctx, now, settings.DurationDays)
	if err != nil {
		return usecaseInterfaces.SweepResult{}, fmt.Errorf("load candidates: %w", err)
	}
	if len(candidates) == 0 {
		return usecaseInterfaces.SweepResult{}, nil
	}

	shopIDs := make([]string, len(candidates))
	for i, c := range candidates {
		shopIDs[i] = c.ShopID
	}
	sentSlots, err := uc.repo.SentSlotsByShop(ctx, shopIDs)
	if err != nil {
		return usecaseInterfaces.SweepResult{}, fmt.Errorf("load sent ledger: %w", err)
	}

	numSlots := len(domain.NudgeTemplateOrder)
	slotGap := time.Duration(settings.GapHours) * time.Hour
	weekGap := time.Duration(settings.WeekGapDays) * 24 * time.Hour
	roundDur := time.Duration(numSlots-1)*slotGap + weekGap
	totalDur := time.Duration(settings.DurationDays) * 24 * time.Hour
	maxRounds := settings.DurationDays + 1

	result := usecaseInterfaces.SweepResult{}

	for _, candidate := range candidates {
		alreadySent := sentSlots[candidate.ShopID]
		deadline := candidate.GoLiveAt.Add(totalDur)
		for round := 0; round <= maxRounds; round++ {
			roundStart := candidate.GoLiveAt.Add(time.Duration(round) * roundDur)
			if roundStart.After(deadline) {
				break
			}
			for slot, key := range domain.NudgeTemplateOrder {
				nominal := roundStart.Add(time.Duration(slot) * slotGap)
				due := nudgeDueAt(nominal, settings.SendHour)
				if due.After(now) || nominal.After(deadline) {
					continue
				}
				day := round*numSlots + slot + 1
				if alreadySent[[2]int{day, slot}] {
					continue
				}
				tmpl, ok := templateByKey[key]
				if !ok {
					result.Skipped++
					continue
				}
				if err := uc.send(ctx, candidate, tmpl); err != nil {
					log.Printf("WARN [OnboardingNudge sweep]: send %s round=%d slot=%d shop=%s failed: %v", key, round, slot, candidate.ShopID, err)
					result.Errors++
					continue
				}
				if err := uc.repo.MarkSent(ctx, candidate.ShopID, day, slot); err != nil {
					log.Printf("WARN [OnboardingNudge sweep]: mark-sent failed shop=%s day=%d slot=%d: %v", candidate.ShopID, day, slot, err)
				}
				result.Sent++
			}
		}
	}
	return result, nil
}

func (uc *onboardingNudgeUseCase) send(ctx context.Context, candidate domain.OnboardingNudgeCandidate, tmpl domain.OnboardingNudgeTemplate) error {
	body := tmpl.Body
	if strings.Contains(body, "{{shop_link}}") {
		link := utils.ShopPublicURL(candidate.ShopID, candidate.ShopName, candidate.City)
		body = strings.ReplaceAll(body, "{{shop_link}}", link)
	}
	data := map[string]string{"event_type": "onboarding_nudge"}
	if tmpl.ImageURL != "" {
		data["image_url"] = tmpl.ImageURL
	}
	if tmpl.Route != "" {
		data["route"] = tmpl.Route
	}
	_, err := uc.notificationUC.SendPushNotification(ctx, request.SendPushRequest{
		OwnerID:   candidate.ShopID,
		OwnerType: "seller",
		Title:     tmpl.Title,
		Body:      body,
		Data:      data,
	})
	return err
}

func (uc *onboardingNudgeUseCase) GetStats(ctx context.Context) (domain.OnboardingNudgeStats, error) {
	now := time.Now().UTC()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return uc.repo.Stats(ctx, todayStart)
}

func (uc *onboardingNudgeUseCase) GetShopCounts(ctx context.Context, shopIDs []string) (map[string]domain.ShopOnboardingNudgeCount, error) {
	return uc.repo.SentCountsByShop(ctx, shopIDs)
}
