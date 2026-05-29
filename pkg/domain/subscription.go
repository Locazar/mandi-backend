package domain

import "time"

// Subscription order statuses
const (
	SubStatusCreated = "created"
	SubStatusPaid    = "paid"
	SubStatusFailed  = "failed"
	SubStatusExpired = "expired"
)

type SubscriptionPlan struct {
	ID           uint   `json:"id" gorm:"primaryKey;not null"`
	Name         string `json:"name" gorm:"unique;not null"`
	PriceMonthly Money  `json:"price_monthly" gorm:"embedded;embeddedPrefix:price_monthly_"`
	DurationDays uint   `json:"duration_days" gorm:"not null;default:30"`
	IsActive     bool   `json:"is_active" gorm:"not null;default:true"`
}

type SubscriptionOrder struct {
	ID                uint       `json:"id" gorm:"primaryKey;not null"`
	UserID            uint       `json:"user_id" gorm:"not null;index"`
	PlanID            uint       `json:"plan_id" gorm:"not null"`
	Price             Money      `json:"price" gorm:"embedded;embeddedPrefix:price_"`
	RazorpayOrderID   string     `json:"razorpay_order_id" gorm:"uniqueIndex;not null"`
	RazorpayPaymentID *string    `json:"razorpay_payment_id" gorm:"uniqueIndex"`
	Status            SubscriptionStatus `json:"status" gorm:"not null;default:'created'"`
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	PaidAt            *time.Time `json:"paid_at"`
}

type UserSubscription struct {
	ID                  uint      `json:"id" gorm:"primaryKey;not null"`
	UserID              uint      `json:"user_id" gorm:"not null;index"`
	PlanID              uint      `json:"plan_id" gorm:"not null"`
	SubscriptionOrderID *uint     `json:"subscription_order_id"`
	IsTrial             bool      `json:"is_trial" gorm:"not null;default:false"`
	StartDate           time.Time `json:"start_date" gorm:"not null"`
	EndDate             time.Time `json:"end_date" gorm:"not null"`
	IsActive            bool      `json:"is_active" gorm:"not null;default:true"`
}
