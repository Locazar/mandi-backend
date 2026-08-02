package usecase_test

import (
	"context"
	"fmt"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/lib/pq"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// This test lives in the external usecase_test package rather than usecase
// itself: it needs pkg/repository's real constructors, and pkg/repository
// (via alert.go) imports pkg/usecase, so an internal test file importing
// pkg/repository would form an import cycle. See export_test.go for the
// unexported-access shim this relies on.

// integrationDB connects to the local docker-compose Postgres, following the
// same convention as pkg/repository/invoice_test.go: skip when TEST_DB_DSN is
// unset so `make test` stays runnable without a database.
//
// Example:
//
//	TEST_DB_DSN="host=localhost user=rohitjangid password=12345 dbname=mandi port=5432 sslmode=disable" \
//	  go test ./pkg/usecase/ -run TestMarkPaidAndActivate -v
func integrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN not set; skipping database-backed test")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	return db
}

// invoiceNumberPattern matches FormatInvoiceNumber's output, e.g.
// "LZ/2026-27/000001".
var invoiceNumberPattern = regexp.MustCompile(`^[A-Za-z0-9]+/\d{4}-\d{2}/\d{6}$`)

// TestMarkPaidAndActivateIssuesExactlyOneInvoice drives markPaidAndActivate
// end-to-end against real Postgres: the payment-critical transaction (order
// status, trial deactivation, subscription activation) and the invoice
// issuance that Task 6 moved OUTSIDE that transaction.
//
// The single most important assertion here is the duplicate guard: calling
// markPaidAndActivate a second time for the same order must never leave a
// second invoice behind. subscription_invoices.subscription_order_id is
// UNIQUE (migration 000034) precisely so that this holds regardless of
// transaction placement — see issueInvoiceForPaidOrder's docs.
func TestMarkPaidAndActivateIssuesExactlyOneInvoice(t *testing.T) {
	db := integrationDB(t)
	ctx := context.Background()

	subRepo := repository.NewSubscriptionRepository(db)
	invoiceRepo := repository.NewInvoiceRepository(db)
	adminRepo := repository.NewAdminRepository(db, nil)

	uc := usecase.NewTestSubscriptionPaymentUseCase(subRepo, invoiceRepo, adminRepo)

	// --- Isolated fixtures: a throwaway plan and order. No shop/admin row is
	// needed — GetShopByOwnerID missing the seller is already a handled,
	// non-fatal path (see [INVOICE_SHOP_LOOKUP_FAILED]) that this test exercises
	// incidentally. ---
	userID := domain.NewID(domain.PrefixAdmin)

	plan := domain.SubscriptionPlan{
		Name:         fmt.Sprintf("Integration Test Plan %s", domain.NewID(domain.PrefixSubscPlan)),
		PriceMonthly: domain.INR(149900),
		DurationDays: 90,
		IsActive:     true,
		Features:     pq.StringArray{},
	}
	require.NoError(t, db.Create(&plan).Error)

	order := domain.SubscriptionOrder{
		UserID:             userID,
		PlanID:             plan.ID,
		Price:              domain.INR(149900),
		GSTRateBasisPoints: 1800,
		GSTAmount:          domain.INR(22866),
		RazorpayOrderID:    "order_" + domain.NewID(domain.PrefixSubscOrder),
		Status:             domain.SubStatusCreated,
	}
	order, err := subRepo.CreateSubscriptionOrder(ctx, order)
	require.NoError(t, err)

	paymentID := "pay_" + domain.NewID(domain.PrefixSubscOrder)

	t.Cleanup(func() {
		// Children before parents so foreign keys are respected.
		db.Exec("DELETE FROM subscription_invoices WHERE subscription_order_id = ?", order.ID)
		db.Exec("DELETE FROM user_subscriptions WHERE subscription_order_id = ?", order.ID)
		db.Exec("DELETE FROM subscription_orders WHERE id = ?", order.ID)
		db.Exec("DELETE FROM subscription_plans WHERE id = ?", plan.ID)
	})

	// --- First call: payment completes, subscription activates, invoice issues. ---
	require.NoError(t, uc.MarkPaidAndActivate(ctx, order, plan, paymentID))

	var updatedOrder domain.SubscriptionOrder
	require.NoError(t, db.First(&updatedOrder, "id = ?", order.ID).Error)
	require.Equal(t, domain.SubscriptionStatusPaid, updatedOrder.Status)
	require.NotNil(t, updatedOrder.RazorpayPaymentID)
	require.Equal(t, paymentID, *updatedOrder.RazorpayPaymentID)

	var activeSubCount int64
	require.NoError(t, db.Model(&domain.UserSubscription{}).
		Where("subscription_order_id = ? AND is_active = ?", order.ID, true).
		Count(&activeSubCount).Error)
	require.Equal(t, int64(1), activeSubCount, "exactly one active user_subscriptions row must exist")

	var invoices []domain.Invoice
	require.NoError(t, db.Where("subscription_order_id = ?", order.ID).Find(&invoices).Error)
	require.Len(t, invoices, 1, "exactly one invoice must exist after the first payment")

	inv := invoices[0]
	require.True(t, invoiceNumberPattern.MatchString(inv.InvoiceNumber),
		"invoice number %q is not well-formed", inv.InvoiceNumber)
	require.Equal(t, inv.Total.AmountMinor, inv.TaxableValue.AmountMinor+inv.GSTAmount.AmountMinor,
		"total must equal taxable_value + gst_amount")

	// --- Second call for the SAME order: must not double-issue. ---
	// The order is already 'paid' in the database by now, so the payment
	// transaction itself will fail (UpdateSubscriptionOrderToPaid's
	// WHERE status='created' guard) and markPaidAndActivate returns an error
	// without reaching invoice issuance at all. That is a legitimate, expected
	// outcome of calling it twice with a stale in-memory order — what this
	// assertion cares about is the end state, not which layer stopped it.
	_ = uc.MarkPaidAndActivate(ctx, order, plan, paymentID)

	var invoicesAfterRetry []domain.Invoice
	require.NoError(t, db.Where("subscription_order_id = ?", order.ID).Find(&invoicesAfterRetry).Error)
	require.Len(t, invoicesAfterRetry, 1,
		"a second markPaidAndActivate call for the same order must not create a second invoice")
	require.Equal(t, inv.ID, invoicesAfterRetry[0].ID, "the surviving invoice must be the original one")
}

// TestIssueInvoiceForPaidOrderIsIdempotent exercises the duplicate guard that
// actually matters now that issuance runs outside the payment transaction.
//
// TestMarkPaidAndActivateIssuesExactlyOneInvoice cannot do this: its second
// call is rejected by UpdateSubscriptionOrderToPaid's `WHERE status='created'`
// guard and never reaches issuance at all, so its "no second invoice"
// assertion would hold even with no unique index whatsoever. This test calls
// issueInvoiceForPaidOrder directly, twice, on an order that is already paid —
// the state a webhook/verify race produces — so the only thing standing
// between it and a duplicate is the subscription_order_id unique index plus
// the existence check.
func TestIssueInvoiceForPaidOrderIsIdempotent(t *testing.T) {
	db := integrationDB(t)
	ctx := context.Background()

	subRepo := repository.NewSubscriptionRepository(db)
	invoiceRepo := repository.NewInvoiceRepository(db)
	adminRepo := repository.NewAdminRepository(db, nil)

	uc := usecase.NewTestSubscriptionPaymentUseCase(subRepo, invoiceRepo, adminRepo)

	userID := domain.NewID(domain.PrefixAdmin)

	plan := domain.SubscriptionPlan{
		Name:         fmt.Sprintf("Idempotency Test Plan %s", domain.NewID(domain.PrefixSubscPlan)),
		PriceMonthly: domain.INR(149900),
		DurationDays: 90,
		IsActive:     true,
		Features:     pq.StringArray{},
	}
	require.NoError(t, db.Create(&plan).Error)

	paidAt := time.Now()
	paymentID := "pay_" + domain.NewID(domain.PrefixSubscOrder)
	order := domain.SubscriptionOrder{
		UserID:             userID,
		PlanID:             plan.ID,
		Price:              domain.INR(149900),
		GSTRateBasisPoints: 1800,
		GSTAmount:          domain.INR(22866),
		RazorpayOrderID:    "order_" + domain.NewID(domain.PrefixSubscOrder),
		// Already paid — the state both racing callers would observe.
		Status:            domain.SubStatusPaid,
		RazorpayPaymentID: &paymentID,
		PaidAt:            &paidAt,
	}
	order, err := subRepo.CreateSubscriptionOrder(ctx, order)
	require.NoError(t, err)

	t.Cleanup(func() {
		db.Exec("DELETE FROM subscription_invoices WHERE subscription_order_id = ?", order.ID)
		db.Exec("DELETE FROM subscription_orders WHERE id = ?", order.ID)
		db.Exec("DELETE FROM subscription_plans WHERE id = ?", plan.ID)
	})

	countInvoices := func() int64 {
		var n int64
		require.NoError(t, db.Model(&domain.Invoice{}).
			Where("subscription_order_id = ?", order.ID).Count(&n).Error)
		return n
	}

	uc.IssueInvoiceForPaidOrder(ctx, order, plan, paymentID)
	require.Equal(t, int64(1), countInvoices(), "first issuance must create exactly one invoice")

	var first domain.Invoice
	require.NoError(t, db.First(&first, "subscription_order_id = ?", order.ID).Error)
	require.NotEmpty(t, first.InvoiceNumber)
	require.Equal(t, first.Total.AmountMinor,
		first.TaxableValue.AmountMinor+first.GSTAmount.AmountMinor)

	// Second issuance for the same already-paid order: must be a no-op, and
	// must not panic or leave a second row behind.
	uc.IssueInvoiceForPaidOrder(ctx, order, plan, paymentID)
	require.Equal(t, int64(1), countInvoices(),
		"a second issuance for the same order must not create a second invoice")

	// The surviving invoice must be the original, unmodified — an issued tax
	// invoice is immutable.
	var after domain.Invoice
	require.NoError(t, db.First(&after, "subscription_order_id = ?", order.ID).Error)
	require.Equal(t, first.ID, after.ID)
	require.Equal(t, first.InvoiceNumber, after.InvoiceNumber)
}
