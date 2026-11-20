package domain

import "time"

type SubscriptionExpiredNudgeSettings struct {
	ID            string    `json:"-" gorm:"primaryKey;type:varchar(16)"`
	Enabled       bool      `json:"enabled" gorm:"not null;default:false"`
	FireHourIST   int       `json:"fire_hour_ist" gorm:"not null;default:9"`
	FireMinuteIST int       `json:"fire_minute_ist" gorm:"not null;default:0"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

const SubscriptionExpiredNudgeSettingsID = "default"

type SubscriptionExpiredNudgeTemplate struct {
	ID        string    `json:"-" gorm:"primaryKey;type:varchar(16)"`
	Title     string    `json:"title" gorm:"type:varchar(200);not null"`
	Body      string    `json:"body" gorm:"type:text;not null"`
	ImageURL  string    `json:"image_url" gorm:"type:text"`
	Route     string    `json:"route" gorm:"type:varchar(200)"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

const SubscriptionExpiredNudgeTemplateID = "default"

type SubscriptionExpiredNudgeCandidate struct {
	UserID    string    `json:"user_id"`
	ShopName  string    `json:"shop_name"`
	OwnerName string    `json:"owner_name"`
	EndDate   time.Time `json:"end_date"`
}

// SubscriptionExpiredNudgeSent is the idempotency ledger: one notification
// per user per calendar day so the daily ticker never double-sends.
type SubscriptionExpiredNudgeSent struct {
	UserID   string    `json:"user_id" gorm:"primaryKey;type:varchar(32)"`
	SentDate string    `json:"sent_date" gorm:"primaryKey;type:char(10)"` // YYYY-MM-DD
	SentAt   time.Time `json:"sent_at" gorm:"autoCreateTime"`
}
