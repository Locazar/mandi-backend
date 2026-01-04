package db

import (
	"fmt"
	"log"

	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// func to connect data base using config(database config) and return address of a new instnce of gorm DB
func ConnectDatabase(cfg config.Config) (*gorm.DB, error) {

	dsn := fmt.Sprintf("host=%s user=%s dbname=%s port=%s password=%s", cfg.DBHost, cfg.DBUser, cfg.DBName, cfg.DBPort, cfg.DBPassword)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		SkipDefaultTransaction: true,
	})

	if err != nil {
		return nil, err
	}

	// migrate the database tables
	err = db.AutoMigrate(

		//auth
		domain.AdminRefreshSession{},
		domain.UserRefreshSession{},
		domain.OtpSession{},
		//user
		domain.User{},
		domain.Country{},
		domain.Address{},
		domain.UserAddress{},

		//admin
		domain.Admin{},
		domain.ShopVerification{},
		domain.ShopVerificationHistory{},

		//product
		domain.Category{},
		domain.Product{},
		domain.Variation{},
		domain.VariationOption{},
		domain.ProductItem{},
		domain.ProductConfiguration{},
		domain.ProductImage{},

		// wish list
		domain.WishList{},

		// cart
		domain.Cart{},
		domain.CartItem{},

		// order
		domain.OrderStatus{},
		domain.ShopOrder{},
		domain.OrderLine{},
		domain.OrderReturn{},

		//offer
		domain.Offer{},
		domain.OfferCategory{},
		domain.OfferProduct{},

		// coupon
		domain.Coupon{},
		domain.CouponUses{},

		//wallet
		domain.Wallet{},
		domain.Transaction{},

		//Advertisement
		domain.Advertisement{},

		//Notification
		domain.Notification{},

		//Shop Details
		domain.ShopDetails{},
		domain.ShopOffer{},

		//Payment Methods
		domain.PaymentMethod{},

		// department
		domain.Department{},
		domain.SubCategory{},	

		domain.SubTypeAttributes{},
		domain.SubTypeAttributeOptions{},
		domain.SubCategoryDetails{},
		domain.ProductItemView{},
		domain.ProductItemFilterType{},
	)

	if err != nil {
		log.Printf("Warning: failed to migrate database models: %v. Continuing with existing schema.", err)
		// Don't return error - continue with existing database schema
	}

	// setup the triggers
	if err := SetUpDBTriggers(db); err != nil {
		log.Printf("Warning: failed to setup database triggers: %v. Continuing without triggers.", err)
		// Don't return error - continue without triggers
	}

	if err := saveAdmin(db, cfg.AdminEmail, cfg.AdminPassword); err != nil {
		log.Printf("Warning: failed to save admin: %v. Continuing without admin setup.", err)
		// Don't return error - continue without admin setup
	}

	if err := saveOrderStatuses(db); err != nil {
		log.Printf("Warning: failed to save order statuses: %v. Continuing.", err)
	}
	if err := savePaymentMethods(db); err != nil {
		log.Printf("Warning: failed to save payment methods: %v. Continuing.", err)
	}

	if err := SeedCountries(db); err != nil {
		log.Printf("Warning: failed to seed countries: %v. Continuing.", err)
	}

	return db, err
}
