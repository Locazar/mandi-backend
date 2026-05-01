package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type subscriptionDatabase struct {
	db *gorm.DB
}

func NewSubscriptionRepository(db *gorm.DB) interfaces.SubscriptionRepository {
	return &subscriptionDatabase{db: db}
}

func (r *subscriptionDatabase) CreateSubscriptionOrder(ctx context.Context, order domain.SubscriptionOrder) (domain.SubscriptionOrder, error) {
	err := r.db.Create(&order).Error
	return order, err
}

func (r *subscriptionDatabase) FindSubscriptionOrderByRazorpayOrderID(ctx context.Context, razorpayOrderID string) (domain.SubscriptionOrder, error) {
	var order domain.SubscriptionOrder
	err := r.db.Where("razorpay_order_id = ?", razorpayOrderID).First(&order).Error
	return order, err
}

func (r *subscriptionDatabase) FindSubscriptionOrderByRazorpayPaymentID(ctx context.Context, paymentID string) (domain.SubscriptionOrder, error) {
	var order domain.SubscriptionOrder
	err := r.db.Where("razorpay_payment_id = ?", paymentID).First(&order).Error
	return order, err
}

// UpdateSubscriptionOrderToPaid atomically updates an order from "created" to "paid".
// The WHERE status='created' guard prevents double-activation races.
func (r *subscriptionDatabase) UpdateSubscriptionOrderToPaid(ctx context.Context, orderID uint, razorpayPaymentID string) error {
	now := time.Now()
	result := r.db.Model(&domain.SubscriptionOrder{}).
		Where("id = ? AND status = ?", orderID, domain.SubStatusCreated).
		Updates(map[string]interface{}{
			"status":              domain.SubStatusPaid,
			"razorpay_payment_id": &razorpayPaymentID,
			"paid_at":             now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("order %d not in 'created' state or not found", orderID)
	}
	return nil
}

func (r *subscriptionDatabase) FindSubscriptionPlanByID(ctx context.Context, planID uint) (domain.SubscriptionPlan, error) {
	var plan domain.SubscriptionPlan
	err := r.db.First(&plan, planID).Error
	return plan, err
}

func (r *subscriptionDatabase) FindActiveSubscriptionByUserID(ctx context.Context, userID uint) (domain.UserSubscription, error) {
	var sub domain.UserSubscription
	err := r.db.Where("user_id = ? AND is_active = ? AND end_date > ?", userID, true, time.Now()).First(&sub).Error
	return sub, err
}

func (r *subscriptionDatabase) ActivateSubscription(ctx context.Context, sub domain.UserSubscription) error {
	return r.db.Create(&sub).Error
}

func (r *subscriptionDatabase) Transaction(fn func(repo interfaces.SubscriptionRepository) error) error {
	trx := r.db.Begin()
	repo := &subscriptionDatabase{db: trx}

	if err := fn(repo); err != nil {
		trx.Rollback()
		return fmt.Errorf("failed to complete transaction: %w", err)
	}

	return trx.Commit().Error
}
