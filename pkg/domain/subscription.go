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
	ID           string `json:"id" gorm:"primaryKey;type:varchar(32)"`
	Name         string `json:"name" gorm:"unique;not null"`
	PriceMonthly Money  `json:"price_monthly" gorm:"embedded;embeddedPrefix:price_monthly_"`
	DurationDays uint   `json:"duration_days" gorm:"not null;default:30"`
	IsActive     bool   `json:"is_active" gorm:"not null;default:true"`
}

type SubscriptionOrder struct {
	ID                string     `json:"id" gorm:"primaryKey;type:varchar(32)"`
	UserID            string     `json:"user_id" gorm:"type:varchar(32);not null;index"`
	PlanID            string     `json:"plan_id" gorm:"type:varchar(32);not null"`
	Price             Money      `json:"price" gorm:"embedded;embeddedPrefix:price_"`
	RazorpayOrderID   string     `json:"razorpay_order_id" gorm:"uniqueIndex;not null"`
	RazorpayPaymentID *string    `json:"razorpay_payment_id" gorm:"uniqueIndex"`
	Status            SubscriptionStatus `json:"status" gorm:"not null;default:'created'"`
	CreatedAt         time.Time  `json:"created_at" gorm:"autoCreateTime"`
	PaidAt            *time.Time `json:"paid_at"`
}

type UserSubscription struct {
	ID                  string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	UserID              string    `json:"user_id" gorm:"type:varchar(32);not null;index"`
	PlanID              string    `json:"plan_id" gorm:"type:varchar(32);not null"`
	SubscriptionOrderID *string   `json:"subscription_order_id" gorm:"type:varchar(32)"`
	IsTrial             bool      `json:"is_trial" gorm:"not null;default:false"`
	StartDate           time.Time `json:"start_date" gorm:"not null"`
	EndDate             time.Time `json:"end_date" gorm:"not null"`
	IsActive            bool      `json:"is_active" gorm:"not null;default:true"`
}
