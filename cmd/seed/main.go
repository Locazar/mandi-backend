package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/db"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/utils"
)

// platformAdmins are the internal ops accounts seeded into the DB.
// Passwords are bcrypt-hashed on insert. Existing emails are skipped (idempotent).
var platformAdmins = []struct {
	FullName string
	UserName string
	Email    string
	Password string
	Role     domain.AdminRole
}{
	{
		FullName: "Super Admin",
		UserName: "superadmin",
		Email:    "admin@locazar.com",
		Password: "Admin@123",
		Role:     domain.AdminRoleSuperAdmin,
	},
	{
		FullName: "Support Staff",
		UserName: "support",
		Email:    "support@locazar.com",
		Password: "Support@123",
		Role:     domain.AdminRoleSupportStaff,
	},
	{
		FullName: "Catalog Manager",
		UserName: "catalog",
		Email:    "catalog@locazar.com",
		Password: "Catalog@123",
		Role:     domain.AdminRoleCatalogManager,
	},
	{
		FullName: "Marketing Manager",
		UserName: "marketing",
		Email:    "marketing@locazar.com",
		Password: "Marketing@123",
		Role:     domain.AdminRoleMarketingManager,
	},
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// ConnectDatabase also runs golang-migrate, so migration 000007 (add role column)
	// will be applied automatically before we insert.
	database, err := db.ConnectDatabase(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	sqlDB, err := database.DB()
	if err != nil {
		log.Fatalf("failed to get sql.DB: %v", err)
	}

	created := 0
	skipped := 0

	for _, a := range platformAdmins {
		var count int
		if err := sqlDB.QueryRow(`SELECT COUNT(*) FROM admins WHERE email = $1`, a.Email).Scan(&count); err != nil {
			log.Fatalf("db error checking %s: %v", a.Email, err)
		}
		if count > 0 {
			fmt.Printf("  SKIP  %-32s (already exists)\n", a.Email)
			skipped++
			continue
		}

		hash, err := utils.GenerateHashFromPassword(a.Password)
		if err != nil {
			log.Fatalf("failed to hash password for %s: %v", a.Email, err)
		}

		now := time.Now()
		adminID := domain.NewID(domain.PrefixAdmin)
		var insertedID string

		err = sqlDB.QueryRow(`
			INSERT INTO admins (id, user_name, full_name, email, password, status, role, verified_seller, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, FALSE, $8, $9)
			RETURNING id`,
			adminID, a.UserName, a.FullName, a.Email, hash,
			string(domain.AdminStatusActive), string(a.Role),
			now, now,
		).Scan(&insertedID)
		if err != nil {
			log.Fatalf("failed to create %s: %v", a.Email, err)
		}

		fmt.Printf("  CREATE %-32s id=%s role=%s\n", a.Email, insertedID, a.Role)
		created++
	}

	fmt.Printf("\nDone. Created: %d  Skipped: %d\n", created, skipped)

	sqlFile := filepath.Join("baseline_seed.sql")
	sqlBytes, err := os.ReadFile(sqlFile)
	if err != nil {
		log.Fatalf("failed to read baseline_seed.sql: %v", err)
	}
	if _, err := sqlDB.Exec(string(sqlBytes)); err != nil {
		log.Fatalf("failed to apply baseline_seed.sql: %v", err)
	}
	fmt.Println("baseline_seed.sql applied.")

	os.Exit(0)
}
