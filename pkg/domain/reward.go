package domain

import (
	"time"

	"github.com/lib/pq"
)

// RewardProgramConfigID is the fixed id of the reward_program_config singleton
// (seeded by migration 000046).
const RewardProgramConfigID = "rwcfg_default"

type RewardOwnerType string

const (
	RewardOwnerShop     RewardOwnerType = "shop"
	RewardOwnerCustomer RewardOwnerType = "customer"
)

type RewardEntryType string

const (
	RewardEntryPurchaseEarn          RewardEntryType = "purchase_earn"
	RewardEntryReferralBonus         RewardEntryType = "referral_bonus"
	RewardEntryRedeemSubscription    RewardEntryType = "redeem_subscription"
	RewardEntryRedeemAdvertisement   RewardEntryType = "redeem_advertisement"
	RewardEntryCustomerRedeemAtShop  RewardEntryType = "customer_redeem_at_shop"
	RewardEntryRedeemReimbursement   RewardEntryType = "redeem_reimbursement"
	RewardEntryExpiry                RewardEntryType = "expiry"
	RewardEntryReversal              RewardEntryType = "reversal"
	RewardEntryAdminAdjust           RewardEntryType = "admin_adjust"
)

type ShopPurchaseStatus string

const (
	ShopPurchasePending  ShopPurchaseStatus = "pending"
	ShopPurchaseClaimed  ShopPurchaseStatus = "claimed"
	ShopPurchaseRejected ShopPurchaseStatus = "rejected"
	ShopPurchaseExpired  ShopPurchaseStatus = "expired"
)

// RewardFormula is points = base + floor(net_paid / ₹100) × per100, capped at
// max (max <= 0 means uncapped).
type RewardFormula struct {
	BasePoints         int64
	PointsPer100Rupees int64
	MaxPoints          int64
}

// RewardProgramConfig is the admin-managed singleton that drives every rule of
// the rewards program. Values are read at decision time; a purchase snapshots
// the points it was awarded, so edits never rewrite history.
type RewardProgramConfig struct {
	ID                         string        `json:"-" gorm:"primaryKey;type:varchar(32)"`
	ProgramEnabled             bool          `json:"program_enabled"`
	AllowedDiscountPercents    pq.Int64Array `json:"allowed_discount_percents" gorm:"type:int[]"`
	SellerBasePoints           int64         `json:"seller_base_points"`
	SellerPointsPer100Rupees   int64         `json:"seller_points_per_100_rupees" gorm:"column:seller_points_per_100_rupees"`
	SellerMaxPoints            int64         `json:"seller_max_points"`
	CustomerBasePoints         int64         `json:"customer_base_points"`
	CustomerPointsPer100Rupees int64         `json:"customer_points_per_100_rupees" gorm:"column:customer_points_per_100_rupees"`
	CustomerMaxPoints          int64         `json:"customer_max_points"`
	MinBillPaise               int64         `json:"min_bill_paise"`
	GPSRadiusM                 int           `json:"gps_radius_m" gorm:"column:gps_radius_m"`
	RepeatWindowHours          int           `json:"repeat_window_hours"`
	ClaimExpiryHours           int           `json:"claim_expiry_hours"`
	MaxClaimsPerShopPerDay     int           `json:"max_claims_per_shop_per_day"`
	MaxCustomerPointsPerDay    int64         `json:"max_customer_points_per_day"`
	PointValuePaise            int64         `json:"point_value_paise"`
	SellerMinRedeemBalance     int64         `json:"seller_min_redeem_balance"`
	CustomerMinRedeemBalance   int64         `json:"customer_min_redeem_balance"`
	CustomerMaxRedeemPctBP     int           `json:"customer_max_redeem_pct_bp" gorm:"column:customer_max_redeem_pct_bp"`
	PointsExpiryMonths         int           `json:"points_expiry_months"`
	ReferralBonusPoints        int64         `json:"referral_bonus_points"`
	MaxReferralsPerMonth       int           `json:"max_referrals_per_month"`
	UpdatedBy                  string        `json:"updated_by"`
	UpdatedAt                  time.Time     `json:"updated_at" gorm:"autoUpdateTime"`
}

func (RewardProgramConfig) TableName() string { return "reward_program_config" }

func (c RewardProgramConfig) SellerFormula() RewardFormula {
	return RewardFormula{BasePoints: c.SellerBasePoints, PointsPer100Rupees: c.SellerPointsPer100Rupees, MaxPoints: c.SellerMaxPoints}
}

func (c RewardProgramConfig) CustomerFormula() RewardFormula {
	return RewardFormula{BasePoints: c.CustomerBasePoints, PointsPer100Rupees: c.CustomerPointsPer100Rupees, MaxPoints: c.CustomerMaxPoints}
}

// ShopRewardSettings is a shop's opt-in to the program (1:1 with shop_details).
type ShopRewardSettings struct {
	ShopID          string     `json:"shop_id" gorm:"primaryKey;type:varchar(32)"`
	AdminID         string     `json:"admin_id"`
	DiscountEnabled bool       `json:"discount_enabled"`
	DiscountPercent int        `json:"discount_percent"`
	OptedInAt       *time.Time `json:"opted_in_at"`
	UpdatedAt       time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
}

func (ShopRewardSettings) TableName() string { return "shop_reward_settings" }

// RewardAccount holds one owner's cached balance. It is only ever changed in
// the same transaction as a RewardLedgerEntry insert.
type RewardAccount struct {
	ID             string          `json:"id" gorm:"primaryKey;type:varchar(32)"`
	OwnerType      RewardOwnerType `json:"owner_type"`
	OwnerID        string          `json:"owner_id"`
	BalancePoints  int64           `json:"balance_points"`
	LifetimeEarned int64           `json:"lifetime_earned"`
	LifetimeSpent  int64           `json:"lifetime_spent"`
	CreatedAt      time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt      time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
}

func (RewardAccount) TableName() string { return "reward_accounts" }

// RewardLedgerEntry is one signed movement of points. Positive entries are
// "lots": RemainingPoints tracks how much of the lot is still unspent and
// unexpired, consumed oldest-expiry-first by debits.
type RewardLedgerEntry struct {
	ID               string          `json:"id" gorm:"primaryKey;type:varchar(32)"`
	AccountID        string          `json:"account_id"`
	DeltaPoints      int64           `json:"delta_points"`
	EntryType        RewardEntryType `json:"entry_type"`
	RefType          string          `json:"ref_type"`
	RefID            string          `json:"ref_id"`
	RemainingPoints  int64           `json:"remaining_points"`
	ExpiresAt        *time.Time      `json:"expires_at"`
	ExpiryRemindedAt *time.Time      `json:"-"`
	Note             string          `json:"note"`
	CreatedBy        string          `json:"created_by"`
	CreatedAt        time.Time       `json:"created_at" gorm:"autoCreateTime"`
}

func (RewardLedgerEntry) TableName() string { return "reward_ledger_entries" }

// ShopPurchase is one customer scan-and-buy at a shop.
type ShopPurchase struct {
	ID                       string             `json:"id" gorm:"primaryKey;type:varchar(32)"`
	ShopID                   string             `json:"shop_id"`
	CustomerID               string             `json:"customer_id"`
	ClientRequestID          string             `json:"-"`
	BillAmountPaise          int64              `json:"bill_amount_paise"`
	DiscountPercent          int                `json:"discount_percent"`
	DiscountPaise            int64              `json:"discount_paise"`
	PointsRedeemed           int64              `json:"points_redeemed"`
	PointsRedeemedValuePaise int64              `json:"points_redeemed_value_paise"`
	NetPaidPaise             int64              `json:"net_paid_paise"`
	CustomerLat              float64            `json:"-" gorm:"column:customer_lat"`
	CustomerLng              float64            `json:"-" gorm:"column:customer_lng"`
	DistanceM                int                `json:"distance_m"`
	Status                   ShopPurchaseStatus `json:"status"`
	SellerPoints             int64              `json:"seller_points"`
	CustomerPoints           int64              `json:"customer_points"`
	ExpiresAt                time.Time          `json:"expires_at"`
	ClaimedAt                *time.Time         `json:"claimed_at"`
	DecidedAt                *time.Time         `json:"decided_at"`
	DecidedBy                string             `json:"-"`
	RejectReason             string             `json:"reject_reason"`
	CreatedAt                time.Time          `json:"created_at" gorm:"autoCreateTime"`
}

func (ShopPurchase) TableName() string { return "shop_purchases" }

// ShopPurchaseView is a purchase joined with what the seller inbox / admin list
// shows: masked customer, shop name, and this customer's history at the shop.
type ShopPurchaseView struct {
	ShopPurchase        `gorm:"embedded"`
	CustomerName        string     `json:"customer_name"`
	CustomerPhoneLast4  string     `json:"customer_phone_last4" gorm:"column:customer_phone_last4"`
	ShopName            string     `json:"shop_name"`
	VisitCount          int64      `json:"visit_count"`
	PrevBillAmountPaise *int64     `json:"prev_bill_amount_paise"`
	PrevPurchaseAt      *time.Time `json:"prev_purchase_at"`
}

// ShopPurchaseFilter narrows ListPurchases; empty fields are ignored.
type ShopPurchaseFilter struct {
	ShopID     string
	CustomerID string
	Status     ShopPurchaseStatus
}

// RewardShop is the slice of shop_details (+ owner's mobile) the rewards rules need.
type RewardShop struct {
	ID          string
	AdminID     string
	ShopName    string
	ShopStatus  ShopStatusType
	Latitude    float64
	Longitude   float64
	OwnerMobile string
}

// RewardCustomer is the slice of users the rewards rules need.
type RewardCustomer struct {
	ID        string
	FirstName string
	LastName  string
	Phone     string
}
