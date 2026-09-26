package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tptr(t time.Time) *time.Time { return &t }

func TestCredit_IsIdempotentPerRef(t *testing.T) {
	f := newFakeRewardRepo()
	ctx := context.Background()
	acct, _ := f.GetOrCreateAccount(ctx, domain.RewardOwnerShop, fxShopID)
	in := creditInput{AccountID: acct.ID, Points: 12, Type: domain.RewardEntryPurchaseEarn, RefType: "shop_purchase", RefID: "spur_9"}

	ok, err := credit(ctx, f, in)
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = credit(ctx, f, in)
	require.NoError(t, err)
	assert.False(t, ok)

	got, _ := f.GetAccountByID(ctx, acct.ID)
	assert.Equal(t, int64(12), got.BalancePoints)
	assert.Equal(t, int64(12), got.LifetimeEarned)
	assert.Len(t, f.entries(acct.ID, ""), 1)
	assert.Equal(t, int64(12), f.entries(acct.ID, "")[0].RemainingPoints)
}

func TestCredit_ZeroPointsIsNoop(t *testing.T) {
	f := newFakeRewardRepo()
	acct, _ := f.GetOrCreateAccount(context.Background(), domain.RewardOwnerShop, fxShopID)
	ok, err := credit(context.Background(), f, creditInput{AccountID: acct.ID, Points: 0, Type: domain.RewardEntryPurchaseEarn, RefID: "x"})
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Empty(t, f.ledger)
}

func TestDebit_ConsumesOldestExpiringLotsFirst(t *testing.T) {
	f := newFakeRewardRepo()
	ctx := context.Background()
	acct, _ := f.GetOrCreateAccount(ctx, domain.RewardOwnerShop, fxShopID)
	base := time.Date(2026, 9, 1, 0, 0, 0, 0, time.UTC)
	_, _ = credit(ctx, f, creditInput{AccountID: acct.ID, Points: 30, Type: domain.RewardEntryPurchaseEarn, RefID: "late", ExpiresAt: tptr(base.AddDate(0, 12, 0))})
	_, _ = credit(ctx, f, creditInput{AccountID: acct.ID, Points: 20, Type: domain.RewardEntryPurchaseEarn, RefID: "early", ExpiresAt: tptr(base.AddDate(0, 6, 0))})

	ok, err := debit(ctx, f, creditInput{AccountID: acct.ID, Points: 25, Type: domain.RewardEntryAdminAdjust, RefType: "admin_adjust", RefID: "adj_1"})
	require.NoError(t, err)
	assert.True(t, ok)

	remaining := map[string]int64{}
	for _, e := range f.entries(acct.ID, domain.RewardEntryPurchaseEarn) {
		remaining[e.RefID] = e.RemainingPoints
	}
	assert.Equal(t, int64(0), remaining["early"], "earliest-expiring lot is spent first")
	assert.Equal(t, int64(25), remaining["late"])
	got, _ := f.GetAccountByID(ctx, acct.ID)
	assert.Equal(t, int64(25), got.BalancePoints)
	assert.Equal(t, int64(25), got.LifetimeSpent)
}

func TestDebit_InsufficientLeavesStateUntouched(t *testing.T) {
	f := newFakeRewardRepo()
	ctx := context.Background()
	acct, _ := f.GetOrCreateAccount(ctx, domain.RewardOwnerShop, fxShopID)
	_, _ = credit(ctx, f, creditInput{AccountID: acct.ID, Points: 10, Type: domain.RewardEntryPurchaseEarn, RefID: "a"})

	_, err := debit(ctx, f, creditInput{AccountID: acct.ID, Points: 11, Type: domain.RewardEntryAdminAdjust, RefID: "adj"})
	assert.ErrorIs(t, err, ErrInsufficientPoints)
	got, _ := f.GetAccountByID(ctx, acct.ID)
	assert.Equal(t, int64(10), got.BalancePoints)
	assert.Len(t, f.ledger, 1)
}

func TestDebit_ReplayAfterBalanceDropIsNoop(t *testing.T) {
	f := newFakeRewardRepo()
	ctx := context.Background()
	acct, _ := f.GetOrCreateAccount(ctx, domain.RewardOwnerShop, fxShopID)
	_, _ = credit(ctx, f, creditInput{AccountID: acct.ID, Points: 100, Type: domain.RewardEntryPurchaseEarn, RefID: "c1"})

	ok, err := debit(ctx, f, creditInput{AccountID: acct.ID, Points: 100, Type: domain.RewardEntryAdminAdjust, RefType: "admin_adjust", RefID: "adj_1"})
	require.NoError(t, err)
	assert.True(t, ok)
	got, _ := f.GetAccountByID(ctx, acct.ID)
	assert.Equal(t, int64(0), got.BalancePoints)

	ok, err = debit(ctx, f, creditInput{AccountID: acct.ID, Points: 100, Type: domain.RewardEntryAdminAdjust, RefType: "admin_adjust", RefID: "adj_1"})
	assert.NoError(t, err)
	assert.False(t, ok, "replaying the same ref after balance dropped must be a no-op, not ErrInsufficientPoints")
	got, _ = f.GetAccountByID(ctx, acct.ID)
	assert.Equal(t, int64(0), got.BalancePoints, "balance must not go negative or change on replay")
	assert.Len(t, f.entries(acct.ID, domain.RewardEntryAdminAdjust), 1, "exactly one debit entry")
}

func TestExpireLot_RemovesOnlyRemainingPoints(t *testing.T) {
	f := newFakeRewardRepo()
	ctx := context.Background()
	acct, _ := f.GetOrCreateAccount(ctx, domain.RewardOwnerShop, fxShopID)
	_, _ = credit(ctx, f, creditInput{AccountID: acct.ID, Points: 40, Type: domain.RewardEntryPurchaseEarn, RefID: "lot"})
	_, _ = debit(ctx, f, creditInput{AccountID: acct.ID, Points: 15, Type: domain.RewardEntryAdminAdjust, RefID: "adj"})
	lot := *f.entries(acct.ID, domain.RewardEntryPurchaseEarn)[0]

	ok, err := expireLot(ctx, f, lot)
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = expireLot(ctx, f, lot)
	require.NoError(t, err)
	assert.False(t, ok, "expiring the same lot twice is a no-op")

	got, _ := f.GetAccountByID(ctx, acct.ID)
	assert.Equal(t, int64(0), got.BalancePoints)
	assert.Equal(t, int64(15), got.LifetimeSpent, "expiry is not counted as spend")
	exp := f.entries(acct.ID, domain.RewardEntryExpiry)
	require.Len(t, exp, 1)
	assert.Equal(t, int64(-25), exp[0].DeltaPoints)
}
