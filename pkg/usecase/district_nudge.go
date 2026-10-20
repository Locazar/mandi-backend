package usecase

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoInterfaces "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	usecaseInterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

type districtNudgeUseCase struct {
	repo           repoInterfaces.DistrictNudgeRepository
	notificationUC usecaseInterfaces.NotificationUseCase
}

func NewDistrictNudgeUseCase(repo repoInterfaces.DistrictNudgeRepository, notificationUC usecaseInterfaces.NotificationUseCase) usecaseInterfaces.DistrictNudgeUseCase {
	return &districtNudgeUseCase{repo: repo, notificationUC: notificationUC}
}

func (uc *districtNudgeUseCase) GetTemplate(ctx context.Context) (domain.DistrictNudgeTemplate, error) {
	return uc.repo.GetTemplate(ctx)
}

func (uc *districtNudgeUseCase) SaveTemplate(ctx context.Context, tmpl domain.DistrictNudgeTemplate) error {
	if strings.TrimSpace(tmpl.Title) == "" || strings.TrimSpace(tmpl.Body) == "" {
		return fmt.Errorf("title and body are required")
	}
	return uc.repo.SaveTemplate(ctx, tmpl)
}

func (uc *districtNudgeUseCase) GetSettings(ctx context.Context) (domain.DistrictNudgeSettings, error) {
	return uc.repo.GetSettings(ctx)
}

func (uc *districtNudgeUseCase) SaveSettings(ctx context.Context, settings domain.DistrictNudgeSettings) error {
	if settings.RadiusKm <= 0 {
		return fmt.Errorf("radius_km must be greater than 0")
	}
	return uc.repo.SaveSettings(ctx, settings)
}

// RunSweep finds every target shop with a stronger, not-yet-nudged-about
// peer within radius and sends the template to each. Idempotent: MarkSent
// is the guard, so running this more than once (the daily ticker, plus a
// manual "Run now") never double-sends the same pairing.
func (uc *districtNudgeUseCase) RunSweep(ctx context.Context) (usecaseInterfaces.DistrictNudgeSweepResult, error) {
	settings, err := uc.repo.GetSettings(ctx)
	if err != nil {
		return usecaseInterfaces.DistrictNudgeSweepResult{}, fmt.Errorf("load settings: %w", err)
	}
	if !settings.Enabled {
		return usecaseInterfaces.DistrictNudgeSweepResult{Disabled: true}, nil
	}

	tmpl, err := uc.repo.GetTemplate(ctx)
	if err != nil {
		return usecaseInterfaces.DistrictNudgeSweepResult{}, fmt.Errorf("load template: %w", err)
	}
	if strings.TrimSpace(tmpl.Title) == "" || strings.TrimSpace(tmpl.Body) == "" {
		return usecaseInterfaces.DistrictNudgeSweepResult{}, nil
	}

	candidates, err := uc.repo.FindCandidates(ctx, settings.State, settings.RadiusKm)
	if err != nil {
		return usecaseInterfaces.DistrictNudgeSweepResult{}, fmt.Errorf("find candidates: %w", err)
	}

	result := usecaseInterfaces.DistrictNudgeSweepResult{}
	for _, candidate := range candidates {
		if err := uc.send(ctx, candidate, tmpl); err != nil {
			log.Printf("WARN [DistrictNudge sweep]: send target=%s top=%s failed: %v", candidate.TargetShopID, candidate.TopShopID, err)
			result.Errors++
			continue
		}
		if err := uc.repo.MarkSent(ctx, candidate.TargetShopID, candidate.TopShopID); err != nil {
			log.Printf("WARN [DistrictNudge sweep]: mark-sent failed target=%s top=%s: %v", candidate.TargetShopID, candidate.TopShopID, err)
		}
		result.Sent++
	}
	return result, nil
}

func (uc *districtNudgeUseCase) send(ctx context.Context, candidate domain.DistrictNudgeCandidate, tmpl domain.DistrictNudgeTemplate) error {
	targetName := candidate.TargetOwnerName
	if targetName == "" {
		targetName = candidate.TargetShopName
	}
	values := map[string]string{
		"district_name":        candidate.TargetCity,
		"top_shop_name":        candidate.TopShopName,
		"product_count":        strconv.Itoa(candidate.TopProductCount),
		"target_merchant_name": targetName,
		"top_shop_link":        utils.ShopPublicURL(candidate.TopShopID, candidate.TopShopName, candidate.TopShopCity),
	}
	title, body := tmpl.Title, tmpl.Body
	for token, value := range values {
		placeholder := "{{" + token + "}}"
		title = strings.ReplaceAll(title, placeholder, value)
		body = strings.ReplaceAll(body, placeholder, value)
	}

	data := map[string]string{"event_type": "district_nudge"}
	if tmpl.ImageURL != "" {
		data["image_url"] = tmpl.ImageURL
	}
	if tmpl.Route != "" {
		data["route"] = tmpl.Route
	}
	_, err := uc.notificationUC.SendPushNotification(ctx, request.SendPushRequest{
		OwnerID:   candidate.TargetShopID,
		OwnerType: "seller",
		Title:     title,
		Body:      body,
		Data:      data,
	})
	return err
}
