package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/db"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/utils"
	"gorm.io/gorm"
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
		Email:    "admin@localzar.com",
		Password: "Admin@123",
		Role:     domain.AdminRoleSuperAdmin,
	},
	{
		FullName: "Support Staff",
		UserName: "support",
		Email:    "support@localzar.com",
		Password: "Support@123",
		Role:     domain.AdminRoleSupportStaff,
	},
	{
		FullName: "Catalog Manager",
		UserName: "catalog",
		Email:    "catalog@localzar.com",
		Password: "Catalog@123",
		Role:     domain.AdminRoleCatalogManager,
	},
	{
		FullName: "Marketing Manager",
		UserName: "marketing",
		Email:    "marketing@localzar.com",
		Password: "Marketing@123",
		Role:     domain.AdminRoleMarketingManager,
	},
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	database, err := db.ConnectDatabase(cfg)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	created := 0
	skipped := 0

	for _, a := range platformAdmins {
		var existing domain.Admin
		err := database.Where("email = ?", a.Email).First(&existing).Error
		if err == nil {
			fmt.Printf("  SKIP  %-30s (already exists, role=%s)\n", a.Email, existing.Role)
			skipped++
			continue
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			log.Fatalf("db error checking %s: %v", a.Email, err)
		}

		hash, err := utils.GenerateHashFromPassword(a.Password)
		if err != nil {
			log.Fatalf("failed to hash password for %s: %v", a.Email, err)
		}

		admin := domain.Admin{
			FullName: a.FullName,
			UserName: a.UserName,
			Email:    a.Email,
			Password: hash,
			Status:   domain.AdminStatusActive,
			Role:     a.Role,
		}

		if err := database.Create(&admin).Error; err != nil {
			log.Fatalf("failed to create %s: %v", a.Email, err)
		}

		fmt.Printf("  CREATE %-30s id=%s role=%s\n", a.Email, admin.ID, admin.Role)
		created++
	}

	fmt.Printf("\nDone. Created: %d  Skipped: %d\n", created, skipped)
	os.Exit(0)
}
