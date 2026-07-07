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
	ID   string `json:"id" gorm:"primaryKey;type:varchar(32)"`
	Name string `json:"name" gorm:"unique;not null"`
	// PriceMonthly is the total price (Money) charged once for the plan's
	// full DurationDays window. The name is historical from when every plan
	// was billed monthly; the value is NOT a per-month figure for the
	// duration-based plans (3 Months / 6 Months).
	PriceMonthly Money `json:"price_monthly" gorm:"embedded;embeddedPrefix:price_monthly_"`
	DurationDays uint  `json:"duration_days" gorm:"not null;default:30"`
	IsActive     bool  `json:"is_active" gorm:"not null;default:true"`
}

type SubscriptionOrder struct {
	ID     string `json:"id" gorm:"primaryKey;type:varchar(32)"`
	UserID string `json:"user_id" gorm:"type:varchar(32);not null;index"`
	PlanID string `json:"plan_id" gorm:"type:varchar(32);not null"`
	Price  Money  `json:"price" gorm:"embedded;embeddedPrefix:price_"`
	// GST is included in Price (prices are tax-inclusive). GSTRateBasisPoints is
	// the rate in effect when the order was created (1800 = 18.00%), snapshotted
	// so historical "GST collected" reporting stays correct across yearly rate
	// changes. GSTAmount is the GST portion of Price, computed once at creation.
	GSTRateBasisPoints int                `json:"gst_rate_basis_points" gorm:"not null;default:0"`
	GSTAmount          Money              `json:"gst_amount" gorm:"embedded;embeddedPrefix:gst_amount_"`
	RazorpayOrderID    string             `json:"razorpay_order_id" gorm:"uniqueIndex;not null"`
	RazorpayPaymentID  *string            `json:"razorpay_payment_id" gorm:"uniqueIndex"`
	Status             SubscriptionStatus `json:"status" gorm:"not null;default:'created'"`
	CreatedAt          time.Time          `json:"created_at" gorm:"autoCreateTime"`
	PaidAt             *time.Time         `json:"paid_at"`
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
