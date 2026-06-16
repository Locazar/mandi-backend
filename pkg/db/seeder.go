package db

import (
	"log"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"gorm.io/gorm"
)

// SeedProductItemFilters inserts default filter types ("Offers", "All") for each
// existing shop that doesn't have them yet. Skips silently if no shops exist.
func SeedProductItemFilters(db *gorm.DB) error {
	var shopIDs []string
	if err := db.Model(&domain.ShopDetails{}).Pluck("id", &shopIDs).Error; err != nil {
		return err
	}
	if len(shopIDs) == 0 {
		log.Println("No shops found, skipping ProductItemFilterType seed")
		return nil
	}

	defaultFilterNames := []string{"Offers", "All"}
	seeded := 0
	for _, shopID := range shopIDs {
		for _, name := range defaultFilterNames {
			var count int64
			db.Model(&domain.ProductItemFilterType{}).
				Where("filter_name = ? AND shop_id = ?", name, shopID).
				Count(&count)
			if count > 0 {
				continue
			}
			f := domain.ProductItemFilterType{FilterName: name, ShopID: shopID}
			if err := db.Create(&f).Error; err != nil {
				return err
			}
			seeded++
		}
	}

	if seeded > 0 {
		log.Printf("Successfully seeded %d ProductItemFilterType records", seeded)
	} else {
		log.Println("ProductItemFilterType data already exists, skipping seed")
	}
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
