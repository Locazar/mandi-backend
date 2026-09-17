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
	return uc.repo.SaveSettings(ctx, settings)
}

func (uc *onboardingNudgeUseCase) RecordGoLive(ctx context.Context, shopID string) error {
	return uc.repo.RecordGoLiveOnce(ctx, shopID, time.Now())
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

	gap := time.Duration(settings.GapHours) * time.Hour
	result := usecaseInterfaces.SweepResult{}

	for _, candidate := range candidates {
		alreadySent := sentSlots[candidate.ShopID]
		for day := 1; day <= settings.DurationDays; day++ {
			dayOffset := time.Duration(day-1) * 24 * time.Hour
			for slot, key := range domain.NudgeTemplateOrder {
				due := candidate.GoLiveAt.Add(dayOffset).Add(time.Duration(slot) * gap)
				if due.After(now) {
					continue
				}
				if alreadySent[[2]int{day, slot}] {
					continue
				}
				tmpl, ok := templateByKey[key]
				if !ok {
					result.Skipped++
					continue
				}
				if err := uc.send(ctx, candidate, tmpl); err != nil {
					log.Printf("WARN [OnboardingNudge sweep]: send %s day=%d slot=%d shop=%s failed: %v", key, day, slot, candidate.ShopID, err)
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
	_, err := uc.notificationUC.SendPushNotification(ctx, request.SendPushRequest{
		OwnerID:   candidate.ShopID,
		OwnerType: "seller",
		Title:     tmpl.Title,
		Body:      body,
		Data:      data,
	})
	return err
}
