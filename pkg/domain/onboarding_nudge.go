package domain

import "time"

// OnboardingNudgeTemplateKey identifies one of the four fixed onboarding
// reminder slots, in the order they fire each day (2 hours apart).
const (
	NudgeAddProducts   = "add_products"
	NudgeUpdatePhoto   = "update_photo"
	NudgeUpdateAddress = "update_address"
	NudgeViewShop      = "view_shop"
)

// NudgeTemplateOrder is the fixed slot order — index is the daily slot number.
var NudgeTemplateOrder = []string{NudgeAddProducts, NudgeUpdatePhoto, NudgeUpdateAddress, NudgeViewShop}

// OnboardingNudgeTemplate is editable copy for one of the four fixed slots.
// Rows are seeded once with sensible defaults (see repository) and never
// created/deleted through the API — only their Title/Body/ImageURL change.
type OnboardingNudgeTemplate struct {
	Key      string `json:"key" gorm:"primaryKey;type:varchar(32)"`
	Title    string `json:"title" gorm:"type:varchar(200);not null"`
	Body     string `json:"body" gorm:"type:text;not null"`
	ImageURL string `json:"image_url" gorm:"type:text"`
	// Route is where tapping the notification lands in the seller app (e.g.
	// "/home?tab=2"), matching admin-portal's action-link.ts catalog. Empty
	// means the tap just opens the app. Only argument-free destinations make
	// sense here — this one template's route applies to every shop it's ever
	// sent to, so a destination needing a specific shop/inquiry id would be
	// wrong for all but one recipient.
	Route     string    `json:"route" gorm:"type:varchar(200)"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// OnboardingNudgeSettings is a single-row schedule config, edited from
// admin-portal. GapHours spaces the 4 daily sends; DurationDays is how many
// days after go-live the sequence runs before stopping.
type OnboardingNudgeSettings struct {
	ID           string    `json:"-" gorm:"primaryKey;type:varchar(16)"`
	Enabled      bool      `json:"enabled" gorm:"not null;default:true"`
	GapHours     int       `json:"gap_hours" gorm:"not null;default:2"`
	DurationDays int       `json:"duration_days" gorm:"not null;default:7"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// OnboardingNudgeSettingsID is the fixed single row id for OnboardingNudgeSettings.
const OnboardingNudgeSettingsID = "default"

// ShopOnboardingNudgeAnchor records the moment a shop went live — the day-0
// anchor the sweep counts every send's due time from. Written once, the
// first time a shop is approved; a later re-approval leaves it untouched so
// the sequence never restarts for the same shop.
type ShopOnboardingNudgeAnchor struct {
	ShopID    string    `json:"shop_id" gorm:"primaryKey;type:varchar(32)"`
	GoLiveAt  time.Time `json:"go_live_at" gorm:"not null"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime"`
}

// OnboardingNudgeCandidate is a read-only join projection (never migrated —
// it's not a table of its own) used by the sweep: a shop's go-live anchor
// plus the name/city its "view your shop" link is built from.
type OnboardingNudgeCandidate struct {
	ShopID   string    `json:"shop_id"`
	GoLiveAt time.Time `json:"go_live_at"`
	ShopName string    `json:"shop_name"`
	City     string    `json:"city"`
}

// ShopOnboardingNudgeSent is the idempotency ledger: one row per
// (shop, day, slot) actually delivered, so a sweep that runs more than once
// (or catches up after downtime) never double-sends the same nudge.
type ShopOnboardingNudgeSent struct {
	ShopID string    `json:"shop_id" gorm:"primaryKey;type:varchar(32)"`
	Day    int       `json:"day" gorm:"primaryKey"`
	Slot   int       `json:"slot" gorm:"primaryKey"`
	SentAt time.Time `json:"sent_at" gorm:"autoCreateTime"`
}
