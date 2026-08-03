package usecase

import (
	"context"
	"errors"
	"log"
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
	subRepo     repoIface.SubscriptionRepository
	userRepo    repoIface.UserRepository
	invoiceRepo repoIface.InvoiceRepository
}

func NewSubscriptionUseCase(
	subRepo repoIface.SubscriptionRepository,
	userRepo repoIface.UserRepository,
	invoiceRepo repoIface.InvoiceRepository,
) service.SubscriptionUseCase {
	return &subscriptionUseCase{
		subRepo:     subRepo,
		userRepo:    userRepo,
		invoiceRepo: invoiceRepo,
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
			Description:  p.Description,
			Features:     []string(p.Features),
		}
	}
	return result, nil
}

// GetBillingHistory returns the user's paid subscription payments (most recent
// first). GST is read from the per-order snapshot stored on each order
// (gst_rate_basis_points + gst_amount), so historical rates stay accurate.
func (uc *subscriptionUseCase) GetBillingHistory(ctx context.Context, userID string) (response.BillingHistoryResponse, error) {
	orders, err := uc.subRepo.FindPaidOrdersByUserID(ctx, userID)
	if err != nil {
		return response.BillingHistoryResponse{}, err
	}

	orderIDs := make([]string, 0, len(orders))
	for _, o := range orders {
		orderIDs = append(orderIDs, o.ID)
	}
	// One query for all invoices rather than N — billing history is a list view.
	invoicesByOrder, err := uc.invoiceRepo.FindInvoiceNumbersByOrderIDs(ctx, orderIDs)
	if err != nil {
		// Billing history is still useful without download links.
		log.Printf("[BILLING_HISTORY] invoice lookup failed for user %s: %v", userID, err)
		invoicesByOrder = map[string]domain.Invoice{}
	}

	payments := make([]response.BillingPaymentResponse, 0, len(orders))
	for _, o := range orders {
		amountMinor := o.Price.AmountMinor
		gstAmountMinor := o.GSTAmount.AmountMinor
		baseAmountMinor := amountMinor - gstAmountMinor

		paidAt := ""
		if o.PaidAt != nil {
			paidAt = o.PaidAt.UTC().Format(time.RFC3339)
		}

		invoiceID, invoiceNumber := "", ""
		if inv, ok := invoicesByOrder[o.ID]; ok {
			invoiceID, invoiceNumber = inv.ID, inv.InvoiceNumber
		}

		payments = append(payments, response.BillingPaymentResponse{
			ID:                 o.ID,
			PlanName:           o.PlanName,
			AmountMinor:        amountMinor,
			Currency:           o.Price.Currency,
			BaseAmountMinor:    baseAmountMinor,
			GSTAmount:          gstAmountMinor,
			GSTRateBasisPoints: o.GSTRateBasisPoints,
			GSTInclusive:       true,
			Status:             string(o.Status),
			RazorpayPaymentID:  o.RazorpayPaymentID,
			RazorpayOrderID:    o.RazorpayOrderID,
			PaidAt:             paidAt,
			InvoiceID:          invoiceID,
			InvoiceNumber:      invoiceNumber,
		})
	}

	resp := response.BillingHistoryResponse{
		GSTInclusive: true,
		LastPayment:  nil,
		Payments:     payments,
	}
	if len(payments) > 0 {
		resp.LastPayment = &payments[0]
	}
	return resp, nil
}
