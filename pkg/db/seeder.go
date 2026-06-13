package db

import (
	"errors"
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

// SeedSubscriptionPlans seeds duration-based subscription plans:
// 1 Month (₹399), 3 Months (₹1150), 6 Months (₹2199). It also deactivates
// the retired tier plans (Silver / Gold / Platinum) so old rows stay
// referenced by historical UserSubscription FKs but no longer appear in
// the paid-plans listing. Idempotent — safe to re-run.
func SeedSubscriptionPlans(db *gorm.DB) error {
	if err := db.Model(&domain.SubscriptionPlan{}).
		Where("name IN ?", []string{"Silver", "Gold", "Platinum"}).
		Update("is_active", false).Error; err != nil {
		return err
	}

	plans := []domain.SubscriptionPlan{
		{Name: "1 Month", PriceMonthly: domain.INR(39900), DurationDays: 30, IsActive: true},
		{Name: "3 Months", PriceMonthly: domain.INR(115000), DurationDays: 90, IsActive: true},
		{Name: "6 Months", PriceMonthly: domain.INR(219900), DurationDays: 180, IsActive: true},
	}
	for i := range plans {
		p := plans[i]
		var existing domain.SubscriptionPlan
		err := db.Where("name = ?", p.Name).First(&existing).Error
		switch {
		case errors.Is(err, gorm.ErrRecordNotFound):
			if err := db.Create(&p).Error; err != nil {
				return err
			}
			log.Printf("Created subscription plan %q", p.Name)
		case err != nil:
			return err
		default:
			existing.PriceMonthly = p.PriceMonthly
			existing.DurationDays = p.DurationDays
			existing.IsActive = true
			if err := db.Save(&existing).Error; err != nil {
				return err
			}
			log.Printf("Updated subscription plan %q", p.Name)
		}
	}
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
