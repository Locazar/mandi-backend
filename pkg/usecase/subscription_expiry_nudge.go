package usecase

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoInterfaces "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	usecaseInterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type subscriptionExpiryNudgeUseCase struct {
	repo           repoInterfaces.SubscriptionExpiryNudgeRepository
	notificationUC usecaseInterfaces.NotificationUseCase
}

func NewSubscriptionExpiryNudgeUseCase(
	repo repoInterfaces.SubscriptionExpiryNudgeRepository,
	notificationUC usecaseInterfaces.NotificationUseCase,
) usecaseInterfaces.SubscriptionExpiryNudgeUseCase {
	return &subscriptionExpiryNudgeUseCase{repo: repo, notificationUC: notificationUC}
}

func (uc *subscriptionExpiryNudgeUseCase) GetSettings(ctx context.Context) (domain.SubscriptionExpiryNudgeSettings, error) {
	return uc.repo.GetSettings(ctx)
}

func (uc *subscriptionExpiryNudgeUseCase) SaveSettings(ctx context.Context, settings domain.SubscriptionExpiryNudgeSettings) error {
	if settings.DaysBeforeExpiry <= 0 {
		return fmt.Errorf("days_before_expiry must be greater than 0")
	}
	return uc.repo.SaveSettings(ctx, settings)
}

func (uc *subscriptionExpiryNudgeUseCase) GetTemplate(ctx context.Context) (domain.SubscriptionExpiryNudgeTemplate, error) {
	return uc.repo.GetTemplate(ctx)
}

func (uc *subscriptionExpiryNudgeUseCase) SaveTemplate(ctx context.Context, tmpl domain.SubscriptionExpiryNudgeTemplate) error {
	if strings.TrimSpace(tmpl.Title) == "" || strings.TrimSpace(tmpl.Body) == "" {
		return fmt.Errorf("title and body are required")
	}
	return uc.repo.SaveTemplate(ctx, tmpl)
}

func (uc *subscriptionExpiryNudgeUseCase) RunSweep(ctx context.Context) (usecaseInterfaces.SubscriptionExpiryNudgeSweepResult, error) {
	settings, err := uc.repo.GetSettings(ctx)
	if err != nil {
		return usecaseInterfaces.SubscriptionExpiryNudgeSweepResult{}, fmt.Errorf("load settings: %w", err)
	}
	if !settings.Enabled {
		return usecaseInterfaces.SubscriptionExpiryNudgeSweepResult{Disabled: true}, nil
	}

	tmpl, err := uc.repo.GetTemplate(ctx)
	if err != nil {
		return usecaseInterfaces.SubscriptionExpiryNudgeSweepResult{}, fmt.Errorf("load template: %w", err)
	}
	if strings.TrimSpace(tmpl.Title) == "" || strings.TrimSpace(tmpl.Body) == "" {
		return usecaseInterfaces.SubscriptionExpiryNudgeSweepResult{}, nil
	}

	candidates, err := uc.repo.FindExpiringSoon(ctx, settings.DaysBeforeExpiry)
	if err != nil {
		return usecaseInterfaces.SubscriptionExpiryNudgeSweepResult{}, fmt.Errorf("find candidates: %w", err)
	}

	result := usecaseInterfaces.SubscriptionExpiryNudgeSweepResult{}
	for _, c := range candidates {
		daysLeft := int(time.Until(c.EndDate).Hours()/24) + 1
		if daysLeft < 1 {
			daysLeft = 1
		}
		title := strings.ReplaceAll(tmpl.Title, "{{days_left}}", strconv.Itoa(daysLeft))
		body := strings.ReplaceAll(tmpl.Body, "{{days_left}}", strconv.Itoa(daysLeft))

		data := map[string]string{"event_type": "subscription_expiry_nudge"}
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
			log.Printf("WARN [SubscriptionExpiryNudge sweep]: send user=%s failed: %v", c.UserID, sendErr)
			result.Errors++
			continue
		}
		if markErr := uc.repo.MarkSent(ctx, c.UserID); markErr != nil {
			log.Printf("WARN [SubscriptionExpiryNudge sweep]: mark-sent failed user=%s: %v", c.UserID, markErr)
		}
		result.Sent++
	}
	return result, nil
}
