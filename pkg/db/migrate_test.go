package db

import "testing"

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
