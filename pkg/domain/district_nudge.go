package domain

import "time"

// DistrictNudgeSettings is a single-row schedule config for the district
// growth nudge: nudging a merchant with a stronger nearby peer shop's
// product count, to drive catalogue growth. State scopes which shops are
// even considered (ships narrow — "Rajasthan" — by design, not because the
// geo math only works there); RadiusKm is how far "nearby" reaches.
type DistrictNudgeSettings struct {
	ID        string    `json:"-" gorm:"primaryKey;type:varchar(16)"`
	Enabled   bool      `json:"enabled" gorm:"not null;default:false"`
	State     string    `json:"state" gorm:"type:varchar(100);not null;default:'Rajasthan'"`
	RadiusKm  float64   `json:"radius_km" gorm:"not null;default:100"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// DistrictNudgeSettingsID is the fixed single row id for DistrictNudgeSettings.
const DistrictNudgeSettingsID = "default"

// DistrictNudgeTemplate is the single editable push sent to a merchant about
// the strongest nearby peer shop. Seeded once with sensible defaults —
// admin only needs to open the page and edit if the default doesn't fit.
type DistrictNudgeTemplate struct {
	ID       string `json:"-" gorm:"primaryKey;type:varchar(16)"`
	Title    string `json:"title" gorm:"type:varchar(200);not null"`
	Body     string `json:"body" gorm:"type:text;not null"`
	ImageURL string `json:"image_url" gorm:"type:text"`
	// Route is where tapping the notification lands in the seller app (e.g.
	// "/home?tab=2"), matching admin-portal's action-link.ts catalog.
	Route     string    `json:"route" gorm:"type:varchar(200)"`
	UpdatedAt time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// DistrictNudgeTemplateID is the fixed single row id for DistrictNudgeTemplate.
const DistrictNudgeTemplateID = "default"

// DistrictNudgeCandidate is a read-only join projection (never migrated —
// it's not a table of its own): one target shop paired with the single
// strongest active peer within radius that it hasn't already been nudged
// about. The sweep finds one row per target shop, not a global "top shop
// per area" — overlapping radii around different shops would make a strict
// partition ambiguous, so each shop's own neighbourhood decides its own
// best peer instead.
type DistrictNudgeCandidate struct {
	TargetShopID    string `json:"target_shop_id"`
	TargetShopName  string `json:"target_shop_name"`
	TargetOwnerName string `json:"target_owner_name"`
	TargetCity      string `json:"target_city"`
	TopShopID       string `json:"top_shop_id"`
	TopShopName     string `json:"top_shop_name"`
	TopShopCity     string `json:"top_shop_city"`
	TopProductCount int    `json:"top_product_count"`
}

// DistrictNudgeSent is the idempotency ledger: once a target shop has been
// nudged about a specific peer, it is never nudged about that exact pairing
// again. If a different shop later becomes its strongest nearby peer,
// that's a new, distinct pairing and gets its own nudge.
type DistrictNudgeSent struct {
	TargetShopID string    `json:"target_shop_id" gorm:"primaryKey;type:varchar(32)"`
	TopShopID    string    `json:"top_shop_id" gorm:"primaryKey;type:varchar(32)"`
	SentAt       time.Time `json:"sent_at" gorm:"autoCreateTime"`
}
