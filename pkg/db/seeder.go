package db

import (
	"errors"
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
