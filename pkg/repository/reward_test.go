package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// testDB lives in invoice_test.go (same package) — reuse it, do not redeclare.

type rewardSeed struct{ adminID, userID, shopID string }

// seedRewardFixtures inserts a minimal admin, user and shop. If the admins or
// shop_details tables gain NOT NULL columns without defaults, add them here.
func seedRewardFixtures(t *testing.T, db *gorm.DB) rewardSeed {
	t.Helper()
	s := rewardSeed{
		adminID: domain.NewID(domain.PrefixAdmin),
		userID:  domain.NewID(domain.PrefixUser),
		shopID:  domain.NewID(domain.PrefixShop),
	}
	phone := "9" + s.userID[len(s.userID)-9:]
	require.NoError(t, db.Exec(`INSERT INTO admins (id, mobile) VALUES (?, '9000000001')`, s.adminID).Error)
	require.NoError(t, db.Exec(`INSERT INTO users (id, phone, password, first_name) VALUES (?, ?, 'x', 'Asha')`, s.userID, phone).Error)
	require.NoError(t, db.Exec(`INSERT INTO shop_details (id, admin_id, shop_name, shop_status, latitude, longitude)
		VALUES (?, ?, 'Test Shop', 'active', 12.9756, 77.6050)`, s.shopID, s.adminID).Error)
	t.Cleanup(func() {
		db.Exec(`DELETE FROM shop_purchases WHERE shop_id = ?`, s.shopID)
		db.Exec(`DELETE FROM reward_ledger_entries WHERE account_id IN (SELECT id FROM reward_accounts WHERE owner_id = ?)`, s.shopID)
		db.Exec(`DELETE FROM reward_accounts WHERE owner_id = ?`, s.shopID)
		db.Exec(`DELETE FROM shop_reward_settings WHERE shop_id = ?`, s.shopID)
		db.Exec(`DELETE FROM shop_details WHERE id = ?`, s.shopID)
		db.Exec(`DELETE FROM users WHERE id = ?`, s.userID)
		db.Exec(`DELETE FROM admins WHERE id = ?`, s.adminID)
	})
	return s
}

func TestRewardRepo_ConfigSeededAndUpdatable(t *testing.T) {
	db := testDB(t)
	repo := NewRewardRepository(db)
	ctx := context.Background()

	cfg, err := repo.GetConfig(ctx)
	require.NoError(t, err)
	assert.Equal(t, []int64{5, 10, 15, 20}, []int64(cfg.AllowedDiscountPercents))
	assert.Equal(t, 200, cfg.GPSRadiusM)

	orig := cfg
	t.Cleanup(func() { _, _ = repo.UpdateConfig(ctx, orig) })
	cfg.GPSRadiusM = 250
	cfg.ProgramEnabled = false // zero value must still be written
	updated, err := repo.UpdateConfig(ctx, cfg)
	require.NoError(t, err)
	assert.Equal(t, 250, updated.GPSRadiusM)
	assert.False(t, updated.ProgramEnabled)
}

func TestRewardRepo_ShopLookupsAndSettings(t *testing.T) {
	db := testDB(t)
	repo := NewRewardRepository(db)
	ctx := context.Background()
	s := seedRewardFixtures(t, db)

	shop, err := repo.GetShop(ctx, s.shopID)
	require.NoError(t, err)
	assert.Equal(t, s.adminID, shop.AdminID)
	assert.Equal(t, "9000000001", shop.OwnerMobile)
	assert.InDelta(t, 12.9756, shop.Latitude, 1e-6)

	byAdmin, err := repo.GetShopByAdminID(ctx, s.adminID)
	require.NoError(t, err)
	assert.Equal(t, s.shopID, byAdmin.ID)

	_, err = repo.GetShop(ctx, "shp_missing")
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	_, err = repo.GetCustomer(ctx, s.adminID) // an admin id is not a customer
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)

	_, err = repo.GetShopSettings(ctx, s.shopID)
	assert.ErrorIs(t, err, gorm.ErrRecordNotFound)
	saved, err := repo.UpsertShopSettings(ctx, domain.ShopRewardSettings{ShopID: s.shopID, AdminID: s.adminID, DiscountEnabled: true, DiscountPercent: 10})
	require.NoError(t, err)
	assert.Equal(t, 10, saved.DiscountPercent)
	saved, err = repo.UpsertShopSettings(ctx, domain.ShopRewardSettings{ShopID: s.shopID, AdminID: s.adminID, DiscountEnabled: false, DiscountPercent: 0})
	require.NoError(t, err)
	assert.False(t, saved.DiscountEnabled)
}

func TestRewardRepo_LedgerInsertIsIdempotent(t *testing.T) {
	db := testDB(t)
	repo := NewRewardRepository(db)
	ctx := context.Background()
	s := seedRewardFixtures(t, db)

	acct, err := repo.GetOrCreateAccount(ctx, domain.RewardOwnerShop, s.shopID)
	require.NoError(t, err)
	again, err := repo.GetOrCreateAccount(ctx, domain.RewardOwnerShop, s.shopID)
	require.NoError(t, err)
	assert.Equal(t, acct.ID, again.ID)

	e := domain.RewardLedgerEntry{AccountID: acct.ID, DeltaPoints: 12, EntryType: domain.RewardEntryPurchaseEarn, RefType: "shop_purchase", RefID: "spur_x", RemainingPoints: 12}
	ok, err := repo.InsertLedgerEntry(ctx, e)
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = repo.InsertLedgerEntry(ctx, e)
	require.NoError(t, err)
	assert.False(t, ok, "duplicate (account, type, ref) must be a no-op")

	require.NoError(t, repo.AdjustAccountBalance(ctx, acct.ID, 12, 12, 0))
	got, err := repo.GetAccountByID(ctx, acct.ID)
	require.NoError(t, err)
	assert.Equal(t, int64(12), got.BalancePoints)

	err = repo.AdjustAccountBalance(ctx, acct.ID, -13, 0, 13)
	assert.Error(t, err, "balance CHECK >= 0 must reject overdraft")
}

func TestRewardRepo_PairLockSerialisesSubmits(t *testing.T) {
	db := testDB(t)
	repo := NewRewardRepository(db)
	ctx := context.Background()
	s := seedRewardFixtures(t, db)
	now := time.Now()

	const workers = 10
	var wg sync.WaitGroup
	created := make(chan string, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = repo.InTx(ctx, func(r interfaces.RewardRepository) error {
				if err := r.LockCustomerShopPair(ctx, s.userID, s.shopID); err != nil {
					return err
				}
				blocking, err := r.FindBlockingPurchase(ctx, s.userID, s.shopID, now.Add(-168*time.Hour), now)
				if err != nil || blocking != nil {
					return err
				}
				p, err := r.CreatePurchase(ctx, domain.ShopPurchase{
					ShopID: s.shopID, CustomerID: s.userID, ClientRequestID: domain.NewID(domain.PrefixShopPurchase),
					BillAmountPaise: 50000, DiscountPercent: 10, DiscountPaise: 5000, NetPaidPaise: 45000,
					Status: domain.ShopPurchasePending, ExpiresAt: now.Add(48 * time.Hour),
				})
				if err == nil {
					created <- p.ID
				}
				return err
			})
		}(i)
	}
	wg.Wait()
	close(created)
	assert.Len(t, created, 1, "exactly one pending purchase per customer/shop window")
}

func TestRewardRepo_ExpireStalePurchasesAndListing(t *testing.T) {
	db := testDB(t)
	repo := NewRewardRepository(db)
	ctx := context.Background()
	s := seedRewardFixtures(t, db)
	now := time.Now()

	p, err := repo.CreatePurchase(ctx, domain.ShopPurchase{
		ShopID: s.shopID, CustomerID: s.userID, ClientRequestID: "crid-1",
		BillAmountPaise: 50000, DiscountPercent: 10, DiscountPaise: 5000, NetPaidPaise: 45000,
		Status: domain.ShopPurchasePending, ExpiresAt: now.Add(-time.Minute),
	})
	require.NoError(t, err)

	dup, err := repo.FindPurchaseByClientRequest(ctx, s.userID, "crid-1")
	require.NoError(t, err)
	require.NotNil(t, dup)
	assert.Equal(t, p.ID, dup.ID)

	blocking, err := repo.FindBlockingPurchase(ctx, s.userID, s.shopID, now.Add(-168*time.Hour), now)
	require.NoError(t, err)
	assert.Nil(t, blocking, "an overdue pending purchase must not block")

	n, err := repo.ExpireStalePurchases(ctx, now)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, n, int64(1))

	views, err := repo.ListPurchases(ctx, domain.ShopPurchaseFilter{ShopID: s.shopID}, request.Pagination{Limit: 10})
	require.NoError(t, err)
	require.Len(t, views, 1)
	assert.Equal(t, domain.ShopPurchaseExpired, views[0].Status)
	assert.Equal(t, "Asha", views[0].CustomerName)
	assert.Equal(t, "Test Shop", views[0].ShopName)
}
