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
		{Name: "Silver", PriceMonthly: domain.INR(19900), DurationDays: 30, IsActive: true},
		{Name: "Gold", PriceMonthly: domain.INR(49900), DurationDays: 30, IsActive: true},
		{Name: "Platinum", PriceMonthly: domain.INR(99900), DurationDays: 30, IsActive: true},
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

// SeedFreeTrialPlan inserts the Free Trial plan if it doesn't exist (additive, safe on existing DBs).
func SeedFreeTrialPlan(db *gorm.DB) error {
	plan := domain.SubscriptionPlan{
		Name:         "Free Trial",
		PriceMonthly: domain.INR(0),
		DurationDays: 90,
		IsActive:     true,
	}
	result := db.Where("name = ?", plan.Name).FirstOrCreate(&plan)
	if result.Error != nil {
		return result.Error
	}
	log.Println("Free Trial plan seeded")
	return nil
}
