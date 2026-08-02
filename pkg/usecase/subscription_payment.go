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

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/razorpay/razorpay-go"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoIface "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	service "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"gorm.io/gorm"
)

// uniqueViolationSQLState is Postgres's SQLSTATE code for a unique-constraint
// violation (23505). Used as the fallback duplicate-invoice check when the
// driver error hasn't been translated into gorm.ErrDuplicatedKey.
const uniqueViolationSQLState = "23505"

const subscriptionOrderExpiryMinutes = 30

type subscriptionPaymentUseCase struct {
	subRepo     repoIface.SubscriptionRepository
	paymentRepo repoIface.PaymentRepository
	userRepo    repoIface.UserRepository
	invoiceRepo repoIface.InvoiceRepository
	adminRepo   repoIface.AdminRepository
	invoiceUC   service.InvoiceUseCase
	config      config.Config
}

func NewSubscriptionPaymentUseCase(
	subRepo repoIface.SubscriptionRepository,
	paymentRepo repoIface.PaymentRepository,
	userRepo repoIface.UserRepository,
	invoiceRepo repoIface.InvoiceRepository,
	adminRepo repoIface.AdminRepository,
	invoiceUC service.InvoiceUseCase,
	cfg config.Config,
) service.SubscriptionPaymentUseCase {
	return &subscriptionPaymentUseCase{
		subRepo:     subRepo,
		paymentRepo: paymentRepo,
		userRepo:    userRepo,
		invoiceRepo: invoiceRepo,
		adminRepo:   adminRepo,
		invoiceUC:   invoiceUC,
		config:      cfg,
	}
}

// CreateSubscriptionOrder creates a Razorpay order for a subscription plan.
func (uc *subscriptionPaymentUseCase) CreateSubscriptionOrder(ctx context.Context, userID string, req request.CreateSubscriptionOrderRequest) (response.SubscriptionOrderResponse, error) {
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
	// Razorpay caps receipt at 40 chars; keep traceable suffixes of both IDs.
	receipt := fmt.Sprintf("sub_%s_%s_%d", lastN(userID, 8), lastN(plan.ID, 8), time.Now().Unix())

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
	// GST is included in the price; snapshot the current rate and the computed
	// GST portion so historical reporting is unaffected by future rate changes.
	// Rate is admin-configurable at runtime (see AdminRepository.
	// GetSubscriptionGSTConfig / admin-portal Settings) rather than fixed via
	// env var, so it's read fresh per order instead of from uc.config.
	gstRate, err := uc.subRepo.GetGSTRateBasisPoints(ctx)
	if err != nil {
		return response.SubscriptionOrderResponse{}, fmt.Errorf("failed to load GST rate: %w", err)
	}
	subOrder := domain.SubscriptionOrder{
		UserID:             userID,
		PlanID:             plan.ID,
		Price:              plan.PriceMonthly,
		GSTRateBasisPoints: gstRate,
		GSTAmount:          plan.PriceMonthly.InclusiveTaxPortion(gstRate),
		RazorpayOrderID:    rzpOrderID,
		Status:             domain.SubStatusCreated,
	}

	subOrder, err = uc.subRepo.CreateSubscriptionOrder(ctx, subOrder)
	if err != nil {
		return response.SubscriptionOrderResponse{}, fmt.Errorf("persist order: %w", err)
	}

	// 7. Return checkout data
	return response.SubscriptionOrderResponse{
		OrderID:            rzpOrderID,
		KeyID:              uc.config.RazorPayKey,
		Amount:             uint(amountPaise),
		Currency:           "INR",
		ShopOrderID:        subOrder.ID,
		GSTRateBasisPoints: subOrder.GSTRateBasisPoints,
		GSTAmount:          uint(subOrder.GSTAmount.AmountMinor),
		Prefill: response.SubscriptionPrefill{
			Email:   user.Email,
			Contact: user.Phone,
		},
	}, nil
}

// VerifySubscriptionPayment performs the 4-step verification.
func (uc *subscriptionPaymentUseCase) VerifySubscriptionPayment(ctx context.Context, userID string, req request.VerifySubscriptionPaymentRequest) (response.SubscriptionVerificationResponse, error) {
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

	if err := uc.markPaidAndActivate(ctx, order, plan, req.PaymentID); err != nil {
		return response.SubscriptionVerificationResponse{}, fmt.Errorf("activate subscription: %w", err)
	}

	return response.SubscriptionVerificationResponse{
		Status:      "success",
		Plan:        plan.Name,
		ActivatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

// markPaidAndActivate is the single implementation of "this payment succeeded".
// Both the synchronous verify path and the webhook path call it, so they cannot
// drift apart.
//
// Only the payment-critical steps run inside a transaction: marking the order
// paid, deactivating any trial, and activating the subscription. If that
// transaction fails, the payment genuinely did not complete and the error
// propagates to the caller.
//
// Invoice issuance happens AFTER the transaction commits, in
// issueInvoiceForPaidOrder. It is deliberately kept out of the transaction: an
// invoice write failure must never roll back an already-captured payment,
// leaving the order stuck at status='created' while Razorpay holds the money.
// That is unrecoverable. A missing invoice, by contrast, is recoverable — it
// can be backfilled — so issueInvoiceForPaidOrder is written to never fail the
// payment; see its docs for how it stays safe against duplicate issuance.
func (uc *subscriptionPaymentUseCase) markPaidAndActivate(
	ctx context.Context,
	order domain.SubscriptionOrder,
	plan domain.SubscriptionPlan,
	paymentID string,
) error {
	orderID := order.ID

	if err := uc.subRepo.Transaction(func(repo repoIface.SubscriptionRepository) error {
		if err := repo.UpdateSubscriptionOrderToPaid(ctx, order.ID, paymentID); err != nil {
			return err
		}
		if err := repo.DeactivateTrialSubscription(ctx, order.UserID); err != nil {
			return err
		}

		now := time.Now()
		if err := repo.ActivateSubscription(ctx, domain.UserSubscription{
			UserID:              order.UserID,
			PlanID:              order.PlanID,
			SubscriptionOrderID: &orderID,
			StartDate:           now,
			EndDate:             now.AddDate(0, 0, int(plan.DurationDays)),
			IsActive:            true,
		}); err != nil {
			return err
		}
		return nil
	}); err != nil {
		return err
	}

	uc.issueInvoiceForPaidOrder(ctx, order, plan, paymentID)
	return nil
}

// issueInvoiceForPaidOrder issues the tax invoice for a payment that has
// already been committed to the database. It runs outside the payment
// transaction and must NEVER return an error that fails the payment — every
// failure path here logs and returns.
//
// Duplicate protection: subscription_invoices.subscription_order_id is UNIQUE
// (migration 000034), so if the webhook and the synchronous verify path race
// each other into this function for the same order, the loser's CreateInvoice
// hits a unique-constraint violation. That is treated as success — the
// invoice already exists — not as an error.
func (uc *subscriptionPaymentUseCase) issueInvoiceForPaidOrder(
	ctx context.Context,
	order domain.SubscriptionOrder,
	plan domain.SubscriptionPlan,
	paymentID string,
) {
	profile, err := uc.invoiceRepo.GetCompanyBillingProfile(ctx)
	if err != nil {
		// A missing profile must not cost the seller their subscription.
		log.Printf("[INVOICE_PROFILE_MISSING] order_id=%s err=%v", order.ID, err)
	}

	shop, err := uc.adminRepo.GetShopByOwnerID(ctx, order.UserID)
	if err != nil {
		log.Printf("[INVOICE_SHOP_LOOKUP_FAILED] order_id=%s user_id=%s err=%v", order.ID, order.UserID, err)
		shop = domain.ShopDetails{}
	}

	// Bail out before allocating a number if this order already has an invoice.
	//
	// Sequence allocation commits on its own now that issuance runs outside the
	// payment transaction, so a number consumed by an insert that then fails the
	// subscription_order_id unique index is burned — a gap in a sequence that is
	// required to be gapless.
	//
	// Today no caller can reach that: markPaidAndActivate only calls this after
	// its transaction commits, and UpdateSubscriptionOrderToPaid's
	// `WHERE status='created'` guard rejects any second attempt before issuance
	// is reached. This check is what makes issueInvoiceForPaidOrder safe to call
	// from anywhere — a reconciler or a retry that does not go through the
	// payment transaction — rather than load-bearing for the current paths.
	if existing, err := uc.invoiceRepo.FindInvoiceBySubscriptionOrderID(ctx, order.ID); err == nil && existing.ID != "" {
		log.Printf("[INVOICE_ALREADY_ISSUED] invoice_number=%s order_id=%s user_id=%s",
			existing.InvoiceNumber, order.ID, order.UserID)
		return
	}

	// The order in hand may still have PaidAt unset (the transaction set it in
	// the database), so stamp it for the snapshot.
	now := time.Now()
	paidOrder := order
	paidOrder.RazorpayPaymentID = &paymentID
	if paidOrder.PaidAt == nil {
		paidOrder.PaidAt = &now
	}

	seq, err := uc.invoiceRepo.AllocateInvoiceSequence(ctx, domain.FinancialYear(*paidOrder.PaidAt))
	if err != nil {
		log.Printf("[INVOICE_ISSUE_FAILED] step=allocate_sequence order_id=%s user_id=%s err=%v",
			order.ID, order.UserID, err)
		return
	}

	inv := BuildInvoice(BuildInvoiceInput{
		Order:          paidOrder,
		Plan:           plan,
		Shop:           shop,
		Profile:        profile,
		SequenceNumber: seq,
	})

	if _, err := uc.invoiceRepo.CreateInvoice(ctx, inv); err != nil {
		if isDuplicateInvoiceError(err) {
			log.Printf("[INVOICE_ALREADY_ISSUED] order_id=%s user_id=%s", order.ID, order.UserID)
			return
		}
		log.Printf("[INVOICE_ISSUE_FAILED] step=create_invoice order_id=%s user_id=%s err=%v",
			order.ID, order.UserID, err)
		return
	}

	log.Printf("[INVOICE_ISSUED] invoice_number=%s order_id=%s user_id=%s",
		inv.InvoiceNumber, order.ID, order.UserID)

	// PDF generation runs in the background and is best-effort: the invoice row
	// is already durable, and the download path re-renders on demand if this
	// never completes. A PDF failure must never fail the payment or block the
	// HTTP response that already succeeded by this point.
	go func() {
		bgCtx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		key, err := uc.invoiceUC.GenerateAndStorePDF(bgCtx, inv)
		if err != nil {
			log.Printf("[INVOICE_PDF] render failed invoice=%s: %v", inv.InvoiceNumber, err)
			return
		}
		if err := uc.invoiceRepo.SetInvoicePDF(bgCtx, inv.ID, key); err != nil {
			log.Printf("[INVOICE_PDF] key persist failed invoice=%s: %v", inv.InvoiceNumber, err)
		}
	}()
}

// isDuplicateInvoiceError reports whether err is a unique-constraint violation
// on subscription_invoices.subscription_order_id — i.e. an invoice for this
// order already exists. Prefers the typed gorm sentinel; falls back to
// inspecting the underlying Postgres error for setups where the driver error
// hasn't been translated (gorm's postgres driver only does that translation
// when TranslateError is enabled).
func isDuplicateInvoiceError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, gorm.ErrDuplicatedKey) {
		return true
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == uniqueViolationSQLState
	}
	return false
}

// HandlePaymentFailure logs a payment failure without changing order state.
func (uc *subscriptionPaymentUseCase) HandlePaymentFailure(ctx context.Context, userID string, req request.PaymentFailureRequest) error {
	log.Printf("[SUBSCRIPTION_PAYMENT_FAILURE] user_id=%s order_id=%s code=%d message=%s time=%s",
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

	if err := uc.markPaidAndActivate(ctx, order, plan, paymentID); err != nil {
		return fmt.Errorf("activate subscription: %w", err)
	}
	return nil
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

func lastN(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[len(s)-n:]
}
