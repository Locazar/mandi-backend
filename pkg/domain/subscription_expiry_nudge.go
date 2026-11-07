package domain

import "time"

type SubscriptionExpiryNudgeSettings struct {
	ID               string    `json:"-" gorm:"primaryKey;type:varchar(16)"`
	Enabled          bool      `json:"enabled" gorm:"not null;default:false"`
	DaysBeforeExpiry int       `json:"days_before_expiry" gorm:"not null;default:5"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

const SubscriptionExpiryNudgeSettingsID = "default"

type SubscriptionExpiryNudgeTemplate struct {
	ID        string    `json:"-" gorm:"primaryKey;type:varchar(16)"`
	Title     string    `json:"title" gorm:"type:varchar(200);not null"`
	Body      string    `json:"body" gorm:"type:text;not null"`
	ImageURL  string    `json:"image_url" gorm:"type:text"`
	Route     string    `json:"route" gorm:"type:varchar(200)"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

const SubscriptionExpiryNudgeTemplateID = "default"

type SubscriptionExpiryNudgeCandidate struct {
	UserID    string    `json:"user_id"`
	ShopName  string    `json:"shop_name"`
	OwnerName string    `json:"owner_name"`
	EndDate   time.Time `json:"end_date"`
}

// SubscriptionExpiryNudgeSent is the idempotency ledger: one notification
// per user per calendar day so the daily ticker never double-sends.
type SubscriptionExpiryNudgeSent struct {
	UserID   string    `json:"user_id" gorm:"primaryKey;type:varchar(32)"`
	SentDate string    `json:"sent_date" gorm:"primaryKey;type:char(10)"` // YYYY-MM-DD
	SentAt   time.Time `json:"sent_at" gorm:"autoCreateTime"`
}
