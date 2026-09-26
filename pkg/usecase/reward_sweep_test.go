package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRunSweep(t *testing.T) {
	f := newFakeRewardRepo().withOptedInShop(10)
	ctx := context.Background()
	stale := f.addPurchase(domain.ShopPurchase{ShopID: fxShopID, CustomerID: fxCustomer, Status: domain.ShopPurchasePending, ExpiresAt: fxNow.Add(-time.Minute)})
	fresh := f.addPurchase(domain.ShopPurchase{ShopID: fxShopID, CustomerID: "usr_2", Status: domain.ShopPurchasePending, ExpiresAt: fxNow.Add(time.Hour)})

	acct, _ := f.GetOrCreateAccount(ctx, domain.RewardOwnerShop, fxShopID)
	_, _ = credit(ctx, f, creditInput{AccountID: acct.ID, Points: 40, Type: domain.RewardEntryPurchaseEarn, RefID: "old", ExpiresAt: tptr(fxNow.Add(-time.Hour))})
	_, _ = credit(ctx, f, creditInput{AccountID: acct.ID, Points: 25, Type: domain.RewardEntryPurchaseEarn, RefID: "soon", ExpiresAt: tptr(fxNow.Add(3 * 24 * time.Hour))})
	_, _ = credit(ctx, f, creditInput{AccountID: acct.ID, Points: 10, Type: domain.RewardEntryPurchaseEarn, RefID: "later", ExpiresAt: tptr(fxNow.AddDate(0, 6, 0))})

	uc, pusher := newTestRewardUseCase(f, fxNow)
	res, err := uc.RunSweep(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, res.ExpiredPurchases)
	assert.Equal(t, 1, res.ExpiredLots)
	assert.Equal(t, 1, res.Reminders)
	assert.Equal(t, 0, res.Errors)

	assert.Equal(t, domain.ShopPurchaseExpired, f.purchases[stale.ID].Status)
	assert.Equal(t, domain.ShopPurchasePending, f.purchases[fresh.ID].Status)
	got, _ := f.GetAccountByID(ctx, acct.ID)
	assert.Equal(t, int64(35), got.BalancePoints)

	require.Len(t, pusher.sent, 1)
	assert.Equal(t, fxAdminID, pusher.sent[0].OwnerID, "shop reminders go to the shop's seller")
	assert.Equal(t, "reward_points_expiring", pusher.sent[0].EventType)
	assert.Equal(t, "25", pusher.sent[0].Data["points"])

	res, err = uc.RunSweep(ctx)
	require.NoError(t, err)
	assert.Equal(t, RewardSweepResult{}, res, "a second sweep finds nothing to do")
	assert.Len(t, pusher.sent, 1, "reminders are sent once per lot")
}
