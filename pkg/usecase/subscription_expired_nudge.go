package usecase

import (
	"context"
	"fmt"
	"log"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoInterfaces "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	usecaseInterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type subscriptionExpiredNudgeUseCase struct {
	repo           repoInterfaces.SubscriptionExpiredNudgeRepository
	notificationUC usecaseInterfaces.NotificationUseCase
}

func NewSubscriptionExpiredNudgeUseCase(
	repo repoInterfaces.SubscriptionExpiredNudgeRepository,
	notificationUC usecaseInterfaces.NotificationUseCase,
) usecaseInterfaces.SubscriptionExpiredNudgeUseCase {
	return &subscriptionExpiredNudgeUseCase{repo: repo, notificationUC: notificationUC}
}

func (uc *subscriptionExpiredNudgeUseCase) GetSettings(ctx context.Context) (domain.SubscriptionExpiredNudgeSettings, error) {
	return uc.repo.GetSettings(ctx)
}

func (uc *subscriptionExpiredNudgeUseCase) SaveSettings(ctx context.Context, settings domain.SubscriptionExpiredNudgeSettings) error {
	return uc.repo.SaveSettings(ctx, settings)
}

func (uc *subscriptionExpiredNudgeUseCase) GetTemplate(ctx context.Context) (domain.SubscriptionExpiredNudgeTemplate, error) {
	return uc.repo.GetTemplate(ctx)
}

func (uc *subscriptionExpiredNudgeUseCase) SaveTemplate(ctx context.Context, tmpl domain.SubscriptionExpiredNudgeTemplate) error {
	if strings.TrimSpace(tmpl.Title) == "" || strings.TrimSpace(tmpl.Body) == "" {
		return fmt.Errorf("title and body are required")
	}
	return uc.repo.SaveTemplate(ctx, tmpl)
}

func (uc *subscriptionExpiredNudgeUseCase) RunSweep(ctx context.Context) (usecaseInterfaces.SubscriptionExpiredNudgeSweepResult, error) {
	settings, err := uc.repo.GetSettings(ctx)
	if err != nil {
		return usecaseInterfaces.SubscriptionExpiredNudgeSweepResult{}, fmt.Errorf("load settings: %w", err)
	}
	if !settings.Enabled {
		return usecaseInterfaces.SubscriptionExpiredNudgeSweepResult{Disabled: true}, nil
	}

	tmpl, err := uc.repo.GetTemplate(ctx)
	if err != nil {
		return usecaseInterfaces.SubscriptionExpiredNudgeSweepResult{}, fmt.Errorf("load template: %w", err)
	}
	if strings.TrimSpace(tmpl.Title) == "" || strings.TrimSpace(tmpl.Body) == "" {
		return usecaseInterfaces.SubscriptionExpiredNudgeSweepResult{}, nil
	}

	candidates, err := uc.repo.FindAlreadyExpired(ctx)
	if err != nil {
		return usecaseInterfaces.SubscriptionExpiredNudgeSweepResult{}, fmt.Errorf("find candidates: %w", err)
	}

	result := usecaseInterfaces.SubscriptionExpiredNudgeSweepResult{}
	for _, c := range candidates {
		// Round up: a subscription that expired 25 hours ago is "2 days expired".
		hoursExpired := time.Since(c.EndDate).Hours()
		daysExpired := int(math.Ceil(hoursExpired / 24))
		if daysExpired < 1 {
			daysExpired = 1
		}
		title := strings.ReplaceAll(tmpl.Title, "{{days_expired}}", strconv.Itoa(daysExpired))
		body := strings.ReplaceAll(tmpl.Body, "{{days_expired}}", strconv.Itoa(daysExpired))

		data := map[string]string{"event_type": "subscription_expired_nudge"}
		if tmpl.ImageURL != "" {
			data["image_url"] = tmpl.ImageURL
		}
		if tmpl.Route != "" {
			data["route"] = tmpl.Route
		}

		_, sendErr := uc.notificationUC.SendPushNotification(ctx, request.SendPushRequest{
			OwnerID:   c.UserID,
			OwnerType: "seller",
			Title:     title,
			Body:      body,
			Data:      data,
		})
		if sendErr != nil {
			log.Printf("WARN [SubscriptionExpiredNudge sweep]: send user=%s failed: %v", c.UserID, sendErr)
			result.Errors++
			continue
		}
		if markErr := uc.repo.MarkSent(ctx, c.UserID); markErr != nil {
			log.Printf("WARN [SubscriptionExpiredNudge sweep]: mark-sent failed user=%s: %v", c.UserID, markErr)
		}
		result.Sent++
	}
	return result, nil
}
