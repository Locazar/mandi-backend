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
