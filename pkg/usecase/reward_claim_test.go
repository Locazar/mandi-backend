package usecase

import (
	"context"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func pendingPurchase(f *fakeRewardRepo, net int64) *domain.ShopPurchase {
	return f.addPurchase(domain.ShopPurchase{ShopID: fxShopID, CustomerID: fxCustomer, BillAmountPaise: net,
		NetPaidPaise: net, Status: domain.ShopPurchasePending, CreatedAt: fxNow.Add(-time.Hour), ExpiresAt: fxNow.Add(47 * time.Hour)})
}

func TestClaimPurchase_CreditsSellerOnce(t *testing.T) {
	f := newFakeRewardRepo().withOptedInShop(10)
	p := pendingPurchase(f, 45000) // ₹450 → 10 + 4 = 14
	uc, pusher := newTestRewardUseCase(f, fxNow)

	got, err := uc.ClaimPurchase(context.Background(), fxAdminID, p.ID)
	require.NoError(t, err)
	assert.Equal(t, domain.ShopPurchaseClaimed, got.Status)
	assert.Equal(t, int64(14), got.SellerPoints)
	assert.Equal(t, int64(0), got.CustomerPoints, "customer points arrive in P4")
	require.NotNil(t, got.ClaimedAt)

	acct := f.accountFor(domain.RewardOwnerShop, fxShopID)
	require.NotNil(t, acct)
	assert.Equal(t, int64(14), acct.BalancePoints)
	earn := f.entries(acct.ID, domain.RewardEntryPurchaseEarn)
	require.Len(t, earn, 1)
	assert.Equal(t, p.ID, earn[0].RefID)
	require.NotNil(t, earn[0].ExpiresAt)
	assert.Equal(t, fxNow.AddDate(0, 12, 0), *earn[0].ExpiresAt)

	require.Len(t, pusher.sent, 1)
	assert.Equal(t, fxCustomer, pusher.sent[0].OwnerID)
	assert.Equal(t, "user", pusher.sent[0].OwnerType)

	_, err = uc.ClaimPurchase(context.Background(), fxAdminID, p.ID)
	assert.ErrorIs(t, err, ErrPurchaseNotPending)
	assert.Equal(t, int64(14), f.accountFor(domain.RewardOwnerShop, fxShopID).BalancePoints)
}

func TestClaimPurchase_Rejections(t *testing.T) {
	tests := map[string]struct {
		arrange func(f *fakeRewardRepo, p *domain.ShopPurchase) (adminID, purchaseID string)
		want    error
		// accountMayExist is true for rows where the shop's reward account is
		// legitimately created (and locked) before the failing check runs —
		// e.g. the daily cap is now counted under the account lock, so
		// GetOrCreateAccount has already run by the time the cap rejects the
		// claim. Those rows assert "no points were credited" instead of "no
		// account was created".
		accountMayExist bool
	}{
		"another seller's purchase": {func(f *fakeRewardRepo, p *domain.ShopPurchase) (string, string) {
			f.shops["shp_2"] = domain.RewardShop{ID: "shp_2", AdminID: "adm_2", ShopStatus: domain.ShopStatusActive}
			return "adm_2", p.ID
		}, ErrPurchaseNotFound, false},
		"unknown purchase":    {func(_ *fakeRewardRepo, _ *domain.ShopPurchase) (string, string) { return fxAdminID, "spur_nope" }, ErrPurchaseNotFound, false},
		"seller without shop": {func(_ *fakeRewardRepo, p *domain.ShopPurchase) (string, string) { return "adm_noshop", p.ID }, ErrShopNotFound, false},
		"past expiry before sweep": {func(_ *fakeRewardRepo, p *domain.ShopPurchase) (string, string) {
			p.ExpiresAt = fxNow.Add(-time.Minute)
			return fxAdminID, p.ID
		}, ErrPurchaseNotPending, false},
		"already rejected": {func(_ *fakeRewardRepo, p *domain.ShopPurchase) (string, string) {
			p.Status = domain.ShopPurchaseRejected
			return fxAdminID, p.ID
		}, ErrPurchaseNotPending, false},
		"program paused": {func(f *fakeRewardRepo, p *domain.ShopPurchase) (string, string) {
			f.cfg.ProgramEnabled = false
			return fxAdminID, p.ID
		}, ErrRewardProgramDisabled, false},
		"daily claim cap (IST day)": {func(f *fakeRewardRepo, p *domain.ShopPurchase) (string, string) {
			f.cfg.MaxClaimsPerShopPerDay = 1
			// 00:30 IST today = 19:00 UTC yesterday — counts toward today.
			claimed := time.Date(2026, 9, 25, 19, 0, 0, 0, time.UTC)
			f.addPurchase(domain.ShopPurchase{ShopID: fxShopID, CustomerID: "usr_2", Status: domain.ShopPurchaseClaimed, ClaimedAt: &claimed})
			return fxAdminID, p.ID
		}, ErrShopDailyCapReached, true},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			f := newFakeRewardRepo().withOptedInShop(10)
			p := pendingPurchase(f, 45000)
			adminID, id := tc.arrange(f, f.purchases[p.ID])
			uc, _ := newTestRewardUseCase(f, fxNow)
			_, err := uc.ClaimPurchase(context.Background(), adminID, id)
			assert.ErrorIs(t, err, tc.want)
			acct := f.accountFor(domain.RewardOwnerShop, fxShopID)
			if tc.accountMayExist {
				if acct != nil {
					assert.Empty(t, f.entries(acct.ID, domain.RewardEntryPurchaseEarn), "cap rejection must not credit any points")
				}
				assert.Equal(t, domain.ShopPurchasePending, f.purchases[p.ID].Status, "cap rejection must leave the purchase decidable")
			} else {
				assert.Nil(t, acct)
			}
		})
	}
}

// TestClaimPurchase_LocksAccountBeforeCountingDailyCap guards against the
// race where two devices claim different pending purchases of the same shop
// concurrently: with cap=N and N-1 already claimed today, both could read
// the same pre-lock count and both pass. Locking the shop's account row
// before counting serialises one shop's claims so the count is accurate.
func TestClaimPurchase_LocksAccountBeforeCountingDailyCap(t *testing.T) {
	f := newFakeRewardRepo().withOptedInShop(10)
	p := pendingPurchase(f, 45000)
	uc, _ := newTestRewardUseCase(f, fxNow)

	_, err := uc.ClaimPurchase(context.Background(), fxAdminID, p.ID)
	require.NoError(t, err)

	lockIdx, countIdx := -1, -1
	for i, c := range f.calls {
		if c == "LockAccount" && lockIdx == -1 {
			lockIdx = i
		}
		if c == "CountClaimedSince" && countIdx == -1 {
			countIdx = i
		}
	}
	require.NotEqual(t, -1, lockIdx, "LockAccount was never called")
	require.NotEqual(t, -1, countIdx, "CountClaimedSince was never called")
	assert.Less(t, lockIdx, countIdx, "the shop's reward account must be locked before the daily claim cap is counted")
}

func TestClaimPurchase_CapFromYesterdayIST_DoesNotCount(t *testing.T) {
	f := newFakeRewardRepo().withOptedInShop(10)
	f.cfg.MaxClaimsPerShopPerDay = 1
	// 23:30 IST yesterday = 18:00 UTC yesterday.
	claimed := time.Date(2026, 9, 25, 18, 0, 0, 0, time.UTC)
	f.addPurchase(domain.ShopPurchase{ShopID: fxShopID, CustomerID: "usr_2", Status: domain.ShopPurchaseClaimed, ClaimedAt: &claimed})
	p := pendingPurchase(f, 45000)
	uc, _ := newTestRewardUseCase(f, fxNow)
	_, err := uc.ClaimPurchase(context.Background(), fxAdminID, p.ID)
	assert.NoError(t, err)
}

func TestRejectPurchase(t *testing.T) {
	f := newFakeRewardRepo().withOptedInShop(10)
	p := pendingPurchase(f, 45000)
	uc, pusher := newTestRewardUseCase(f, fxNow)

	got, err := uc.RejectPurchase(context.Background(), fxAdminID, p.ID, "  customer did not buy  ")
	require.NoError(t, err)
	assert.Equal(t, domain.ShopPurchaseRejected, got.Status)
	assert.Equal(t, "customer did not buy", got.RejectReason)
	assert.Nil(t, f.accountFor(domain.RewardOwnerShop, fxShopID))
	require.Len(t, pusher.sent, 1)
	assert.Equal(t, "rejected", pusher.sent[0].Data["status"])

	_, err = uc.RejectPurchase(context.Background(), fxAdminID, p.ID, "")
	assert.ErrorIs(t, err, ErrPurchaseNotPending)
}

// TestRejectPurchase_TruncatesReasonByRunes guards against cutting a
// multi-byte UTF-8 reason mid-rune: a byte-slice truncation of a 250-rune
// Devanagari reason at 200 bytes would land inside a rune and produce
// invalid UTF-8 that Postgres would reject on write.
func TestRejectPurchase_TruncatesReasonByRunes(t *testing.T) {
	f := newFakeRewardRepo().withOptedInShop(10)
	p := pendingPurchase(f, 45000)
	uc, _ := newTestRewardUseCase(f, fxNow)

	reason := strings.Repeat("क", 250)
	got, err := uc.RejectPurchase(context.Background(), fxAdminID, p.ID, reason)
	require.NoError(t, err)
	assert.True(t, utf8.ValidString(got.RejectReason), "truncated reason must remain valid UTF-8")
	assert.Equal(t, 200, utf8.RuneCountInString(got.RejectReason))
}

func TestListSellerPurchases_ScopedToOwnShop(t *testing.T) {
	f := newFakeRewardRepo().withOptedInShop(10)
	pendingPurchase(f, 45000)
	f.addPurchase(domain.ShopPurchase{ShopID: "shp_2", CustomerID: fxCustomer, Status: domain.ShopPurchasePending})
	uc, _ := newTestRewardUseCase(f, fxNow)
	views, err := uc.ListSellerPurchases(context.Background(), fxAdminID, "", request.Pagination{})
	require.NoError(t, err)
	require.Len(t, views, 1)
	assert.Equal(t, fxShopID, views[0].ShopID)
}
