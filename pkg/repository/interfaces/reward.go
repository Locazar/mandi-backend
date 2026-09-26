package interfaces

import (
	"context"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// RewardRepository persists the rewards program. Methods that lock rows
// (Lock*, GetLotForUpdate, ListOpenLots, LockCustomerShopPair) only make sense
// on the repository handed to InTx's callback.
type RewardRepository interface {
	// InTx runs fn in one database transaction with a transaction-bound repository.
	InTx(ctx context.Context, fn func(r RewardRepository) error) error

	GetConfig(ctx context.Context) (domain.RewardProgramConfig, error)
	UpdateConfig(ctx context.Context, cfg domain.RewardProgramConfig) (domain.RewardProgramConfig, error)

	GetShop(ctx context.Context, shopID string) (domain.RewardShop, error)
	GetShopByAdminID(ctx context.Context, adminID string) (domain.RewardShop, error)
	GetCustomer(ctx context.Context, userID string) (domain.RewardCustomer, error)

	GetShopSettings(ctx context.Context, shopID string) (domain.ShopRewardSettings, error)
	UpsertShopSettings(ctx context.Context, s domain.ShopRewardSettings) (domain.ShopRewardSettings, error)

	GetOrCreateAccount(ctx context.Context, ownerType domain.RewardOwnerType, ownerID string) (domain.RewardAccount, error)
	GetAccountByID(ctx context.Context, accountID string) (domain.RewardAccount, error)
	LockAccount(ctx context.Context, accountID string) (domain.RewardAccount, error)
	AdjustAccountBalance(ctx context.Context, accountID string, delta, earnedDelta, spentDelta int64) error

	// InsertLedgerEntry returns false (and no error) when an entry with the same
	// (account_id, entry_type, ref_id) already exists.
	InsertLedgerEntry(ctx context.Context, e domain.RewardLedgerEntry) (bool, error)
	ListOpenLots(ctx context.Context, accountID string) ([]domain.RewardLedgerEntry, error)
	GetLotForUpdate(ctx context.Context, entryID string) (domain.RewardLedgerEntry, error)
	SetLotRemaining(ctx context.Context, entryID string, remaining int64) error
	ListLedger(ctx context.Context, accountID string, p request.Pagination) ([]domain.RewardLedgerEntry, error)
	SumExpiringPoints(ctx context.Context, accountID string, before time.Time) (int64, *time.Time, error)
	ListExpiredLots(ctx context.Context, now time.Time, limit int) ([]domain.RewardLedgerEntry, error)
	ListLotsDueReminder(ctx context.Context, from, to time.Time, limit int) ([]domain.RewardLedgerEntry, error)
	MarkLotsReminded(ctx context.Context, entryIDs []string, at time.Time) error

	LockCustomerShopPair(ctx context.Context, customerID, shopID string) error
	FindPurchaseByClientRequest(ctx context.Context, customerID, clientRequestID string) (*domain.ShopPurchase, error)
	FindBlockingPurchase(ctx context.Context, customerID, shopID string, since, now time.Time) (*domain.ShopPurchase, error)
	CountOpenOrClaimedCreatedSince(ctx context.Context, shopID string, since, now time.Time) (int64, error)
	CountClaimedSince(ctx context.Context, shopID string, since time.Time) (int64, error)
	CustomerShopHistory(ctx context.Context, customerID, shopID, excludePurchaseID string) (int64, *domain.ShopPurchase, error)
	CreatePurchase(ctx context.Context, p domain.ShopPurchase) (domain.ShopPurchase, error)
	LockPurchase(ctx context.Context, purchaseID string) (domain.ShopPurchase, error)
	UpdatePurchaseDecision(ctx context.Context, p domain.ShopPurchase) error
	ExpireStalePurchases(ctx context.Context, now time.Time) (int64, error)
	ListPurchases(ctx context.Context, f domain.ShopPurchaseFilter, p request.Pagination) ([]domain.ShopPurchaseView, error)
}
