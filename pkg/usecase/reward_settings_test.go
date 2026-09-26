package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateSellerSettings(t *testing.T) {
	f := newFakeRewardRepo().withOptedInShop(10)
	delete(f.settings, fxShopID)
	uc, _ := newTestRewardUseCase(f, fxNow)
	ctx := context.Background()

	_, err := uc.UpdateSellerSettings(ctx, fxAdminID, true, 12)
	assert.ErrorIs(t, err, ErrInvalidDiscountPercent)

	got, err := uc.UpdateSellerSettings(ctx, fxAdminID, true, 15)
	require.NoError(t, err)
	assert.True(t, got.DiscountEnabled)
	assert.Equal(t, 15, got.DiscountPercent)
	require.NotNil(t, got.OptedInAt)
	assert.Equal(t, fxNow, *got.OptedInAt)
	assert.NotNil(t, f.accountFor(domain.RewardOwnerShop, fxShopID), "opting in opens the points account")

	later, _ := newTestRewardUseCase(f, fxNow.Add(24*time.Hour))
	got, err = later.UpdateSellerSettings(ctx, fxAdminID, true, 20)
	require.NoError(t, err)
	assert.Equal(t, fxNow, *got.OptedInAt, "first opt-in time is preserved")

	got, err = later.UpdateSellerSettings(ctx, fxAdminID, false, 20)
	require.NoError(t, err)
	assert.False(t, got.DiscountEnabled)
	assert.Equal(t, 0, got.DiscountPercent)

	_, err = uc.UpdateSellerSettings(ctx, "adm_noshop", true, 10)
	assert.ErrorIs(t, err, ErrShopNotFound)
}

func TestGetSellerSettings_DefaultsWhenNeverSaved(t *testing.T) {
	f := newFakeRewardRepo().withOptedInShop(10)
	delete(f.settings, fxShopID)
	uc, _ := newTestRewardUseCase(f, fxNow)
	got, err := uc.GetSellerSettings(context.Background(), fxAdminID)
	require.NoError(t, err)
	assert.False(t, got.DiscountEnabled)
	assert.True(t, got.ProgramEnabled)
	assert.Equal(t, []int64{5, 10, 15, 20}, got.AllowedDiscountPercents)
}

func TestGetSellerAccount_ProgressAndExpiringSoon(t *testing.T) {
	f := newFakeRewardRepo().withOptedInShop(10)
	uc, _ := newTestRewardUseCase(f, fxNow)
	ctx := context.Background()
	acct, _ := f.GetOrCreateAccount(ctx, domain.RewardOwnerShop, fxShopID)
	_, _ = credit(ctx, f, creditInput{AccountID: acct.ID, Points: 300, Type: domain.RewardEntryPurchaseEarn, RefID: "a", ExpiresAt: tptr(fxNow.AddDate(0, 0, 10))})
	_, _ = credit(ctx, f, creditInput{AccountID: acct.ID, Points: 100, Type: domain.RewardEntryPurchaseEarn, RefID: "b", ExpiresAt: tptr(fxNow.AddDate(0, 6, 0))})

	s, err := uc.GetSellerAccount(ctx, fxAdminID)
	require.NoError(t, err)
	assert.Equal(t, int64(400), s.BalancePoints)
	assert.Equal(t, int64(1000), s.RedeemThreshold)
	assert.Equal(t, int64(600), s.PointsToThreshold)
	assert.False(t, s.CanRedeem)
	assert.Equal(t, int64(300), s.ExpiringSoonPoints, "only points expiring within 30 days")
	require.NotNil(t, s.ExpiringSoonAt)
}

func TestGetSellerAccount_NoAccountYetIsZero(t *testing.T) {
	f := newFakeRewardRepo().withOptedInShop(10)
	uc, _ := newTestRewardUseCase(f, fxNow)
	s, err := uc.GetSellerAccount(context.Background(), fxAdminID)
	require.NoError(t, err)
	assert.Equal(t, int64(0), s.BalancePoints)
	assert.Equal(t, int64(1000), s.PointsToThreshold)
}
