package usecase

import (
	"context"
	"testing"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	fxOpsAdmin    = "adm_ops"
	fxSellerAdmin = "adm_seller"
	fxBlankAdmin  = "adm_blank"
)

// withAdminRoles registers a platform admin, a seller and a blank-role account.
func (f *fakeRewardRepo) withAdminRoles() *fakeRewardRepo {
	f.adminRoles[fxOpsAdmin] = domain.AdminRoleSuperAdmin
	f.adminRoles[fxSellerAdmin] = domain.AdminRoleSeller
	f.adminRoles[fxBlankAdmin] = ""
	return f
}

var nonPlatformCallers = map[string]string{
	"seller role":   fxSellerAdmin,
	"blank role":    fxBlankAdmin,
	"unknown admin": "adm_missing",
}

func TestGetProgramConfig_PlatformAdminOnly(t *testing.T) {
	for name, caller := range nonPlatformCallers {
		t.Run(name, func(t *testing.T) {
			f := newFakeRewardRepo().withAdminRoles()
			uc, _ := newTestRewardUseCase(f, fxNow)
			_, err := uc.GetProgramConfig(context.Background(), caller)
			assert.ErrorIs(t, err, ErrRewardAdminOnly)
		})
	}
	f := newFakeRewardRepo().withAdminRoles()
	uc, _ := newTestRewardUseCase(f, fxNow)
	got, err := uc.GetProgramConfig(context.Background(), fxOpsAdmin)
	require.NoError(t, err)
	assert.Equal(t, defaultRewardConfig().GPSRadiusM, got.GPSRadiusM)
}

func TestUpdateProgramConfig_PlatformAdminOnly(t *testing.T) {
	for name, caller := range nonPlatformCallers {
		t.Run(name, func(t *testing.T) {
			f := newFakeRewardRepo().withAdminRoles()
			uc, _ := newTestRewardUseCase(f, fxNow)
			cfg := defaultRewardConfig()
			cfg.GPSRadiusM = 999
			_, err := uc.UpdateProgramConfig(context.Background(), caller, cfg)
			assert.ErrorIs(t, err, ErrRewardAdminOnly)
			assert.Equal(t, defaultRewardConfig().GPSRadiusM, f.cfg.GPSRadiusM, "config must be unchanged")
			assert.Empty(t, f.cfg.UpdatedBy)
		})
	}
	f := newFakeRewardRepo().withAdminRoles()
	uc, _ := newTestRewardUseCase(f, fxNow)
	cfg := defaultRewardConfig()
	cfg.GPSRadiusM = 999
	_, err := uc.UpdateProgramConfig(context.Background(), fxOpsAdmin, cfg)
	require.NoError(t, err)
	assert.Equal(t, 999, f.cfg.GPSRadiusM)
}

func TestListAllPurchases_PlatformAdminOnly(t *testing.T) {
	for name, caller := range nonPlatformCallers {
		t.Run(name, func(t *testing.T) {
			f := newFakeRewardRepo().withOptedInShop(10).withAdminRoles()
			pendingPurchase(f, 45000)
			uc, _ := newTestRewardUseCase(f, fxNow)
			views, err := uc.ListAllPurchases(context.Background(), caller, domain.ShopPurchaseFilter{}, request.Pagination{})
			assert.ErrorIs(t, err, ErrRewardAdminOnly)
			assert.Empty(t, views)
		})
	}
	f := newFakeRewardRepo().withOptedInShop(10).withAdminRoles()
	pendingPurchase(f, 45000)
	uc, _ := newTestRewardUseCase(f, fxNow)
	views, err := uc.ListAllPurchases(context.Background(), fxOpsAdmin, domain.ShopPurchaseFilter{}, request.Pagination{})
	require.NoError(t, err)
	assert.Len(t, views, 1)
}

func TestAdjustAccount_PlatformAdminOnly(t *testing.T) {
	for name, caller := range nonPlatformCallers {
		t.Run(name, func(t *testing.T) {
			f := newFakeRewardRepo().withOptedInShop(10).withAdminRoles()
			uc, _ := newTestRewardUseCase(f, fxNow)
			acct, _ := f.GetOrCreateAccount(context.Background(), domain.RewardOwnerShop, fxShopID)
			_, err := uc.AdjustAccount(context.Background(), caller, acct.ID, 5000, "minting points for myself")
			assert.ErrorIs(t, err, ErrRewardAdminOnly)
			assert.Equal(t, int64(0), f.accounts[acct.ID].BalancePoints, "balance must be unchanged")
			assert.Empty(t, f.entries(acct.ID, ""), "no ledger entry may be written")
		})
	}
	t.Run("check runs before validation", func(t *testing.T) {
		f := newFakeRewardRepo().withAdminRoles()
		uc, _ := newTestRewardUseCase(f, fxNow)
		_, err := uc.AdjustAccount(context.Background(), fxSellerAdmin, "rwa_missing", 0, "")
		assert.ErrorIs(t, err, ErrRewardAdminOnly)
	})
	f := newFakeRewardRepo().withOptedInShop(10).withAdminRoles()
	uc, _ := newTestRewardUseCase(f, fxNow)
	acct, _ := f.GetOrCreateAccount(context.Background(), domain.RewardOwnerShop, fxShopID)
	got, err := uc.AdjustAccount(context.Background(), fxOpsAdmin, acct.ID, 50, "goodwill for app outage")
	require.NoError(t, err)
	assert.Equal(t, int64(50), got.BalancePoints)
}

func TestListSellerPurchases_MasksCustomer(t *testing.T) {
	f := newFakeRewardRepo().withOptedInShop(10).withAdminRoles()
	c := f.customers[fxCustomer]
	c.LastName = "Kumar"
	f.customers[fxCustomer] = c
	pendingPurchase(f, 45000)
	uc, _ := newTestRewardUseCase(f, fxNow)

	views, err := uc.ListSellerPurchases(context.Background(), fxAdminID, "", request.Pagination{})
	require.NoError(t, err)
	require.Len(t, views, 1)
	assert.Equal(t, "Ravi K.", views[0].CustomerName)
	assert.Empty(t, views[0].CustomerPhoneLast4)
	assert.Empty(t, views[0].CustomerID)

	all, err := uc.ListAllPurchases(context.Background(), fxOpsAdmin, domain.ShopPurchaseFilter{}, request.Pagination{})
	require.NoError(t, err)
	require.Len(t, all, 1)
	assert.Equal(t, "Ravi Kumar", all[0].CustomerName, "admin list keeps full detail")
	assert.Equal(t, "3210", all[0].CustomerPhoneLast4)
	assert.Equal(t, fxCustomer, all[0].CustomerID)
}
