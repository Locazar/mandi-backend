package usecase

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoIface "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	service "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"gorm.io/gorm"
)

const freeTrialPlanName = "Free Trial"

type subscriptionUseCase struct {
	subRepo  repoIface.SubscriptionRepository
	userRepo repoIface.UserRepository
}

func NewSubscriptionUseCase(
	subRepo repoIface.SubscriptionRepository,
	userRepo repoIface.UserRepository,
) service.SubscriptionUseCase {
	return &subscriptionUseCase{
		subRepo:  subRepo,
		userRepo: userRepo,
	}
}

func (uc *subscriptionUseCase) GetSubscriptionStatus(ctx context.Context, userID string) (response.SubscriptionStatusResponse, error) {
	// Check for active subscription
	sub, err := uc.subRepo.FindActiveSubscriptionByUserID(ctx, userID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return response.SubscriptionStatusResponse{}, err
	}

	if err == nil {
		// Active subscription found
		plan, _ := uc.subRepo.FindSubscriptionPlanByID(ctx, sub.PlanID)
		daysRemaining := int(math.Ceil(time.Until(sub.EndDate).Hours() / 24))
		if daysRemaining < 0 {
			daysRemaining = 0
		}

		status := "subscribed"
		if sub.IsTrial {
			status = "trial_active"
		}

		return response.SubscriptionStatusResponse{
			Status:        status,
			PlanName:      plan.Name,
			StartDate:     sub.StartDate.Format("2006-01-02"),
			EndDate:       sub.EndDate.Format("2006-01-02"),
			DaysRemaining: daysRemaining,
			IsTrial:       sub.IsTrial,
		}, nil
	}

	// No active subscription — check trial eligibility
	trialUsed, err := uc.userRepo.IsTrialUsed(ctx, userID)
	if err != nil {
		return response.SubscriptionStatusResponse{}, err
	}

	if !trialUsed {
		return response.SubscriptionStatusResponse{
			Status: "trial_eligible",
		}, nil
	}

	// Trial was used but no active sub — check if trial expired or just no subscription
	return response.SubscriptionStatusResponse{
		Status: "trial_expired",
	}, nil
}

func (uc *subscriptionUseCase) StartTrial(ctx context.Context, userID string) (response.SubscriptionStatusResponse, error) {
	// 1. Check trial already used
	trialUsed, err := uc.userRepo.IsTrialUsed(ctx, userID)
	if err != nil {
		return response.SubscriptionStatusResponse{}, err
	}
	if trialUsed {
		return response.SubscriptionStatusResponse{}, ErrTrialAlreadyUsed
	}

	// 2. Check no active subscription
	_, err = uc.subRepo.FindActiveSubscriptionByUserID(ctx, userID)
	if err == nil {
		return response.SubscriptionStatusResponse{}, ErrActiveSubscriptionExists
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return response.SubscriptionStatusResponse{}, err
	}

	// 3. Find the Free Trial plan
	plan, err := uc.subRepo.FindSubscriptionPlanByName(ctx, freeTrialPlanName)
	if err != nil {
		return response.SubscriptionStatusResponse{}, ErrTrialPlanNotFound
	}

	// 4. Atomic: set TrialUsed + activate trial subscription
	now := time.Now()
	endDate := now.AddDate(0, 0, int(plan.DurationDays))

	err = uc.subRepo.Transaction(func(trxRepo repoIface.SubscriptionRepository) error {
		if err := uc.userRepo.UpdateTrialUsed(ctx, userID); err != nil {
			return err
		}
		return trxRepo.ActivateSubscription(ctx, domain.UserSubscription{
			UserID:    userID,
			PlanID:    plan.ID,
			IsTrial:   true,
			StartDate: now,
			EndDate:   endDate,
			IsActive:  true,
		})
	})
	if err != nil {
		return response.SubscriptionStatusResponse{}, err
	}

	daysRemaining := int(math.Ceil(time.Until(endDate).Hours() / 24))

	return response.SubscriptionStatusResponse{
		Status:        "trial_active",
		PlanName:      plan.Name,
		StartDate:     now.Format("2006-01-02"),
		EndDate:       endDate.Format("2006-01-02"),
		DaysRemaining: daysRemaining,
		IsTrial:       true,
	}, nil
}

func (uc *subscriptionUseCase) GetPaidPlans(ctx context.Context) ([]response.SubscriptionPlanResponse, error) {
	plans, err := uc.subRepo.FindPaidSubscriptionPlans(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]response.SubscriptionPlanResponse, len(plans))
	for i, p := range plans {
		result[i] = response.SubscriptionPlanResponse{
			ID:           p.ID,
			Name:         p.Name,
			PriceMonthly: uint(p.PriceMonthly.AmountMinor),
			DurationDays: p.DurationDays,
		}
	}
	return result, nil
}
