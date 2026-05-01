package db

import (
	"log"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"gorm.io/gorm"
)

// SeedProductItemFilters inserts sample data into ProductItemFilterType table
func SeedProductItemFilters(db *gorm.DB) error {
	// Sample data for ProductItemFilterType
	filters := []domain.ProductItemFilterType{
		{
			FilterName: "Offers",
		},
		{
			FilterName: "All",
		},
	}

	// Check if data already exists to avoid duplicates
	var count int64
	db.Model(&domain.ProductItemFilterType{}).Count(&count)
	if count > 0 {
		log.Println("ProductItemFilterType data already exists, skipping seed")
		return nil
	}

	// Insert data
	err := db.CreateInBatches(&filters, 10).Error
	if err != nil {
		return err
	}

	log.Println("Successfully seeded ProductItemFilterType data")
	return nil
}

// SeedSubscriptionPlans inserts default subscription plans if they don't exist.
func SeedSubscriptionPlans(db *gorm.DB) error {
	plans := []domain.SubscriptionPlan{
		{Name: "Silver", PriceMonthly: 199, DurationDays: 30, IsActive: true},
		{Name: "Gold", PriceMonthly: 499, DurationDays: 30, IsActive: true},
		{Name: "Platinum", PriceMonthly: 999, DurationDays: 30, IsActive: true},
	}

	var count int64
	db.Model(&domain.SubscriptionPlan{}).Count(&count)
	if count > 0 {
		log.Println("SubscriptionPlan data already exists, skipping seed")
		return nil
	}

	if err := db.CreateInBatches(&plans, 10).Error; err != nil {
		return err
	}

	log.Println("Successfully seeded SubscriptionPlan data")
	return nil
}
