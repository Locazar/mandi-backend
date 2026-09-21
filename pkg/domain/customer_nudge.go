package domain

import "time"

// Customer onboarding nudge slots, in the order they fire each day.
const (
	CustomerNudgeExploreShops = "explore_shops"
	CustomerNudgeSetLocation  = "set_location"
	CustomerNudgeSendEnquiry  = "send_enquiry"
	CustomerNudgeShareApp     = "share_app"
)

// CustomerNudgeTemplateOrder is the fixed slot order — index is the daily slot number.
var CustomerNudgeTemplateOrder = []string{CustomerNudgeExploreShops, CustomerNudgeSetLocation, CustomerNudgeSendEnquiry, CustomerNudgeShareApp}

// CustomerNudgeTemplate is editable copy for one of the four fixed slots.
// Body/Title may use {{name}} (the customer's first name, or "there").
type CustomerNudgeTemplate struct {
	Key      string `json:"key" gorm:"primaryKey;type:varchar(32)"`
	Title    string `json:"title" gorm:"type:varchar(200);not null"`
	Body     string `json:"body" gorm:"type:text;not null"`
	ImageURL string `json:"image_url" gorm:"type:text"`
	// Route is the customer-app screen the tap opens. Empty just opens the app.
	Route     string    `json:"route" gorm:"type:varchar(200)"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// CustomerNudgeSettings is the single-row schedule config. StartsAt is stamped
// whenever the feature is switched on, so only customers who sign up after that
// moment enter the sequence — existing customers never get a catch-up burst.
type CustomerNudgeSettings struct {
	ID           string    `json:"-" gorm:"primaryKey;type:varchar(16)"`
	Enabled      bool      `json:"enabled" gorm:"not null;default:false"`
	GapHours     int       `json:"gap_hours" gorm:"not null;default:2"`
	DurationDays int       `json:"duration_days" gorm:"not null;default:7"`
	StartsAt     time.Time `json:"starts_at"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// CustomerNudgeSent is the idempotency ledger: one row per (user, day, slot)
// processed. Delivered is false when the customer had no reachable device.
type CustomerNudgeSent struct {
	UserID    string    `json:"user_id" gorm:"primaryKey;type:varchar(32)"`
	Day       int       `json:"day" gorm:"primaryKey"`
	Slot      int       `json:"slot" gorm:"primaryKey"`
	Delivered bool      `json:"delivered" gorm:"not null"`
	SentAt    time.Time `json:"sent_at" gorm:"autoCreateTime"`
}

// CustomerNudgeCandidate is a read-only projection of a recent signup.
type CustomerNudgeCandidate struct {
	UserID    string    `json:"user_id"`
	FirstName string    `json:"first_name"`
	StartAt   time.Time `json:"start_at"`
}

// CustomerNudgeStats aggregates delivered nudges for admin-portal.
type CustomerNudgeStats struct {
	TotalSent        int64                     `json:"total_sent"`
	SentToday        int64                     `json:"sent_today"`
	CustomersReached int64                     `json:"customers_reached"`
	NotDelivered     int64                     `json:"not_delivered"`
	BySlot           []OnboardingNudgeSlotStat `json:"by_slot"`
	Last7Days        []OnboardingNudgeDayStat  `json:"last_7_days"`
}
