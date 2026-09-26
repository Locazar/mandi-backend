package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const defaultRewardPageSize = 25

type rewardDatabase struct {
	db *gorm.DB
}

func NewRewardRepository(db *gorm.DB) interfaces.RewardRepository {
	return &rewardDatabase{db: db}
}

func (r *rewardDatabase) InTx(ctx context.Context, fn func(interfaces.RewardRepository) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(&rewardDatabase{db: tx})
	})
}

func rewardPage(q *gorm.DB, p request.Pagination) *gorm.DB {
	limit := int(p.Limit)
	if limit <= 0 {
		limit = defaultRewardPageSize
	}
	return q.Limit(limit).Offset(int(p.Offset))
}

// ── config ──────────────────────────────────────────────────────────────────

func (r *rewardDatabase) GetConfig(ctx context.Context) (domain.RewardProgramConfig, error) {
	var cfg domain.RewardProgramConfig
	err := r.db.WithContext(ctx).Where("id = ?", domain.RewardProgramConfigID).First(&cfg).Error
	return cfg, err
}

func (r *rewardDatabase) UpdateConfig(ctx context.Context, cfg domain.RewardProgramConfig) (domain.RewardProgramConfig, error) {
	cfg.ID = domain.RewardProgramConfigID
	// Select("*") so false/0 values are written; only "id" omitted (omitting
	// updated_at would disable its autoUpdateTime tracking).
	err := r.db.WithContext(ctx).Model(&domain.RewardProgramConfig{}).
		Where("id = ?", domain.RewardProgramConfigID).
		Select("*").Omit("id").Updates(&cfg).Error
	if err != nil {
		return domain.RewardProgramConfig{}, err
	}
	return r.GetConfig(ctx)
}

// ── shop / customer lookups ─────────────────────────────────────────────────

const rewardShopSelect = `
	SELECT s.id, s.admin_id, COALESCE(s.shop_name, '') AS shop_name,
	       COALESCE(s.shop_status, '') AS shop_status,
	       COALESCE(s.latitude, 0) AS latitude, COALESCE(s.longitude, 0) AS longitude,
	       COALESCE(a.mobile, '') AS owner_mobile
	FROM shop_details s
	JOIN admins a ON a.id = s.admin_id
	WHERE s.deleted_at IS NULL AND `

func (r *rewardDatabase) scanShop(ctx context.Context, where string, arg string) (domain.RewardShop, error) {
	var shop domain.RewardShop
	res := r.db.WithContext(ctx).Raw(rewardShopSelect+where+" LIMIT 1", arg).Scan(&shop)
	if res.Error != nil {
		return domain.RewardShop{}, res.Error
	}
	if res.RowsAffected == 0 {
		return domain.RewardShop{}, gorm.ErrRecordNotFound
	}
	return shop, nil
}

func (r *rewardDatabase) GetShop(ctx context.Context, shopID string) (domain.RewardShop, error) {
	return r.scanShop(ctx, "s.id = ?", shopID)
}

func (r *rewardDatabase) GetShopByAdminID(ctx context.Context, adminID string) (domain.RewardShop, error) {
	return r.scanShop(ctx, "s.admin_id = ?", adminID)
}

func (r *rewardDatabase) GetCustomer(ctx context.Context, userID string) (domain.RewardCustomer, error) {
	var c domain.RewardCustomer
	res := r.db.WithContext(ctx).Raw(`
		SELECT id, COALESCE(first_name, '') AS first_name, COALESCE(last_name, '') AS last_name,
		       COALESCE(phone, '') AS phone
		FROM users WHERE id = ? LIMIT 1`, userID).Scan(&c)
	if res.Error != nil {
		return domain.RewardCustomer{}, res.Error
	}
	if res.RowsAffected == 0 {
		return domain.RewardCustomer{}, gorm.ErrRecordNotFound
	}
	return c, nil
}

// ── shop settings ───────────────────────────────────────────────────────────

func (r *rewardDatabase) GetShopSettings(ctx context.Context, shopID string) (domain.ShopRewardSettings, error) {
	var s domain.ShopRewardSettings
	err := r.db.WithContext(ctx).Where("shop_id = ?", shopID).First(&s).Error
	return s, err
}

func (r *rewardDatabase) UpsertShopSettings(ctx context.Context, s domain.ShopRewardSettings) (domain.ShopRewardSettings, error) {
	s.UpdatedAt = time.Now()
	err := r.db.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "shop_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"admin_id", "discount_enabled", "discount_percent", "opted_in_at", "updated_at"}),
	}).Create(&s).Error
	if err != nil {
		return domain.ShopRewardSettings{}, err
	}
	return r.GetShopSettings(ctx, s.ShopID)
}

// ── accounts & ledger ───────────────────────────────────────────────────────

func (r *rewardDatabase) GetOrCreateAccount(ctx context.Context, ownerType domain.RewardOwnerType, ownerID string) (domain.RewardAccount, error) {
	acct := domain.RewardAccount{OwnerType: ownerType, OwnerID: ownerID}
	if err := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&acct).Error; err != nil {
		return domain.RewardAccount{}, err
	}
	var got domain.RewardAccount
	err := r.db.WithContext(ctx).Where("owner_type = ? AND owner_id = ?", ownerType, ownerID).First(&got).Error
	return got, err
}

func (r *rewardDatabase) GetAccountByID(ctx context.Context, accountID string) (domain.RewardAccount, error) {
	var a domain.RewardAccount
	err := r.db.WithContext(ctx).Where("id = ?", accountID).First(&a).Error
	return a, err
}

func (r *rewardDatabase) LockAccount(ctx context.Context, accountID string) (domain.RewardAccount, error) {
	var a domain.RewardAccount
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", accountID).First(&a).Error
	return a, err
}

func (r *rewardDatabase) AdjustAccountBalance(ctx context.Context, accountID string, delta, earnedDelta, spentDelta int64) error {
	return r.db.WithContext(ctx).Exec(`
		UPDATE reward_accounts
		SET balance_points = balance_points + ?, lifetime_earned = lifetime_earned + ?,
		    lifetime_spent = lifetime_spent + ?, updated_at = NOW()
		WHERE id = ?`, delta, earnedDelta, spentDelta, accountID).Error
}

func (r *rewardDatabase) InsertLedgerEntry(ctx context.Context, e domain.RewardLedgerEntry) (bool, error) {
	res := r.db.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&e)
	if res.Error != nil {
		return false, res.Error
	}
	return res.RowsAffected == 1, nil
}

func (r *rewardDatabase) ListOpenLots(ctx context.Context, accountID string) ([]domain.RewardLedgerEntry, error) {
	lots := []domain.RewardLedgerEntry{}
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("account_id = ? AND remaining_points > 0", accountID).
		Order("expires_at ASC NULLS LAST, created_at ASC, id ASC").
		Find(&lots).Error
	return lots, err
}

func (r *rewardDatabase) GetLotForUpdate(ctx context.Context, entryID string) (domain.RewardLedgerEntry, error) {
	var e domain.RewardLedgerEntry
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", entryID).First(&e).Error
	return e, err
}

func (r *rewardDatabase) SetLotRemaining(ctx context.Context, entryID string, remaining int64) error {
	return r.db.WithContext(ctx).Model(&domain.RewardLedgerEntry{}).
		Where("id = ?", entryID).Update("remaining_points", remaining).Error
}

func (r *rewardDatabase) ListLedger(ctx context.Context, accountID string, p request.Pagination) ([]domain.RewardLedgerEntry, error) {
	entries := []domain.RewardLedgerEntry{}
	err := rewardPage(r.db.WithContext(ctx).Where("account_id = ?", accountID).
		Order("created_at DESC, id DESC"), p).Find(&entries).Error
	return entries, err
}

func (r *rewardDatabase) SumExpiringPoints(ctx context.Context, accountID string, before time.Time) (int64, *time.Time, error) {
	var row struct {
		Total    int64
		Earliest *time.Time
	}
	err := r.db.WithContext(ctx).Raw(`
		SELECT COALESCE(SUM(remaining_points), 0) AS total, MIN(expires_at) AS earliest
		FROM reward_ledger_entries
		WHERE account_id = ? AND remaining_points > 0 AND expires_at IS NOT NULL AND expires_at <= ?`,
		accountID, before).Scan(&row).Error
	return row.Total, row.Earliest, err
}

func (r *rewardDatabase) ListExpiredLots(ctx context.Context, now time.Time, limit int) ([]domain.RewardLedgerEntry, error) {
	lots := []domain.RewardLedgerEntry{}
	err := r.db.WithContext(ctx).
		Where("remaining_points > 0 AND expires_at IS NOT NULL AND expires_at <= ?", now).
		Order("expires_at ASC").Limit(limit).Find(&lots).Error
	return lots, err
}

func (r *rewardDatabase) ListLotsDueReminder(ctx context.Context, from, to time.Time, limit int) ([]domain.RewardLedgerEntry, error) {
	lots := []domain.RewardLedgerEntry{}
	err := r.db.WithContext(ctx).
		Where("remaining_points > 0 AND expiry_reminded_at IS NULL AND expires_at > ? AND expires_at <= ?", from, to).
		Order("account_id, expires_at").Limit(limit).Find(&lots).Error
	return lots, err
}

func (r *rewardDatabase) MarkLotsReminded(ctx context.Context, entryIDs []string, at time.Time) error {
	if len(entryIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&domain.RewardLedgerEntry{}).
		Where("id IN ?", entryIDs).Update("expiry_reminded_at", at).Error
}

// ── purchases ───────────────────────────────────────────────────────────────

// LockCustomerShopPair takes a transaction-scoped advisory lock so concurrent
// submits for the same customer+shop serialise; must run inside InTx.
func (r *rewardDatabase) LockCustomerShopPair(ctx context.Context, customerID, shopID string) error {
	return r.db.WithContext(ctx).Exec(`SELECT pg_advisory_xact_lock(hashtextextended(?, 0))`,
		"reward_pair:"+customerID+"|"+shopID).Error
}

func (r *rewardDatabase) findOnePurchase(q *gorm.DB) (*domain.ShopPurchase, error) {
	var p domain.ShopPurchase
	err := q.First(&p).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *rewardDatabase) FindPurchaseByClientRequest(ctx context.Context, customerID, clientRequestID string) (*domain.ShopPurchase, error) {
	return r.findOnePurchase(r.db.WithContext(ctx).
		Where("customer_id = ? AND client_request_id = ?", customerID, clientRequestID))
}

func (r *rewardDatabase) FindBlockingPurchase(ctx context.Context, customerID, shopID string, since, now time.Time) (*domain.ShopPurchase, error) {
	return r.findOnePurchase(r.db.WithContext(ctx).
		Where("customer_id = ? AND shop_id = ? AND created_at > ?", customerID, shopID, since).
		Where("status = ? OR (status = ? AND expires_at > ?)", domain.ShopPurchaseClaimed, domain.ShopPurchasePending, now).
		Order("created_at DESC"))
}

func (r *rewardDatabase) CountOpenOrClaimedCreatedSince(ctx context.Context, shopID string, since, now time.Time) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&domain.ShopPurchase{}).
		Where("shop_id = ? AND created_at >= ?", shopID, since).
		Where("status = ? OR (status = ? AND expires_at > ?)", domain.ShopPurchaseClaimed, domain.ShopPurchasePending, now).
		Count(&n).Error
	return n, err
}

func (r *rewardDatabase) CountClaimedSince(ctx context.Context, shopID string, since time.Time) (int64, error) {
	var n int64
	err := r.db.WithContext(ctx).Model(&domain.ShopPurchase{}).
		Where("shop_id = ? AND status = ? AND claimed_at >= ?", shopID, domain.ShopPurchaseClaimed, since).
		Count(&n).Error
	return n, err
}

func (r *rewardDatabase) CustomerShopHistory(ctx context.Context, customerID, shopID, excludePurchaseID string) (int64, *domain.ShopPurchase, error) {
	base := func() *gorm.DB {
		return r.db.WithContext(ctx).Model(&domain.ShopPurchase{}).
			Where("customer_id = ? AND shop_id = ? AND status = ? AND id <> ?", customerID, shopID, domain.ShopPurchaseClaimed, excludePurchaseID)
	}
	var visits int64
	if err := base().Count(&visits).Error; err != nil {
		return 0, nil, err
	}
	last, err := r.findOnePurchase(base().Order("created_at DESC"))
	return visits, last, err
}

func (r *rewardDatabase) CreatePurchase(ctx context.Context, p domain.ShopPurchase) (domain.ShopPurchase, error) {
	err := r.db.WithContext(ctx).Create(&p).Error
	return p, err
}

func (r *rewardDatabase) LockPurchase(ctx context.Context, purchaseID string) (domain.ShopPurchase, error) {
	var p domain.ShopPurchase
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", purchaseID).First(&p).Error
	return p, err
}

func (r *rewardDatabase) UpdatePurchaseDecision(ctx context.Context, p domain.ShopPurchase) error {
	return r.db.WithContext(ctx).Model(&domain.ShopPurchase{}).Where("id = ?", p.ID).
		Updates(map[string]interface{}{
			"status":          p.Status,
			"seller_points":   p.SellerPoints,
			"customer_points": p.CustomerPoints,
			"claimed_at":      p.ClaimedAt,
			"decided_at":      p.DecidedAt,
			"decided_by":      p.DecidedBy,
			"reject_reason":   p.RejectReason,
		}).Error
}

func (r *rewardDatabase) ExpireStalePurchases(ctx context.Context, now time.Time) (int64, error) {
	res := r.db.WithContext(ctx).Exec(`
		UPDATE shop_purchases SET status = ?, decided_at = ?
		WHERE status = ? AND expires_at <= ?`,
		domain.ShopPurchaseExpired, now, domain.ShopPurchasePending, now)
	return res.RowsAffected, res.Error
}

func (r *rewardDatabase) ListPurchases(ctx context.Context, f domain.ShopPurchaseFilter, p request.Pagination) ([]domain.ShopPurchaseView, error) {
	where := []string{"1=1"}
	args := []interface{}{}
	if f.ShopID != "" {
		where = append(where, "p.shop_id = ?")
		args = append(args, f.ShopID)
	}
	if f.CustomerID != "" {
		where = append(where, "p.customer_id = ?")
		args = append(args, f.CustomerID)
	}
	if f.Status != "" {
		where = append(where, "p.status = ?")
		args = append(args, f.Status)
	}
	limit := p.Limit
	if limit == 0 {
		limit = defaultRewardPageSize
	}
	args = append(args, limit, p.Offset)

	views := []domain.ShopPurchaseView{}
	err := r.db.WithContext(ctx).Raw(`
		SELECT p.*,
		       TRIM(COALESCE(u.first_name, '') || ' ' || COALESCE(u.last_name, '')) AS customer_name,
		       RIGHT(COALESCE(u.phone, ''), 4) AS customer_phone_last4,
		       COALESCE(s.shop_name, '') AS shop_name,
		       (SELECT COUNT(*) FROM shop_purchases c
		         WHERE c.shop_id = p.shop_id AND c.customer_id = p.customer_id AND c.status = 'claimed') AS visit_count,
		       prev.bill_amount_paise AS prev_bill_amount_paise,
		       prev.created_at AS prev_purchase_at
		FROM shop_purchases p
		JOIN users u ON u.id = p.customer_id
		JOIN shop_details s ON s.id = p.shop_id
		LEFT JOIN LATERAL (
		    SELECT q.bill_amount_paise, q.created_at FROM shop_purchases q
		    WHERE q.shop_id = p.shop_id AND q.customer_id = p.customer_id
		      AND q.status = 'claimed' AND q.created_at < p.created_at
		    ORDER BY q.created_at DESC LIMIT 1
		) prev ON TRUE
		WHERE `+strings.Join(where, " AND ")+`
		ORDER BY p.created_at DESC, p.id DESC
		LIMIT ? OFFSET ?`, args...).Scan(&views).Error
	return views, err
}
