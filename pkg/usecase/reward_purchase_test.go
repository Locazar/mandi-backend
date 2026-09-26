package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var fxNow = time.Date(2026, 9, 26, 6, 0, 0, 0, time.UTC) // 11:30 IST

func submitIn() SubmitPurchaseInput {
	return SubmitPurchaseInput{CustomerID: fxCustomer, ShopID: fxShopID, ClientRequestID: "crid-1",
		BillAmountPaise: 50000, Lat: fxShopLat, Lng: fxShopLng}
}

func TestSubmitPurchase_HappyPath(t *testing.T) {
	f := newFakeRewardRepo().withOptedInShop(10)
	uc, pusher := newTestRewardUseCase(f, fxNow)

	p, err := uc.SubmitPurchase(context.Background(), submitIn())
	require.NoError(t, err)
	assert.Equal(t, domain.ShopPurchasePending, p.Status)
	assert.Equal(t, 10, p.DiscountPercent)
	assert.Equal(t, int64(5000), p.DiscountPaise)
	assert.Equal(t, int64(45000), p.NetPaidPaise)
	assert.Equal(t, fxNow.Add(48*time.Hour), p.ExpiresAt)
	assert.Equal(t, 0, p.DistanceM)
	assert.Equal(t, 1, f.pairLocks, "submit must take the customer/shop lock")

	require.Len(t, pusher.sent, 1)
	assert.Equal(t, fxAdminID, pusher.sent[0].OwnerID)
	assert.Equal(t, "seller", pusher.sent[0].OwnerType)
	assert.Equal(t, "reward_purchase", pusher.sent[0].EventType)
	assert.Equal(t, p.ID, pusher.sent[0].Data["purchase_id"])
	assert.Equal(t, "Ravi", pusher.sent[0].Data["customer_name"])
}

func TestSubmitPurchase_DuplicateClientRequestReturnsOriginal(t *testing.T) {
	f := newFakeRewardRepo().withOptedInShop(10)
	uc, pusher := newTestRewardUseCase(f, fxNow)
	first, err := uc.SubmitPurchase(context.Background(), submitIn())
	require.NoError(t, err)
	second, err := uc.SubmitPurchase(context.Background(), submitIn())
	require.NoError(t, err)
	assert.Equal(t, first.ID, second.ID)
	assert.Len(t, f.purchases, 1)
	assert.Len(t, pusher.sent, 1, "a retry must not push the seller again")
}

func TestSubmitPurchase_Rejections(t *testing.T) {
	tests := map[string]struct {
		arrange func(f *fakeRewardRepo, in *SubmitPurchaseInput)
		want    error
	}{
		"program paused":      {func(f *fakeRewardRepo, _ *SubmitPurchaseInput) { f.cfg.ProgramEnabled = false }, ErrRewardProgramDisabled},
		"seller token":        {func(_ *fakeRewardRepo, in *SubmitPurchaseInput) { in.CustomerID = fxAdminID }, ErrRewardCustomerOnly},
		"unknown shop":        {func(_ *fakeRewardRepo, in *SubmitPurchaseInput) { in.ShopID = "shp_nope" }, ErrRewardShopNotFound},
		"shop never opted in": {func(f *fakeRewardRepo, _ *SubmitPurchaseInput) { delete(f.settings, fxShopID) }, ErrShopNotOptedIn},
		"shop opted out": {func(f *fakeRewardRepo, _ *SubmitPurchaseInput) {
			s := f.settings[fxShopID]
			s.DiscountEnabled = false
			f.settings[fxShopID] = s
		}, ErrShopNotOptedIn},
		"shop not active": {func(f *fakeRewardRepo, _ *SubmitPurchaseInput) {
			s := f.shops[fxShopID]
			s.ShopStatus = domain.ShopStatusSuspended
			f.shops[fxShopID] = s
		}, ErrShopNotOptedIn},
		"own shop": {func(f *fakeRewardRepo, _ *SubmitPurchaseInput) {
			c := f.customers[fxCustomer]
			c.Phone = "09000000001"
			f.customers[fxCustomer] = c
		}, ErrSelfPurchase},
		"shop location missing": {func(f *fakeRewardRepo, _ *SubmitPurchaseInput) {
			s := f.shops[fxShopID]
			s.Latitude, s.Longitude = 0, 0
			f.shops[fxShopID] = s
		}, ErrShopLocationMissing},
		"too far":        {func(_ *fakeRewardRepo, in *SubmitPurchaseInput) { in.Lat = fxShopLat + 0.003 }, ErrTooFarFromShop},
		"below min bill": {func(_ *fakeRewardRepo, in *SubmitPurchaseInput) { in.BillAmountPaise = 9999 }, ErrBelowMinBill},
		"zero bill":      {func(_ *fakeRewardRepo, in *SubmitPurchaseInput) { in.BillAmountPaise = 0 }, ErrInvalidBillAmount},
		"absurd bill":    {func(_ *fakeRewardRepo, in *SubmitPurchaseInput) { in.BillAmountPaise = maxBillPaise + 1 }, ErrInvalidBillAmount},
		"claimed 3 days ago": {func(f *fakeRewardRepo, _ *SubmitPurchaseInput) {
			f.addPurchase(domain.ShopPurchase{ShopID: fxShopID, CustomerID: fxCustomer, Status: domain.ShopPurchaseClaimed, CreatedAt: fxNow.Add(-72 * time.Hour)})
		}, ErrRepeatPurchaseWindow},
		"still pending": {func(f *fakeRewardRepo, _ *SubmitPurchaseInput) {
			f.addPurchase(domain.ShopPurchase{ShopID: fxShopID, CustomerID: fxCustomer, Status: domain.ShopPurchasePending, CreatedAt: fxNow.Add(-time.Hour), ExpiresAt: fxNow.Add(47 * time.Hour)})
		}, ErrRepeatPurchaseWindow},
		"shop daily cap": {func(f *fakeRewardRepo, _ *SubmitPurchaseInput) {
			f.cfg.MaxClaimsPerShopPerDay = 1
			f.addPurchase(domain.ShopPurchase{ShopID: fxShopID, CustomerID: "usr_other", Status: domain.ShopPurchaseClaimed, CreatedAt: fxNow.Add(-time.Hour)})
		}, ErrShopDailyCapReached},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			f := newFakeRewardRepo().withOptedInShop(10)
			in := submitIn()
			tc.arrange(f, &in)
			uc, pusher := newTestRewardUseCase(f, fxNow)
			_, err := uc.SubmitPurchase(context.Background(), in)
			assert.ErrorIs(t, err, tc.want)
			assert.Empty(t, pusher.sent)
		})
	}
}

func TestSubmitPurchase_NonBlockingHistoryAllowsPurchase(t *testing.T) {
	tests := map[string]domain.ShopPurchase{
		"rejected yesterday":           {Status: domain.ShopPurchaseRejected, CreatedAt: fxNow.Add(-24 * time.Hour)},
		"expired status":               {Status: domain.ShopPurchaseExpired, CreatedAt: fxNow.Add(-50 * time.Hour)},
		"pending past expiry unswept":  {Status: domain.ShopPurchasePending, CreatedAt: fxNow.Add(-49 * time.Hour), ExpiresAt: fxNow.Add(-time.Hour)},
		"claimed just over 7 days ago": {Status: domain.ShopPurchaseClaimed, CreatedAt: fxNow.Add(-168*time.Hour - time.Minute)},
	}
	for name, prior := range tests {
		t.Run(name, func(t *testing.T) {
			f := newFakeRewardRepo().withOptedInShop(10)
			prior.ShopID, prior.CustomerID = fxShopID, fxCustomer
			f.addPurchase(prior)
			uc, _ := newTestRewardUseCase(f, fxNow)
			_, err := uc.SubmitPurchase(context.Background(), submitIn())
			assert.NoError(t, err)
		})
	}
}

func TestGetEligibility(t *testing.T) {
	t.Run("eligible", func(t *testing.T) {
		f := newFakeRewardRepo().withOptedInShop(15)
		uc, _ := newTestRewardUseCase(f, fxNow)
		e, err := uc.GetEligibility(context.Background(), fxCustomer, fxShopID, fxShopLat, fxShopLng)
		require.NoError(t, err)
		assert.True(t, e.Eligible)
		assert.Equal(t, 15, e.DiscountPercent)
		assert.Equal(t, "Asha Stores", e.ShopName)
		assert.Equal(t, int64(10000), e.MinBillPaise)
		assert.Equal(t, 200, e.GPSRadiusM)
	})
	t.Run("repeat window gives next available time", func(t *testing.T) {
		f := newFakeRewardRepo().withOptedInShop(10)
		claimedAt := fxNow.Add(-72 * time.Hour)
		f.addPurchase(domain.ShopPurchase{ShopID: fxShopID, CustomerID: fxCustomer, Status: domain.ShopPurchaseClaimed, CreatedAt: claimedAt})
		uc, _ := newTestRewardUseCase(f, fxNow)
		e, err := uc.GetEligibility(context.Background(), fxCustomer, fxShopID, fxShopLat, fxShopLng)
		require.NoError(t, err)
		assert.False(t, e.Eligible)
		assert.Equal(t, "repeat_window", e.Reason)
		require.NotNil(t, e.NextAvailableAt)
		assert.Equal(t, claimedAt.Add(168*time.Hour), *e.NextAvailableAt)
	})
	t.Run("rule failures become reason codes", func(t *testing.T) {
		f := newFakeRewardRepo().withOptedInShop(10)
		s := f.shops[fxShopID]
		s.Latitude, s.Longitude = 0, 0
		f.shops[fxShopID] = s
		uc, _ := newTestRewardUseCase(f, fxNow)
		e, err := uc.GetEligibility(context.Background(), fxCustomer, fxShopID, fxShopLat, fxShopLng)
		require.NoError(t, err)
		assert.False(t, e.Eligible)
		assert.Equal(t, "shop_location_missing", e.Reason)
	})
	t.Run("seller token is an error, not a reason", func(t *testing.T) {
		f := newFakeRewardRepo().withOptedInShop(10)
		uc, _ := newTestRewardUseCase(f, fxNow)
		_, err := uc.GetEligibility(context.Background(), fxAdminID, fxShopID, fxShopLat, fxShopLng)
		assert.ErrorIs(t, err, ErrRewardCustomerOnly)
	})
}
