package db

import (
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/rohit221990/mandi-backend/pkg/config"
)

// func to connect data base using config(database config) and return address of a new instnce of gorm DB
func ConnectDatabase(cfg config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s dbname=%s port=%s password=%s sslmode=disable", cfg.DBHost, cfg.DBUser, cfg.DBName, cfg.DBPort, cfg.DBPassword)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
		// Silence GORM's internal "insufficient arguments" logger formatting
		// bugs that appear during AutoMigrate schema checks on some Postgres
		// versions.  Warnings and errors from our own code still use log.Printf.
		Logger: gormlogger.Default.LogMode(gormlogger.Error),
	})
	if err != nil {
		return nil, err
	}

	// configure underlying sql.DB for connection pooling
	sqlDB, err := db.DB()
	if err == nil {
		// use ~80% of Postgres max_connections as safe pool size
		sqlDB.SetMaxOpenConns(240)
		// keep a fraction of connections idle for quick bursts
		sqlDB.SetMaxIdleConns(60)
		// recycle connections periodically (avoid stale network state)
		sqlDB.SetConnMaxLifetime(30 * time.Minute)
		sqlDB.SetConnMaxIdleTime(5 * time.Minute)

		// verify connectivity with a short timeout
		pingCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = sqlDB.PingContext(pingCtx) // record error if necessary
	}

	// run versioned migrations (fail-fast on any error or dirty state)
	if err := RunMigrations(db); err != nil {
		return nil, fmt.Errorf("database migration failed: %w", err)
	}

	// setup the triggers
	if err := SetUpDBTriggers(db); err != nil {
		log.Printf("Warning: failed to setup database triggers: %v. Continuing without triggers.", err)
		// Don't return error - continue without triggers
	}

	// create search indexes (must run after AutoMigrate so tables exist)
	if err := SetupSearchIndexes(db); err != nil {
		log.Printf("Warning: failed to setup search indexes: %v. Continuing without search indexes.", err)
	}

	if err := saveAdmin(db, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Printf("Warning: failed to save admin: %v. Continuing without admin setup.", err)
		// Don't return error - continue without admin setup
	}

	if err := savePaymentMethods(db); err != nil {
		log.Printf("Warning: failed to save payment methods: %v. Continuing.", err)
	}

	if err := SeedCountries(db); err != nil {
		log.Printf("Warning: failed to seed countries: %v. Continuing.", err)
	}

	if err := SeedSubscriptionPlans(db); err != nil {
		log.Printf("Warning: failed to seed subscription plans: %v. Continuing.", err)
	}

	if err := SeedFreeTrialPlan(db); err != nil {
		log.Printf("Warning: failed to seed free trial plan: %v. Continuing.", err)
	}

	return db, nil
}
