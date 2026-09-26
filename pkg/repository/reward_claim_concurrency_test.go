package repository

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Two devices tapping "Claim Rewards" at once must credit the seller once.
func TestRewardClaim_ConcurrentClaimsCreditOnce(t *testing.T) {
	db := testDB(t)
	repo := NewRewardRepository(db)
	ctx := context.Background()
	s := seedRewardFixtures(t, db)

	p, err := repo.CreatePurchase(ctx, domain.ShopPurchase{
		ShopID: s.shopID, CustomerID: s.userID, ClientRequestID: "crid-claim",
		BillAmountPaise: 50000, DiscountPercent: 10, DiscountPaise: 5000, NetPaidPaise: 45000,
		Status: domain.ShopPurchasePending, ExpiresAt: time.Now().Add(48 * time.Hour),
	})
	require.NoError(t, err)

	uc := usecase.NewRewardUseCase(repo, nil)
	const workers = 8
	var wg sync.WaitGroup
	var mu sync.Mutex
	successes := 0
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := uc.ClaimPurchase(ctx, s.adminID, p.ID); err == nil {
				mu.Lock()
				successes++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	assert.Equal(t, 1, successes)

	acct, err := repo.GetOrCreateAccount(ctx, domain.RewardOwnerShop, s.shopID)
	require.NoError(t, err)
	assert.Equal(t, int64(14), acct.BalancePoints)
}
