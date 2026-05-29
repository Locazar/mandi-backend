package usecase

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/razorpay/razorpay-go"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoIface "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	service "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"gorm.io/gorm"
)

const subscriptionOrderExpiryMinutes = 30

type subscriptionPaymentUseCase struct {
	subRepo     repoIface.SubscriptionRepository
	paymentRepo repoIface.PaymentRepository
	userRepo    repoIface.UserRepository
	config      config.Config
}

func NewSubscriptionPaymentUseCase(
	subRepo repoIface.SubscriptionRepository,
	paymentRepo repoIface.PaymentRepository,
	userRepo repoIface.UserRepository,
	cfg config.Config,
) service.SubscriptionPaymentUseCase {
	return &subscriptionPaymentUseCase{
		subRepo:     subRepo,
		paymentRepo: paymentRepo,
		userRepo:    userRepo,
		config:      cfg,
	}
}

// CreateSubscriptionOrder creates a Razorpay order for a subscription plan.
func (uc *subscriptionPaymentUseCase) CreateSubscriptionOrder(ctx context.Context, userID uint, req request.CreateSubscriptionOrderRequest) (response.SubscriptionOrderResponse, error) {
	// 1. Look up plan
	plan, err := uc.subRepo.FindSubscriptionPlanByID(ctx, req.PlanID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.SubscriptionOrderResponse{}, ErrSubscriptionPlanNotFound
		}
		return response.SubscriptionOrderResponse{}, fmt.Errorf("find plan: %w", err)
	}
	if !plan.IsActive {
		return response.SubscriptionOrderResponse{}, ErrSubscriptionPlanInactive
	}

	// 2. Check Razorpay not blocked
	pm, err := uc.paymentRepo.FindPaymentMethodByType(ctx, domain.RazopayPayment)
	if err != nil {
		return response.SubscriptionOrderResponse{}, fmt.Errorf("find payment method: %w", err)
	}
	if pm.BlockStatus {
		return response.SubscriptionOrderResponse{}, ErrBlockedPayment
	}

	// 3. Check no active subscription (allow upgrade from trial)
	activeSub, err := uc.subRepo.FindActiveSubscriptionByUserID(ctx, userID)
	if err == nil {
		if !activeSub.IsTrial {
			return response.SubscriptionOrderResponse{}, ErrActiveSubscriptionExists
		}
		// Trial active — allow upgrade to paid plan
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return response.SubscriptionOrderResponse{}, fmt.Errorf("check active sub: %w", err)
	}

	// 4. Get user details for prefill
	user, err := uc.userRepo.FindUserByUserID(ctx, userID)
	if err != nil {
		return response.SubscriptionOrderResponse{}, fmt.Errorf("find user: %w", err)
	}

	// 5. Create Razorpay order
	// PriceMonthly is stored in minor units (paise), which is exactly what Razorpay expects.
	amountPaise := plan.PriceMonthly.AmountMinor
	client := razorpay.NewClient(uc.config.RazorPayKey, uc.config.RazorPaySecret)
	receipt := fmt.Sprintf("sub_%d_%d_%d", userID, plan.ID, time.Now().Unix())

	rzpData := map[string]interface{}{
		"amount":   amountPaise,
		"currency": "INR",
		"receipt":  receipt,
	}

	rzpRes, err := client.Order.Create(rzpData, nil)
	if err != nil {
		return response.SubscriptionOrderResponse{}, fmt.Errorf("razorpay order create: %w", err)
	}

	rzpOrderID, _ := rzpRes["id"].(string)

	// 6. Persist subscription order
	subOrder := domain.SubscriptionOrder{
		UserID:          userID,
		PlanID:          plan.ID,
		Price:           plan.PriceMonthly,
		RazorpayOrderID: rzpOrderID,
		Status:          domain.SubStatusCreated,
	}

	subOrder, err = uc.subRepo.CreateSubscriptionOrder(ctx, subOrder)
	if err != nil {
		return response.SubscriptionOrderResponse{}, fmt.Errorf("persist order: %w", err)
	}

	// 7. Return checkout data
	return response.SubscriptionOrderResponse{
		OrderID:     rzpOrderID,
		KeyID:       uc.config.RazorPayKey,
		Amount:      uint(amountPaise),
		Currency:    "INR",
		ShopOrderID: subOrder.ID,
		Prefill: response.SubscriptionPrefill{
			Email:   user.Email,
			Contact: user.Phone,
		},
	}, nil
}

// VerifySubscriptionPayment performs the 4-step verification.
func (uc *subscriptionPaymentUseCase) VerifySubscriptionPayment(ctx context.Context, userID uint, req request.VerifySubscriptionPaymentRequest) (response.SubscriptionVerificationResponse, error) {
	// Step A: Signature verification
	if err := uc.verifySignature(req.OrderID, req.PaymentID, req.Signature, uc.config.RazorPaySecret); err != nil {
		return response.SubscriptionVerificationResponse{}, err
	}

	// Step B: Order ownership validation
	order, err := uc.subRepo.FindSubscriptionOrderByRazorpayOrderID(ctx, req.OrderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.SubscriptionVerificationResponse{}, ErrSubscriptionOrderNotFound
		}
		return response.SubscriptionVerificationResponse{}, fmt.Errorf("find order: %w", err)
	}
	if order.UserID != userID {
		return response.SubscriptionVerificationResponse{}, ErrSubscriptionOrderNotOwned
	}

	// Step D (early): Idempotency — already paid?
	if order.Status == domain.SubStatusPaid {
		plan, _ := uc.subRepo.FindSubscriptionPlanByID(ctx, order.PlanID)
		return response.SubscriptionVerificationResponse{
			Status: "success",
			Plan:   plan.Name,
		}, nil
	}

	if order.Status != domain.SubStatusCreated {
		return response.SubscriptionVerificationResponse{}, ErrSubscriptionOrderNotCreated
	}
	if time.Since(order.CreatedAt) > subscriptionOrderExpiryMinutes*time.Minute {
		return response.SubscriptionVerificationResponse{}, ErrSubscriptionOrderExpired
	}

	// Step C: Payment cross-check via Razorpay API
	if err := uc.crossCheckPayment(req.PaymentID, req.OrderID, uint(order.Price.AmountMinor)); err != nil {
		return response.SubscriptionVerificationResponse{}, err
	}

	// Step D: Idempotency by payment_id
	_, err = uc.subRepo.FindSubscriptionOrderByRazorpayPaymentID(ctx, req.PaymentID)
	if err == nil {
		plan, _ := uc.subRepo.FindSubscriptionPlanByID(ctx, order.PlanID)
		return response.SubscriptionVerificationResponse{
			Status: "success",
			Plan:   plan.Name,
		}, nil
	}

	// Atomic: mark paid + activate
	plan, err := uc.subRepo.FindSubscriptionPlanByID(ctx, order.PlanID)
	if err != nil {
		return response.SubscriptionVerificationResponse{}, fmt.Errorf("find plan: %w", err)
	}

	orderID := order.ID
	err = uc.subRepo.Transaction(func(trxRepo repoIface.SubscriptionRepository) error {
		if err := trxRepo.UpdateSubscriptionOrderToPaid(ctx, order.ID, req.PaymentID); err != nil {
			return err
		}
		// Deactivate any active trial before activating paid subscription
		if err := trxRepo.DeactivateTrialSubscription(ctx, userID); err != nil {
			return err
		}
		now := time.Now()
		return trxRepo.ActivateSubscription(ctx, domain.UserSubscription{
			UserID:              userID,
			PlanID:              order.PlanID,
			SubscriptionOrderID: &orderID,
			StartDate:           now,
			EndDate:             now.AddDate(0, 0, int(plan.DurationDays)),
			IsActive:            true,
		})
	})
	if err != nil {
		return response.SubscriptionVerificationResponse{}, fmt.Errorf("activate subscription: %w", err)
	}

	return response.SubscriptionVerificationResponse{
		Status:      "success",
		Plan:        plan.Name,
		ActivatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// HandlePaymentFailure logs a payment failure without changing order state.
func (uc *subscriptionPaymentUseCase) HandlePaymentFailure(ctx context.Context, userID uint, req request.PaymentFailureRequest) error {
	log.Printf("[SUBSCRIPTION_PAYMENT_FAILURE] user_id=%d order_id=%s code=%d message=%s time=%s",
		userID, req.OrderID, req.Code, req.Message, time.Now().Format(time.RFC3339))
	return nil
}

// HandleWebhook processes Razorpay webhook events.
func (uc *subscriptionPaymentUseCase) HandleWebhook(ctx context.Context, signature string, rawBody []byte) error {
	// 1. Verify webhook signature (uses webhook_secret, NOT api key_secret)
	webhookSecret := uc.config.RazorPayWebhookSecret
	h := hmac.New(sha256.New, []byte(webhookSecret))
	h.Write(rawBody)
	expected := hex.EncodeToString(h.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) != 1 {
		log.Printf("[WEBHOOK_SIGNATURE_FAIL] signature mismatch")
		return ErrInvalidWebhookSignature
	}

	// 2. Parse event
	var event struct {
		Event   string `json:"event"`
		Payload struct {
			Payment struct {
				Entity struct {
					ID      string  `json:"id"`
					OrderID string  `json:"order_id"`
					Amount  float64 `json:"amount"`
					Status  string  `json:"status"`
				} `json:"entity"`
			} `json:"payment"`
		} `json:"payload"`
	}
	if err := json.Unmarshal(rawBody, &event); err != nil {
		return fmt.Errorf("parse webhook: %w", err)
	}

	// 3. Only process payment.captured
	if event.Event != "payment.captured" {
		return nil
	}

	paymentID := event.Payload.Payment.Entity.ID
	rzpOrderID := event.Payload.Payment.Entity.OrderID

	// 4. Idempotency check
	_, err := uc.subRepo.FindSubscriptionOrderByRazorpayPaymentID(ctx, paymentID)
	if err == nil {
		return nil // already processed
	}

	// 5. Find subscription order
	order, err := uc.subRepo.FindSubscriptionOrderByRazorpayOrderID(ctx, rzpOrderID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil // not a subscription order, ignore
		}
		return fmt.Errorf("find order: %w", err)
	}

	if order.Status != domain.SubStatusCreated {
		return nil // already paid or expired
	}

	// 6. Cross-check amount
	webhookAmountPaise := uint(event.Payload.Payment.Entity.Amount)
	if webhookAmountPaise != uint(order.Price.AmountMinor) {
		log.Printf("[WEBHOOK_AMOUNT_MISMATCH] order_id=%s expected=%d got=%d",
			rzpOrderID, order.Price.AmountMinor, webhookAmountPaise)
		return ErrPaymentAmountMismatch
	}

	// 7. Atomic: mark paid + activate
	plan, err := uc.subRepo.FindSubscriptionPlanByID(ctx, order.PlanID)
	if err != nil {
		return fmt.Errorf("find plan: %w", err)
	}

	webhookOrderID := order.ID
	return uc.subRepo.Transaction(func(trxRepo repoIface.SubscriptionRepository) error {
		if err := trxRepo.UpdateSubscriptionOrderToPaid(ctx, order.ID, paymentID); err != nil {
			return err
		}
		// Deactivate any active trial before activating paid subscription
		if err := trxRepo.DeactivateTrialSubscription(ctx, order.UserID); err != nil {
			return err
		}
		now := time.Now()
		return trxRepo.ActivateSubscription(ctx, domain.UserSubscription{
			UserID:              order.UserID,
			PlanID:              order.PlanID,
			SubscriptionOrderID: &webhookOrderID,
			StartDate:           now,
			EndDate:             now.AddDate(0, 0, int(plan.DurationDays)),
			IsActive:            true,
		})
	})
}

// verifySignature performs HMAC-SHA256 signature verification.
func (uc *subscriptionPaymentUseCase) verifySignature(orderID, paymentID, signature, secret string) error {
	data := orderID + "|" + paymentID
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	expected := hex.EncodeToString(h.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(expected), []byte(signature)) != 1 {
		return ErrPaymentNotApproved
	}
	return nil
}

// crossCheckPayment fetches payment from Razorpay API and cross-checks order_id, amount, status.
func (uc *subscriptionPaymentUseCase) crossCheckPayment(paymentID, expectedOrderID string, expectedAmountPaise uint) error {
	client := razorpay.NewClient(uc.config.RazorPayKey, uc.config.RazorPaySecret)
	payment, err := client.Payment.Fetch(paymentID, nil, nil)
	if err != nil {
		return fmt.Errorf("razorpay payment fetch: %w", err)
	}

	// Check order_id
	if pOrderID, _ := payment["order_id"].(string); pOrderID != expectedOrderID {
		return ErrPaymentOrderMismatch
	}

	// Check amount
	var pAmount uint
	switch v := payment["amount"].(type) {
	case float64:
		pAmount = uint(v)
	case json.Number:
		n, _ := v.Int64()
		pAmount = uint(n)
	}
	if pAmount != expectedAmountPaise {
		return ErrPaymentAmountMismatch
	}

	// Check status
	status, _ := payment["status"].(string)
	if status != "captured" && status != "authorized" {
		return ErrPaymentNotApproved
	}

	return nil
}
