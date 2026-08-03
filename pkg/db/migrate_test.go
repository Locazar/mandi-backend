package db

import (
	"testing"

	"github.com/golang-migrate/migrate/v4/source/iofs"
)

// TestMigrationSourcesLoad is a startup guard, not a nicety.
//
// golang-migrate's iofs source refuses to load a directory containing two
// migrations with the same version number, so a single duplicate makes
// RunMigrations fail and the API unable to boot — against every database, not
// just a fresh one. That is exactly what a duplicate 000027 did between
// 2026-07-16 and this fix, and nothing in the test suite noticed.
func TestMigrationSourcesLoad(t *testing.T) {
	if _, err := iofs.New(migrationsFS, "migrations"); err != nil {
		t.Fatalf("migrations source failed to load (duplicate version number?): %v", err)
	}
	if _, err := iofs.New(legacyMigrationsFS, "legacy_migrations"); err != nil {
		t.Fatalf("legacy_migrations source failed to load (duplicate version number?): %v", err)
	}
}

func TestMigrationsFSNotEmpty(t *testing.T) {
	entries, err := migrationsFS.ReadDir("migrations")
	if err != nil {
		t.Fatal(err)
	}
	var up bool
	for _, e := range entries {
		if e.Name() == "000001_baseline.up.sql" {
			up = true
		}
	}
	if !up {
		t.Fatal("baseline up migration not embedded")
	}
}

func TestLegacyMigrationsFSNotEmpty(t *testing.T) {
	entries, err := legacyMigrationsFS.ReadDir("legacy_migrations")
	if err != nil {
		t.Fatal(err)
	}
	var up bool
	for _, e := range entries {
		if e.Name() == "000001_legacy_backup_compat.up.sql" {
			up = true
		}
	}
	if !up {
		t.Fatal("legacy compatibility migration not embedded")
	}
}
