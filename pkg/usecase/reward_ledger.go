package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
)

// Lock order everywhere: account row first, then its lots. Never lock a lot
// before its account — that is what keeps debit and the expiry sweep from
// deadlocking each other.

type creditInput struct {
	AccountID string
	Points    int64
	Type      domain.RewardEntryType
	RefType   string
	RefID     string
	ExpiresAt *time.Time // credits only
	Note      string
	CreatedBy string
}

// credit adds a lot to an account. Returns false when this (type, ref) was
// already credited — callers treat that as success.
func credit(ctx context.Context, r repo.RewardRepository, in creditInput) (bool, error) {
	if in.Points <= 0 {
		return false, nil
	}
	if _, err := r.LockAccount(ctx, in.AccountID); err != nil {
		return false, err
	}
	inserted, err := r.InsertLedgerEntry(ctx, domain.RewardLedgerEntry{
		AccountID: in.AccountID, DeltaPoints: in.Points, EntryType: in.Type,
		RefType: in.RefType, RefID: in.RefID, RemainingPoints: in.Points,
		ExpiresAt: in.ExpiresAt, Note: in.Note, CreatedBy: in.CreatedBy,
	})
	if err != nil || !inserted {
		return false, err
	}
	return true, r.AdjustAccountBalance(ctx, in.AccountID, in.Points, in.Points, 0)
}

// debit spends points oldest-expiry-first. Returns ErrInsufficientPoints
// before writing anything when the balance is too low.
func debit(ctx context.Context, r repo.RewardRepository, in creditInput) (bool, error) {
	if in.Points <= 0 {
		return false, nil
	}
	acct, err := r.LockAccount(ctx, in.AccountID)
	if err != nil {
		return false, err
	}
	if acct.BalancePoints < in.Points {
		return false, ErrInsufficientPoints
	}
	inserted, err := r.InsertLedgerEntry(ctx, domain.RewardLedgerEntry{
		AccountID: in.AccountID, DeltaPoints: -in.Points, EntryType: in.Type,
		RefType: in.RefType, RefID: in.RefID, Note: in.Note, CreatedBy: in.CreatedBy,
	})
	if err != nil || !inserted {
		return false, err
	}
	lots, err := r.ListOpenLots(ctx, in.AccountID)
	if err != nil {
		return false, err
	}
	left := in.Points
	for _, lot := range lots {
		if left == 0 {
			break
		}
		take := lot.RemainingPoints
		if take > left {
			take = left
		}
		if err := r.SetLotRemaining(ctx, lot.ID, lot.RemainingPoints-take); err != nil {
			return false, err
		}
		left -= take
	}
	if left > 0 {
		return false, fmt.Errorf("reward ledger: account %s balance %d not backed by open lots", in.AccountID, acct.BalancePoints)
	}
	return true, r.AdjustAccountBalance(ctx, in.AccountID, -in.Points, 0, in.Points)
}

// expireLot writes off whatever is left of one lot. The caller passes a lot
// it has not locked; this locks the account, then re-reads the lot under lock.
func expireLot(ctx context.Context, r repo.RewardRepository, lot domain.RewardLedgerEntry) (bool, error) {
	if _, err := r.LockAccount(ctx, lot.AccountID); err != nil {
		return false, err
	}
	cur, err := r.GetLotForUpdate(ctx, lot.ID)
	if err != nil {
		return false, err
	}
	if cur.RemainingPoints <= 0 {
		return false, nil
	}
	inserted, err := r.InsertLedgerEntry(ctx, domain.RewardLedgerEntry{
		AccountID: cur.AccountID, DeltaPoints: -cur.RemainingPoints, EntryType: domain.RewardEntryExpiry,
		RefType: "ledger_entry", RefID: cur.ID, Note: "points expired",
	})
	if err != nil || !inserted {
		return false, err
	}
	if err := r.SetLotRemaining(ctx, cur.ID, 0); err != nil {
		return false, err
	}
	return true, r.AdjustAccountBalance(ctx, cur.AccountID, -cur.RemainingPoints, 0, 0)
}
