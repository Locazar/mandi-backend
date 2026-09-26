package usecase

import (
	"context"
	"testing"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateProgramConfig_Validation(t *testing.T) {
	tests := map[string]func(c *domain.RewardProgramConfig){
		"no discount options":      func(c *domain.RewardProgramConfig) { c.AllowedDiscountPercents = nil },
		"discount option above 90": func(c *domain.RewardProgramConfig) { c.AllowedDiscountPercents = []int64{5, 95} },
		"duplicate discount":       func(c *domain.RewardProgramConfig) { c.AllowedDiscountPercents = []int64{5, 5} },
		"negative base":            func(c *domain.RewardProgramConfig) { c.SellerBasePoints = -1 },
		"zero seller max":          func(c *domain.RewardProgramConfig) { c.SellerMaxPoints = 0 },
		"gps radius too small":     func(c *domain.RewardProgramConfig) { c.GPSRadiusM = 5 },
		"gps radius too large":     func(c *domain.RewardProgramConfig) { c.GPSRadiusM = 6000 },
		"zero repeat window":       func(c *domain.RewardProgramConfig) { c.RepeatWindowHours = 0 },
		"zero claim expiry":        func(c *domain.RewardProgramConfig) { c.ClaimExpiryHours = 0 },
		"zero point value":         func(c *domain.RewardProgramConfig) { c.PointValuePaise = 0 },
		"expiry months zero":       func(c *domain.RewardProgramConfig) { c.PointsExpiryMonths = 0 },
		"negative min bill":        func(c *domain.RewardProgramConfig) { c.MinBillPaise = -1 },
		"redeem pct over 100%":     func(c *domain.RewardProgramConfig) { c.CustomerMaxRedeemPctBP = 10001 },
	}
	for name, mutate := range tests {
		t.Run(name, func(t *testing.T) {
			f := newFakeRewardRepo().withAdminRoles()
			uc, _ := newTestRewardUseCase(f, fxNow)
			cfg := defaultRewardConfig()
			mutate(&cfg)
			_, err := uc.UpdateProgramConfig(context.Background(), "adm_ops", cfg)
			assert.ErrorIs(t, err, ErrInvalidRewardConfig)
		})
	}
}

func TestUpdateProgramConfig_SavesAndStampsEditor(t *testing.T) {
	f := newFakeRewardRepo().withAdminRoles()
	uc, _ := newTestRewardUseCase(f, fxNow)
	cfg := defaultRewardConfig()
	cfg.GPSRadiusM = 300
	cfg.AllowedDiscountPercents = []int64{20, 5, 10}
	got, err := uc.UpdateProgramConfig(context.Background(), "adm_ops", cfg)
	require.NoError(t, err)
	assert.Equal(t, 300, got.GPSRadiusM)
	assert.Equal(t, "adm_ops", got.UpdatedBy)
	assert.Equal(t, []int64{5, 10, 20}, []int64(got.AllowedDiscountPercents), "options are stored sorted")
}

func TestAdjustAccount(t *testing.T) {
	f := newFakeRewardRepo().withOptedInShop(10).withAdminRoles()
	uc, _ := newTestRewardUseCase(f, fxNow)
	acct, _ := f.GetOrCreateAccount(context.Background(), domain.RewardOwnerShop, fxShopID)

	got, err := uc.AdjustAccount(context.Background(), "adm_ops", acct.ID, 100, "goodwill for app outage", "")
	require.NoError(t, err)
	assert.Equal(t, int64(100), got.BalancePoints)
	lot := f.entries(acct.ID, domain.RewardEntryAdminAdjust)[0]
	require.NotNil(t, lot.ExpiresAt, "admin credits expire like any other points")
	assert.Equal(t, "adm_ops", lot.CreatedBy)

	got, err = uc.AdjustAccount(context.Background(), "adm_ops", acct.ID, -40, "reversing duplicate credit", "")
	require.NoError(t, err)
	assert.Equal(t, int64(60), got.BalancePoints)

	_, err = uc.AdjustAccount(context.Background(), "adm_ops", acct.ID, -61, "too much", "")
	assert.ErrorIs(t, err, ErrInsufficientPoints)
	_, err = uc.AdjustAccount(context.Background(), "adm_ops", acct.ID, 10, "hi", "")
	assert.ErrorIs(t, err, ErrInvalidAdjustment)
	_, err = uc.AdjustAccount(context.Background(), "adm_ops", acct.ID, 0, "zero delta reason", "")
	assert.ErrorIs(t, err, ErrInvalidAdjustment)
	_, err = uc.AdjustAccount(context.Background(), "adm_ops", "rwa_missing", 10, "missing account", "")
	assert.ErrorIs(t, err, ErrRewardAccountNotFound)
}
