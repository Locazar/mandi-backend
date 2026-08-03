package repository

import (
	"context"
	"sync"
	"testing"

	"github.com/rohit221990/mandi-backend/pkg/domain"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testDB lives in invoice_test.go (same package) — reuse it, do not redeclare.

// Invoice numbers must be gapless and unique even when several payments are
// verified at the same instant — the exact situation a naive
// "SELECT max+1 then INSERT" would corrupt.
func TestAllocateInvoiceSequenceIsConcurrencySafe(t *testing.T) {
	db := testDB(t)
	repo := NewInvoiceRepository(db)

	const fy = "2099-00" // isolated from real data
	require.NoError(t, db.Exec("DELETE FROM invoice_number_sequences WHERE financial_year = ?", fy).Error)
	t.Cleanup(func() {
		db.Exec("DELETE FROM invoice_number_sequences WHERE financial_year = ?", fy)
	})

	const workers = 50
	var (
		wg   sync.WaitGroup
		mu   sync.Mutex
		seen = make(map[int]bool, workers)
	)

	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			seq, err := repo.AllocateInvoiceSequence(context.Background(), fy)
			assert.NoError(t, err)
			mu.Lock()
			seen[seq] = true
			mu.Unlock()
		}()
	}
	wg.Wait()

	// Exactly the numbers 1..workers, each once — no duplicates, no gaps.
	assert.Len(t, seen, workers)
	for i := 1; i <= workers; i++ {
		assert.True(t, seen[i], "sequence %d missing — allocation is not gapless", i)
	}
}

// TestCreateInvoiceWithSequenceRollsBackAllocationOnInsertFailure is the test
// for the reason CreateInvoiceWithSequence exists: AllocateInvoiceSequence
// commits on its own, so a failed insert immediately after it — anything from
// a transient DB error to a CHECK-constraint violation — would otherwise burn
// a number permanently in a series required to be gapless. Wrapping both in
// one transaction means a failed insert also undoes the allocation.
func TestCreateInvoiceWithSequenceRollsBackAllocationOnInsertFailure(t *testing.T) {
	db := testDB(t)
	repo := NewInvoiceRepository(db)
	ctx := context.Background()

	const fy = "2099-01" // isolated from real data and from the concurrency test's FY
	var orderID string   // set once the second half creates a real order row
	cleanup := func() {
		// Children before parents: subscription_invoices references
		// subscription_orders via FK, and a single t.Cleanup registered once
		// (rather than two, which would run in reverse/LIFO order) keeps that
		// order guaranteed regardless of where orderID gets set.
		db.Exec("DELETE FROM subscription_invoices WHERE financial_year = ?", fy)
		if orderID != "" {
			db.Exec("DELETE FROM subscription_orders WHERE id = ?", orderID)
		}
		db.Exec("DELETE FROM invoice_number_sequences WHERE financial_year = ?", fy)
	}
	cleanup()
	t.Cleanup(func() { cleanup() })

	// This build callback returns an invoice that violates
	// ck_subscription_invoices_total_split (total != taxable_value +
	// gst_amount) — a real, non-duplicate insert failure.
	_, err := repo.CreateInvoiceWithSequence(ctx, fy, func(seq int) domain.Invoice {
		return domain.Invoice{
			ID:                  domain.NewID(domain.PrefixInvoice),
			SubscriptionOrderID: domain.NewID(domain.PrefixSubscOrder), // no matching order row: also violates the FK, either way this must not insert
			UserID:              "adm_rollback_probe",
			InvoiceNumber:       domain.FormatInvoiceNumber("LZ", fy, seq),
			FinancialYear:       fy,
			SequenceNumber:      seq,
			SellerLegalName:     "Probe",
			Total:               domain.INR(100),
			TaxableValue:        domain.INR(1), // deliberately wrong: 1 + 0 != 100
			GSTAmount:           domain.INR(0),
		}
	})
	require.Error(t, err, "the CHECK-constraint-violating insert should fail")

	var count int64
	require.NoError(t, db.Model(&domain.Invoice{}).
		Where("financial_year = ?", fy).Count(&count).Error)
	assert.Equal(t, int64(0), count, "the failed insert must not have left a row behind")

	// The real assertion: the sequence must NOT have advanced past 0. If it
	// did, the number allocated for the failed attempt is gone forever.
	var lastSeq int
	err = db.Raw("SELECT last_sequence FROM invoice_number_sequences WHERE financial_year = ?", fy).
		Scan(&lastSeq).Error
	require.NoError(t, err)
	assert.Equal(t, 0, lastSeq,
		"sequence advanced to %d despite the insert failing — a number was burned", lastSeq)

	// And a subsequent successful call gets sequence 1, not 2 — proving the
	// number was genuinely reclaimed, not just left unincremented by accident.
	// This insert needs a real order row to satisfy the FK.
	var planID string
	require.NoError(t, db.Raw("SELECT id FROM subscription_plans LIMIT 1").Scan(&planID).Error)
	if planID == "" {
		t.Skip("no subscription_plans row to hang a probe order off")
	}
	orderID = domain.NewID(domain.PrefixSubscOrder)
	require.NoError(t, db.Exec(`
		INSERT INTO subscription_orders
			(id, user_id, plan_id, price_amount_minor, razorpay_order_id, status)
		VALUES (?,?,?,?,?,'paid')`,
		orderID, "adm_rollback_probe", planID, 100, "order_"+orderID,
	).Error)

	inv, err := repo.CreateInvoiceWithSequence(ctx, fy, func(seq int) domain.Invoice {
		return domain.Invoice{
			ID:                  domain.NewID(domain.PrefixInvoice),
			SubscriptionOrderID: orderID,
			UserID:              "adm_rollback_probe",
			InvoiceNumber:       domain.FormatInvoiceNumber("LZ", fy, seq),
			FinancialYear:       fy,
			SequenceNumber:      seq,
			SellerLegalName:     "Probe",
			Total:               domain.INR(100),
			TaxableValue:        domain.INR(100),
			GSTAmount:           domain.INR(0),
		}
	})
	require.NoError(t, err)
	assert.Equal(t, 1, inv.SequenceNumber, "the reclaimed number must be reused, not skipped")
}
