package db

import (
	"fmt"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/utils"
	"gorm.io/gorm"
)

// To save predefined order statuses on database if its not exist
func saveOrderStatuses(db *gorm.DB) error {

	statuses := []domain.OrderStatusType{
		domain.StatusPaymentPending,
		domain.StatusOrderPlaced,
		domain.StatusOrderCancelled,
		domain.StatusOrderDelivered,
		domain.StatusReturnRequested,
		domain.StatusReturnApproved,
		domain.StatusReturnCancelled,
		domain.StatusOrderReturned,
	}

	var (
		searchQuery = `SELECT CASE WHEN id != 0 THEN 'T' ELSE 'F' END as exist 
		FROM order_statuses WHERE status = $1`
		insertQuery = `INSERT INTO order_statuses (status) VALUES ($1)`
		exist       bool
		err         error
	)

	for _, status := range statuses {

		err = db.Raw(searchQuery, status).Scan(&exist).Error
		if err != nil {
			return fmt.Errorf("failed to check order status already exist err: %w", err)
		}

		if !exist {
			err = db.Exec(insertQuery, status).Error
			if err != nil {
				return fmt.Errorf("failed to save status %w", err)
			}
		}
		exist = false
	}
	return nil
}

// To save predefined payment methods on database if its not exist
func savePaymentMethods(db *gorm.DB) error {
	paymentMethods := []domain.PaymentMethod{
		{
			Name:          domain.CodPayment,
			MaximumAmount: domain.CodMaximumAmount,
		},
		{
			Name:          domain.RazopayPayment,
			MaximumAmount: domain.RazorPayMaximumAmount,
		},
		{
			Name:          domain.StripePayment,
			MaximumAmount: domain.StripeMaximumAmount,
		},
	}

	var (
		searchQuery = `SELECT CASE WHEN id != 0 THEN 'T' ELSE 'F' END as exist FROM payment_methods WHERE name = $1`
		insertQuery = `INSERT INTO payment_methods (name, maximum_amount) VALUES ($1, $2)`
		exist       bool
		err         error
	)

	for _, paymentMethod := range paymentMethods {

		err = db.Raw(searchQuery, paymentMethod.Name).Scan(&exist).Error
		if err != nil {
			return fmt.Errorf("failed to check payment methods already exist %w", err)
		}
		if !exist {
			err = db.Exec(insertQuery, paymentMethod.Name, paymentMethod.MaximumAmount).Error
			if err != nil {
				return fmt.Errorf("failed to save payment method %w", err)
			}
		}
		exist = false
	}
	return nil
}

func saveAdmin(db *gorm.DB, email, userName, password string) error {

	var (
		searchQuery = `SELECT CASE WHEN id != 0 THEN 'T' ELSE 'F' END as exist FROM admins WHERE email = $1`
		insertQuery = `INSERT INTO admins (user_name, email, mobile, password_hash, shop_name, gstin, shop_id,
		address_line1, address_line2, city, state, country, pincode,
		bank_account_number, bank_ifsc, pan, aadhar, agree_to_terms,
		verified, status, created_at, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
		$11, $12, $13, $14, $15, $16, $17, $18, $19, $20, $21)`
		exist bool
		err   error
	)

	err = db.Raw(searchQuery, email).Scan(&exist).Error
	if err != nil {
		return fmt.Errorf("failed to check admin already exist err:%w", err)
	}

	if !exist {
		hashPass, err := utils.GetHashedPassword(password)
		if err != nil {
			return fmt.Errorf("failed to hash password err: %w", err)
		}
		createdAt := time.Now()
		updatedAt := createdAt

		userName := "john_doe"
		email := "john@example.com"
		mobile := "9999999999"
		shop_name := "John's Furniture"
		gstin := "29ABCDE1234F2Z5"
		shop_id := "SHOP12345"
		address_line1 := "123, Main Street"
		address_line2 := "2nd Floor"
		city := "Bangalore"
		state := "Karnataka"
		country := "India"
		pincode := "560001"
		bank_account_number := "123456789012"
		bank_ifsc := "HDFC0001234"
		pan := "ABCDE1234F"
		aadhar := "123412341234"
		agree_to_terms := true
		verified := "pending"
		status := "active"
		err = db.Exec(insertQuery, email, userName, hashPass, shop_name, gstin, shop_id,
			address_line1, address_line2, city, state, mobile, country, pincode,
			bank_account_number, bank_ifsc, pan, aadhar, agree_to_terms,
			verified, status, createdAt, updatedAt).Error
		if err != nil {
			return fmt.Errorf("failed to save admin details %w", err)
		}
	}
	return nil
}
