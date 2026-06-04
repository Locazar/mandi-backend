package db

import (
	"fmt"

	"gorm.io/gorm"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// To save predefined payment methods on database if its not exist
func savePaymentMethods(db *gorm.DB) error {
	paymentMethods := []domain.PaymentMethod{
		{
			Name:          domain.CodPayment,
			MaximumAmount: domain.INR(domain.CodMaximumAmount),
		},
		{
			Name:          domain.RazopayPayment,
			MaximumAmount: domain.INR(domain.RazorPayMaximumAmount),
		},
		{
			Name:          domain.StripePayment,
			MaximumAmount: domain.INR(domain.StripeMaximumAmount),
		},
	}

	var (
		searchQuery = `SELECT EXISTS(SELECT 1 FROM payment_methods WHERE name = $1) AS exist`
		insertQuery = `INSERT INTO payment_methods (id, name, maximum_amount_amount_minor, maximum_amount_currency) VALUES ($1, $2, $3, $4)`
		exist       bool
		err         error
	)

	for _, paymentMethod := range paymentMethods {

		err = db.Raw(searchQuery, paymentMethod.Name).Scan(&exist).Error
		if err != nil {
			return fmt.Errorf("failed to check payment methods already exist %w", err)
		}
		if !exist {
			paymentMethod.ID = domain.NewID(domain.PrefixPaymentMethod)
			err = db.Exec(insertQuery, paymentMethod.ID, paymentMethod.Name,
				paymentMethod.MaximumAmount.AmountMinor, paymentMethod.MaximumAmount.Currency).Error
			if err != nil {
				return fmt.Errorf("failed to save payment method %w", err)
			}
		}
		exist = false
	}
	return nil
}

func saveAdmin(db *gorm.DB, email, password string) error {
	var (
		searchQuery = `SELECT COUNT(*) > 0 as exist FROM admins WHERE email = $1`
		exist       bool
		err         error
	)

	err = db.Raw(searchQuery, email).Scan(&exist).Error
	if err != nil {
		// It's okay if the admins table doesn't exist yet (GORM will create it)
		// Just log and continue
		return nil
	}

	// if !exist {
	// 	hashPass, err := utils.GetHashedPassword(password)
	// 	if err != nil {
	// 		return fmt.Errorf("failed to hash password err: %w", err)
	// 	}
	// 	createdAt := time.Now()
	// 	updatedAt := createdAt

	// 	userName := "john_doe"
	// 	email := "john@example.com"
	// 	mobile := "9999999999"
	// 	shop_name := "John's Furniture"
	// 	gstin := "29ABCDE1234F2Z5"
	// 	shop_id := "SHOP12345"
	// 	address_line1 := "123, Main Street"
	// 	address_line2 := "2nd Floor"
	// 	city := "Bangalore"
	// 	state := "Karnataka"
	// 	country := "India"
	// 	pincode := "560001"
	// 	bank_account_number := "123456789012"
	// 	bank_ifsc := "HDFC0001234"
	// 	pan := "ABCDE1234F"
	// 	aadhar := "123412341234"
	// 	agree_to_terms := true
	// 	verified := "pending"
	// 	status := "active"
	// 	err = db.Exec(insertQuery, userName, email, mobile, hashPass, shop_name, gstin, shop_id,
	// 		address_line1, address_line2, city, state, country, pincode,
	// 		bank_account_number, bank_ifsc, pan, aadhar, agree_to_terms,
	// 		verified, status, createdAt, updatedAt).Error
	// 	if err != nil {
	// 		return fmt.Errorf("failed to save admin details %w", err)
	// 	}
	// }
	return nil
}

func SeedCountries(db *gorm.DB) error {
	countries := []domain.Country{
		{CountryName: "India", ISOCode: "IN"},
		{CountryName: "United States", ISOCode: "US"},
		{CountryName: "United Kingdom", ISOCode: "GB"},
		{CountryName: "Canada", ISOCode: "CA"},
		{CountryName: "Australia", ISOCode: "AU"},
		{CountryName: "Germany", ISOCode: "DE"},
		{CountryName: "France", ISOCode: "FR"},
		{CountryName: "Japan", ISOCode: "JP"},
		{CountryName: "China", ISOCode: "CN"},
		{CountryName: "Brazil", ISOCode: "BR"},
	}

	for _, country := range countries {
		if err := db.FirstOrCreate(&country, domain.Country{ISOCode: country.ISOCode}).Error; err != nil {
			return err
		}
	}

	return nil
}
