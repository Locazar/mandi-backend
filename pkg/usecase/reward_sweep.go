package usecase

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
)

const (
	rewardSweepBatch   = 500
	rewardReminderLead = 7 * 24 * time.Hour
)

type RewardSweepResult struct {
	ExpiredPurchases int `json:"expired_purchases"`
	ExpiredLots      int `json:"expired_lots"`
	Reminders        int `json:"reminders"`
	Errors           int `json:"errors"`
}

// RunSweep is idempotent: every step is either a conditional UPDATE or guarded
// by the ledger's unique key, so overlapping or repeated runs are harmless.
func (u *RewardUseCase) RunSweep(ctx context.Context) (RewardSweepResult, error) {
	var res RewardSweepResult
	now := u.now()

	n, err := u.repo.ExpireStalePurchases(ctx, now)
	if err != nil {
		return res, fmt.Errorf("expire stale purchases: %w", err)
	}
	res.ExpiredPurchases = int(n)

	lots, err := u.repo.ListExpiredLots(ctx, now, rewardSweepBatch)
	if err != nil {
		return res, fmt.Errorf("list expired lots: %w", err)
	}
	for _, lot := range lots {
		var expired bool
		err := u.repo.InTx(ctx, func(r repo.RewardRepository) error {
			var err error
			expired, err = expireLot(ctx, r, lot)
			return err
		})
		if err != nil {
			log.Printf("WARN [reward sweep]: expire lot %s: %v", lot.ID, err)
			res.Errors++
			continue
		}
		if expired {
			res.ExpiredLots++
		}
	}

	due, err := u.repo.ListLotsDueReminder(ctx, now, now.Add(rewardReminderLead), rewardSweepBatch)
	if err != nil {
		return res, fmt.Errorf("list lots due reminder: %w", err)
	}
	byAccount := map[string][]domain.RewardLedgerEntry{}
	order := []string{}
	for _, lot := range due {
		if _, ok := byAccount[lot.AccountID]; !ok {
			order = append(order, lot.AccountID)
		}
		byAccount[lot.AccountID] = append(byAccount[lot.AccountID], lot)
	}
	for _, accountID := range order {
		if err := u.remindExpiring(ctx, accountID, byAccount[accountID], now); err != nil {
			log.Printf("WARN [reward sweep]: reminder for account %s: %v", accountID, err)
			res.Errors++
			continue
		}
		res.Reminders++
	}
	return res, nil
}

func (u *RewardUseCase) remindExpiring(ctx context.Context, accountID string, lots []domain.RewardLedgerEntry, now time.Time) error {
	acct, err := u.repo.GetAccountByID(ctx, accountID)
	if err != nil {
		return err
	}
	var points int64
	earliest := *lots[0].ExpiresAt
	ids := make([]string, 0, len(lots))
	for _, l := range lots {
		points += l.RemainingPoints
		if l.ExpiresAt.Before(earliest) {
			earliest = *l.ExpiresAt
		}
		ids = append(ids, l.ID)
	}
	req := request.SendPushRequest{
		Title:     "Reward points expiring soon",
		Body:      fmt.Sprintf("%d Locazar reward points expire on %s. Use them before they're gone.", points, earliest.In(rewardTZ).Format("2 Jan")),
		EventType: "reward_points_expiring",
		Data:      map[string]string{"points": strconv.FormatInt(points, 10), "expires_at": earliest.UTC().Format(time.RFC3339)},
	}
	switch acct.OwnerType {
	case domain.RewardOwnerShop:
		shop, err := u.repo.GetShop(ctx, acct.OwnerID)
		if err != nil {
			return err
		}
		req.OwnerID, req.OwnerType = shop.AdminID, "seller"
	default:
		req.OwnerID, req.OwnerType = acct.OwnerID, "user"
	}
	// Mark first: a missed push is better than a duplicate every hour.
	if err := u.repo.MarkLotsReminded(ctx, ids, now); err != nil {
		return err
	}
	u.push(req)
	return nil
}
