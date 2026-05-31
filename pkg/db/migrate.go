package db

import (
	"database/sql"
	"embed"
	"fmt"
	"io/fs"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"gorm.io/gorm"
)

//go:embed migrations/*.sql
var migrationsFS embed.FS

//go:embed legacy_migrations/*.sql
var legacyMigrationsFS embed.FS

// RunMigrations applies all pending migrations and returns an error (so the
// caller can abort startup) on failure or a dirty version.
func RunMigrations(gormDB *gorm.DB) error {
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("migrate: get sql.DB: %w", err)
	}
	driver, err := postgres.WithInstance(sqlDB, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("migrate: postgres driver: %w", err)
	}
	legacyBootstrap, err := shouldUseLegacyBootstrap(sqlDB)
	if err != nil {
		return fmt.Errorf("migrate: detect legacy bootstrap: %w", err)
	}

	var sourceFS fs.FS = migrationsFS
	sourceDir := "migrations"
	if legacyBootstrap {
		sourceFS = legacyMigrationsFS
		sourceDir = "legacy_migrations"
	}

	src, err := iofs.New(sourceFS, sourceDir)
	if err != nil {
		return fmt.Errorf("migrate: iofs source: %w", err)
	}
	m, err := migrate.NewWithInstance("iofs", src, "postgres", driver)
	if err != nil {
		return fmt.Errorf("migrate: new: %w", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		return fmt.Errorf("migrate: up: %w", err)
	}
	return nil
}

func shouldUseLegacyBootstrap(db *sql.DB) (bool, error) {
	var hasSchemaMigrations bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT 1
			FROM information_schema.tables
			WHERE table_schema = 'public' AND table_name = 'schema_migrations'
		)
	`).Scan(&hasSchemaMigrations)
	if err != nil {
		return false, err
	}
	if hasSchemaMigrations {
		return false, nil
	}

	var isLegacy bool
	err = db.QueryRow(`
		SELECT (
			EXISTS (
				SELECT 1
				FROM information_schema.tables
				WHERE table_schema = 'public' AND table_name = 'refresh_sessions'
			)
			OR EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'users' AND column_name = 'id' AND data_type = 'bigint'
			)
			OR EXISTS (
				SELECT 1
				FROM information_schema.columns
				WHERE table_schema = 'public' AND table_name = 'admins' AND column_name = 'id' AND data_type = 'bigint'
			)
		)
	`).Scan(&isLegacy)
	if err != nil {
		return false, err
	}

	return isLegacy, nil
}
