package repository

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// testDB connects to the local docker-compose Postgres. It skips when
// TEST_DB_DSN is unset so `make test` stays runnable without a database.
//
// Example:
//
//	TEST_DB_DSN="host=localhost user=rohitjangid password=12345 dbname=mandi port=5432 sslmode=disable" \
//	  go test ./pkg/repository/ -v
func testDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		t.Skip("TEST_DB_DSN not set; skipping database-backed test")
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err)
	return db
}

// TestUpdateCompanyBillingProfileAdvancesUpdatedAt guards a GORM footgun.
//
// CompanyBillingProfile.UpdatedAt is tagged autoUpdateTime, but naming an
// autoUpdateTime column in Omit() is GORM's documented idiom for *disabling*
// timestamp tracking. An earlier revision used Omit("id", "updated_at"), which
// silently froze updated_at at its seed value forever — every edit looked
// untouched, destroying the only audit signal this singleton has. The bug is
// invisible without a live database, which is why this test needs one.
func TestUpdateCompanyBillingProfileAdvancesUpdatedAt(t *testing.T) {
	db := testDB(t)
	repo := NewInvoiceRepository(db)
	ctx := context.Background()

	before, err := repo.GetCompanyBillingProfile(ctx)
	require.NoError(t, err)

	originalCity := before.City
	t.Cleanup(func() {
		restore := before
		restore.City = originalCity
		_, _ = repo.UpdateCompanyBillingProfile(ctx, restore)
	})

	probe := before
	probe.City = "TimestampProbe"
	updated, err := repo.UpdateCompanyBillingProfile(ctx, probe)
	require.NoError(t, err)

	require.True(t, updated.UpdatedAt.After(before.UpdatedAt),
		"updated_at did not advance (%s -> %s): autoUpdateTime is disabled, "+
			"check that Omit() does not name updated_at",
		before.UpdatedAt, updated.UpdatedAt)
	require.Equal(t, "TimestampProbe", updated.City)
	// The singleton PK must survive a write untouched.
	require.Equal(t, domain.CompanyBillingProfileID, updated.ID)
}

// TestUpdateCompanyBillingProfileWritesClearedFields covers the other half of
// the Select("*") decision: GORM's default Updates skips zero values, so
// without Select("*") an admin could never clear an optional field.
func TestUpdateCompanyBillingProfileWritesClearedFields(t *testing.T) {
	db := testDB(t)
	repo := NewInvoiceRepository(db)
	ctx := context.Background()

	before, err := repo.GetCompanyBillingProfile(ctx)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = repo.UpdateCompanyBillingProfile(ctx, before)
	})

	seeded := before
	seeded.AddressLine2 = "Second line"
	_, err = repo.UpdateCompanyBillingProfile(ctx, seeded)
	require.NoError(t, err)

	cleared := seeded
	cleared.AddressLine2 = ""
	updated, err := repo.UpdateCompanyBillingProfile(ctx, cleared)
	require.NoError(t, err)

	require.Empty(t, updated.AddressLine2,
		"cleared field was not persisted: Updates skipped the zero value, "+
			"check that Select(\"*\") is present")
}

// TestInvoiceListingsTreatZeroLimitAsDefault guards a GORM footgun that is
// invisible in unit tests: Limit(0) emits a literal `LIMIT 0`, which returns no
// rows rather than all of them. request.GetPagination parses `?limit=0`
// successfully and overwrites its own default of 25, so without a guard a
// seller hitting /subscriptions/invoices?limit=0 would silently be told they
// have no bills at all.
func TestInvoiceListingsTreatZeroLimitAsDefault(t *testing.T) {
	db := testDB(t)
	repo := NewInvoiceRepository(db)
	ctx := context.Background()

	const (
		userID = "adm_zerolimit_probe"
		fy     = "2098-99"
	)

	// Isolated fixture: a fictitious user and financial year, so this never
	// collides with real invoices. subscription_invoices.subscription_order_id is
	// NOT NULL with an FK, so a throwaway order has to exist first.
	const (
		orderID   = "subo_zerolimit_probe"
		invoiceID = "inv_zerolimit_probe"
	)
	cleanup := func() {
		db.Exec("DELETE FROM subscription_invoices WHERE id = ?", invoiceID)
		db.Exec("DELETE FROM subscription_orders WHERE id = ?", orderID)
	}
	cleanup()
	t.Cleanup(cleanup)

	var planID string
	if err := db.Raw("SELECT id FROM subscription_plans LIMIT 1").Scan(&planID).Error; err != nil || planID == "" {
		t.Skip("no subscription_plans row to hang a probe order off")
	}

	require.NoError(t, db.Exec(`
		INSERT INTO subscription_orders
			(id, user_id, plan_id, price_amount_minor, razorpay_order_id, status)
		VALUES (?,?,?,?,?,'paid')`,
		orderID, userID, planID, 100, "order_"+orderID,
	).Error)

	require.NoError(t, db.Exec(`
		INSERT INTO subscription_invoices
			(id, subscription_order_id, user_id, invoice_number, financial_year,
			 sequence_number, invoice_date, seller_legal_name,
			 total_amount_minor, taxable_value_amount_minor, gst_amount_amount_minor)
		VALUES (?,?,?,?,?,?,?,?,?,?,?)`,
		invoiceID, orderID, userID, "ZZ/2098-99/000001", fy,
		1, time.Now(), "Probe",
		int64(100), int64(85), int64(15),
	).Error)

	t.Run("FindInvoicesByUserID", func(t *testing.T) {
		got, err := repo.FindInvoicesByUserID(ctx, userID, request.Pagination{Limit: 0, Offset: 0})
		require.NoError(t, err)
		require.NotEmpty(t, got, "zero limit returned no rows — GORM emitted LIMIT 0")
	})

	t.Run("ListInvoices", func(t *testing.T) {
		got, err := repo.ListInvoices(ctx, domain.InvoiceFilter{UserID: userID, Limit: 0})
		require.NoError(t, err)
		require.NotEmpty(t, got, "zero limit returned no rows — GORM emitted LIMIT 0")
	})
}
