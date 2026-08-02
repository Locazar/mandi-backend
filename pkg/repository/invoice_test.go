package repository

import (
	"context"
	"os"
	"testing"

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
