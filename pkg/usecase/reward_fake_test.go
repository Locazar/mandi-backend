package usecase

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoIface "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	service "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"gorm.io/gorm"
)

var _ repoIface.RewardRepository = (*fakeRewardRepo)(nil)

// fakeRewardRepo is an in-memory RewardRepository. InTx does not roll back, so
// tests assert state only on success paths or on failures that occur before
// any write.
type fakeRewardRepo struct {
	cfg       domain.RewardProgramConfig
	shops     map[string]domain.RewardShop
	customers map[string]domain.RewardCustomer
	settings  map[string]domain.ShopRewardSettings
	accounts  map[string]*domain.RewardAccount
	ledger    []*domain.RewardLedgerEntry
	purchases map[string]*domain.ShopPurchase
	seq       int
	pairLocks int
	// calls records the order repository methods were invoked in, for tests
	// that assert on lock-before-read ordering (e.g. LockAccount must precede
	// CountClaimedSince in ClaimPurchase).
	calls []string
}

func defaultRewardConfig() domain.RewardProgramConfig {
	return domain.RewardProgramConfig{
		ID: domain.RewardProgramConfigID, ProgramEnabled: true,
		AllowedDiscountPercents: []int64{5, 10, 15, 20},
		SellerBasePoints:        10, SellerPointsPer100Rupees: 1, SellerMaxPoints: 50,
		CustomerBasePoints: 5, CustomerPointsPer100Rupees: 1, CustomerMaxPoints: 25,
		MinBillPaise: 10000, GPSRadiusM: 200, RepeatWindowHours: 168, ClaimExpiryHours: 48,
		MaxClaimsPerShopPerDay: 30, MaxCustomerPointsPerDay: 100, PointValuePaise: 100,
		SellerMinRedeemBalance: 1000, CustomerMinRedeemBalance: 200, CustomerMaxRedeemPctBP: 1000,
		PointsExpiryMonths: 12, ReferralBonusPoints: 50, MaxReferralsPerMonth: 10,
	}
}

func newFakeRewardRepo() *fakeRewardRepo {
	return &fakeRewardRepo{
		cfg:       defaultRewardConfig(),
		shops:     map[string]domain.RewardShop{},
		customers: map[string]domain.RewardCustomer{},
		settings:  map[string]domain.ShopRewardSettings{},
		accounts:  map[string]*domain.RewardAccount{},
		purchases: map[string]*domain.ShopPurchase{},
	}
}

func (f *fakeRewardRepo) nextID(prefix string) string {
	f.seq++
	return fmt.Sprintf("%s_%d", prefix, f.seq)
}

// Fixture helpers.
const (
	fxShopID   = "shp_1"
	fxAdminID  = "adm_1"
	fxCustomer = "usr_1"
	fxShopLat  = 12.9756
	fxShopLng  = 77.6050
)

func (f *fakeRewardRepo) withOptedInShop(percent int) *fakeRewardRepo {
	f.shops[fxShopID] = domain.RewardShop{ID: fxShopID, AdminID: fxAdminID, ShopName: "Asha Stores",
		ShopStatus: domain.ShopStatusActive, Latitude: fxShopLat, Longitude: fxShopLng, OwnerMobile: "9000000001"}
	f.customers[fxCustomer] = domain.RewardCustomer{ID: fxCustomer, FirstName: "Ravi", Phone: "+91 98765 43210"}
	f.settings[fxShopID] = domain.ShopRewardSettings{ShopID: fxShopID, AdminID: fxAdminID, DiscountEnabled: true, DiscountPercent: percent}
	return f
}

func (f *fakeRewardRepo) addPurchase(p domain.ShopPurchase) *domain.ShopPurchase {
	if p.ID == "" {
		p.ID = f.nextID("spur")
	}
	cp := p
	f.purchases[p.ID] = &cp
	return &cp
}

func (f *fakeRewardRepo) accountFor(ownerType domain.RewardOwnerType, ownerID string) *domain.RewardAccount {
	for _, a := range f.accounts {
		if a.OwnerType == ownerType && a.OwnerID == ownerID {
			return a
		}
	}
	return nil
}

func (f *fakeRewardRepo) entries(accountID string, t domain.RewardEntryType) []*domain.RewardLedgerEntry {
	out := []*domain.RewardLedgerEntry{}
	for _, e := range f.ledger {
		if e.AccountID == accountID && (t == "" || e.EntryType == t) {
			out = append(out, e)
		}
	}
	return out
}

// ── RewardRepository ────────────────────────────────────────────────────────

func (f *fakeRewardRepo) InTx(_ context.Context, fn func(repoIface.RewardRepository) error) error {
	return fn(f)
}
func (f *fakeRewardRepo) GetConfig(context.Context) (domain.RewardProgramConfig, error) {
	return f.cfg, nil
}
func (f *fakeRewardRepo) UpdateConfig(_ context.Context, c domain.RewardProgramConfig) (domain.RewardProgramConfig, error) {
	c.ID = domain.RewardProgramConfigID
	f.cfg = c
	return c, nil
}
func (f *fakeRewardRepo) GetShop(_ context.Context, id string) (domain.RewardShop, error) {
	s, ok := f.shops[id]
	if !ok {
		return s, gorm.ErrRecordNotFound
	}
	return s, nil
}
func (f *fakeRewardRepo) GetShopByAdminID(_ context.Context, adminID string) (domain.RewardShop, error) {
	for _, s := range f.shops {
		if s.AdminID == adminID {
			return s, nil
		}
	}
	return domain.RewardShop{}, gorm.ErrRecordNotFound
}
func (f *fakeRewardRepo) GetCustomer(_ context.Context, id string) (domain.RewardCustomer, error) {
	c, ok := f.customers[id]
	if !ok {
		return c, gorm.ErrRecordNotFound
	}
	return c, nil
}
func (f *fakeRewardRepo) GetShopSettings(_ context.Context, shopID string) (domain.ShopRewardSettings, error) {
	s, ok := f.settings[shopID]
	if !ok {
		return s, gorm.ErrRecordNotFound
	}
	return s, nil
}
func (f *fakeRewardRepo) UpsertShopSettings(_ context.Context, s domain.ShopRewardSettings) (domain.ShopRewardSettings, error) {
	f.settings[s.ShopID] = s
	return s, nil
}
func (f *fakeRewardRepo) GetOrCreateAccount(_ context.Context, t domain.RewardOwnerType, ownerID string) (domain.RewardAccount, error) {
	if a := f.accountFor(t, ownerID); a != nil {
		return *a, nil
	}
	a := &domain.RewardAccount{ID: f.nextID("rwa"), OwnerType: t, OwnerID: ownerID}
	f.accounts[a.ID] = a
	return *a, nil
}
func (f *fakeRewardRepo) GetAccountByID(_ context.Context, id string) (domain.RewardAccount, error) {
	a, ok := f.accounts[id]
	if !ok {
		return domain.RewardAccount{}, gorm.ErrRecordNotFound
	}
	return *a, nil
}
func (f *fakeRewardRepo) LockAccount(ctx context.Context, id string) (domain.RewardAccount, error) {
	f.calls = append(f.calls, "LockAccount")
	return f.GetAccountByID(ctx, id)
}
func (f *fakeRewardRepo) AdjustAccountBalance(_ context.Context, id string, delta, earned, spent int64) error {
	a := f.accounts[id]
	if a.BalancePoints+delta < 0 {
		return fmt.Errorf("balance check violated")
	}
	a.BalancePoints += delta
	a.LifetimeEarned += earned
	a.LifetimeSpent += spent
	return nil
}
func (f *fakeRewardRepo) InsertLedgerEntry(_ context.Context, e domain.RewardLedgerEntry) (bool, error) {
	for _, x := range f.ledger {
		if x.AccountID == e.AccountID && x.EntryType == e.EntryType && x.RefID == e.RefID {
			return false, nil
		}
	}
	e.ID = f.nextID("rwl")
	if e.CreatedAt.IsZero() {
		e.CreatedAt = time.Unix(int64(f.seq), 0)
	}
	cp := e
	f.ledger = append(f.ledger, &cp)
	return true, nil
}
func (f *fakeRewardRepo) LedgerEntryExists(_ context.Context, accountID string, entryType domain.RewardEntryType, refID string) (bool, error) {
	for _, e := range f.ledger {
		if e.AccountID == accountID && e.EntryType == entryType && e.RefID == refID {
			return true, nil
		}
	}
	return false, nil
}
func (f *fakeRewardRepo) ListOpenLots(_ context.Context, accountID string) ([]domain.RewardLedgerEntry, error) {
	out := []domain.RewardLedgerEntry{}
	for _, e := range f.ledger {
		if e.AccountID == accountID && e.RemainingPoints > 0 {
			out = append(out, *e)
		}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i].ExpiresAt, out[j].ExpiresAt
		switch {
		case a == nil && b == nil:
			return out[i].CreatedAt.Before(out[j].CreatedAt)
		case a == nil:
			return false
		case b == nil:
			return true
		default:
			return a.Before(*b)
		}
	})
	return out, nil
}
func (f *fakeRewardRepo) GetLotForUpdate(_ context.Context, id string) (domain.RewardLedgerEntry, error) {
	for _, e := range f.ledger {
		if e.ID == id {
			return *e, nil
		}
	}
	return domain.RewardLedgerEntry{}, gorm.ErrRecordNotFound
}
func (f *fakeRewardRepo) SetLotRemaining(_ context.Context, id string, remaining int64) error {
	for _, e := range f.ledger {
		if e.ID == id {
			e.RemainingPoints = remaining
		}
	}
	return nil
}
func (f *fakeRewardRepo) ListLedger(_ context.Context, accountID string, _ request.Pagination) ([]domain.RewardLedgerEntry, error) {
	out := []domain.RewardLedgerEntry{}
	for _, e := range f.entries(accountID, "") {
		out = append(out, *e)
	}
	return out, nil
}
func (f *fakeRewardRepo) SumExpiringPoints(_ context.Context, accountID string, before time.Time) (int64, *time.Time, error) {
	var total int64
	var earliest *time.Time
	for _, e := range f.ledger {
		if e.AccountID == accountID && e.RemainingPoints > 0 && e.ExpiresAt != nil && !e.ExpiresAt.After(before) {
			total += e.RemainingPoints
			if earliest == nil || e.ExpiresAt.Before(*earliest) {
				t := *e.ExpiresAt
				earliest = &t
			}
		}
	}
	return total, earliest, nil
}
func (f *fakeRewardRepo) ListExpiredLots(_ context.Context, now time.Time, _ int) ([]domain.RewardLedgerEntry, error) {
	out := []domain.RewardLedgerEntry{}
	for _, e := range f.ledger {
		if e.RemainingPoints > 0 && e.ExpiresAt != nil && !e.ExpiresAt.After(now) {
			out = append(out, *e)
		}
	}
	return out, nil
}
func (f *fakeRewardRepo) ListLotsDueReminder(_ context.Context, from, to time.Time, _ int) ([]domain.RewardLedgerEntry, error) {
	out := []domain.RewardLedgerEntry{}
	for _, e := range f.ledger {
		if e.RemainingPoints > 0 && e.ExpiryRemindedAt == nil && e.ExpiresAt != nil && e.ExpiresAt.After(from) && !e.ExpiresAt.After(to) {
			out = append(out, *e)
		}
	}
	return out, nil
}
func (f *fakeRewardRepo) MarkLotsReminded(_ context.Context, ids []string, at time.Time) error {
	for _, e := range f.ledger {
		for _, id := range ids {
			if e.ID == id {
				t := at
				e.ExpiryRemindedAt = &t
			}
		}
	}
	return nil
}
func (f *fakeRewardRepo) LockCustomerShopPair(context.Context, string, string) error {
	f.pairLocks++
	return nil
}
func (f *fakeRewardRepo) FindPurchaseByClientRequest(_ context.Context, customerID, crid string) (*domain.ShopPurchase, error) {
	for _, p := range f.purchases {
		if p.CustomerID == customerID && p.ClientRequestID == crid {
			cp := *p
			return &cp, nil
		}
	}
	return nil, nil
}
func blocks(p *domain.ShopPurchase, now time.Time) bool {
	return p.Status == domain.ShopPurchaseClaimed || (p.Status == domain.ShopPurchasePending && p.ExpiresAt.After(now))
}
func (f *fakeRewardRepo) FindBlockingPurchase(_ context.Context, customerID, shopID string, since, now time.Time) (*domain.ShopPurchase, error) {
	var best *domain.ShopPurchase
	for _, p := range f.purchases {
		if p.CustomerID == customerID && p.ShopID == shopID && p.CreatedAt.After(since) && blocks(p, now) {
			if best == nil || p.CreatedAt.After(best.CreatedAt) {
				cp := *p
				best = &cp
			}
		}
	}
	return best, nil
}
func (f *fakeRewardRepo) CountOpenOrClaimedCreatedSince(_ context.Context, shopID string, since, now time.Time) (int64, error) {
	var n int64
	for _, p := range f.purchases {
		if p.ShopID == shopID && !p.CreatedAt.Before(since) && blocks(p, now) {
			n++
		}
	}
	return n, nil
}
func (f *fakeRewardRepo) CountClaimedSince(_ context.Context, shopID string, since time.Time) (int64, error) {
	f.calls = append(f.calls, "CountClaimedSince")
	var n int64
	for _, p := range f.purchases {
		if p.ShopID == shopID && p.Status == domain.ShopPurchaseClaimed && p.ClaimedAt != nil && !p.ClaimedAt.Before(since) {
			n++
		}
	}
	return n, nil
}
func (f *fakeRewardRepo) CustomerShopHistory(_ context.Context, customerID, shopID, exclude string) (int64, *domain.ShopPurchase, error) {
	var n int64
	var last *domain.ShopPurchase
	for _, p := range f.purchases {
		if p.CustomerID == customerID && p.ShopID == shopID && p.Status == domain.ShopPurchaseClaimed && p.ID != exclude {
			n++
			if last == nil || p.CreatedAt.After(last.CreatedAt) {
				cp := *p
				last = &cp
			}
		}
	}
	return n, last, nil
}
func (f *fakeRewardRepo) CreatePurchase(_ context.Context, p domain.ShopPurchase) (domain.ShopPurchase, error) {
	p.ID = f.nextID("spur")
	cp := p
	f.purchases[p.ID] = &cp
	return p, nil
}
func (f *fakeRewardRepo) LockPurchase(_ context.Context, id string) (domain.ShopPurchase, error) {
	p, ok := f.purchases[id]
	if !ok {
		return domain.ShopPurchase{}, gorm.ErrRecordNotFound
	}
	return *p, nil
}
func (f *fakeRewardRepo) UpdatePurchaseDecision(_ context.Context, p domain.ShopPurchase) error {
	cur := f.purchases[p.ID]
	cur.Status, cur.SellerPoints, cur.CustomerPoints = p.Status, p.SellerPoints, p.CustomerPoints
	cur.ClaimedAt, cur.DecidedAt, cur.DecidedBy, cur.RejectReason = p.ClaimedAt, p.DecidedAt, p.DecidedBy, p.RejectReason
	return nil
}
func (f *fakeRewardRepo) ExpireStalePurchases(_ context.Context, now time.Time) (int64, error) {
	var n int64
	for _, p := range f.purchases {
		if p.Status == domain.ShopPurchasePending && !p.ExpiresAt.After(now) {
			p.Status = domain.ShopPurchaseExpired
			t := now
			p.DecidedAt = &t
			n++
		}
	}
	return n, nil
}
func (f *fakeRewardRepo) ListPurchases(_ context.Context, flt domain.ShopPurchaseFilter, _ request.Pagination) ([]domain.ShopPurchaseView, error) {
	out := []domain.ShopPurchaseView{}
	for _, p := range f.purchases {
		if (flt.ShopID == "" || p.ShopID == flt.ShopID) && (flt.CustomerID == "" || p.CustomerID == flt.CustomerID) && (flt.Status == "" || p.Status == flt.Status) {
			out = append(out, domain.ShopPurchaseView{ShopPurchase: *p})
		}
	}
	return out, nil
}

// ── push stub ───────────────────────────────────────────────────────────────

type stubPusher struct {
	service.NotificationUseCase
	mu   sync.Mutex
	sent []request.SendPushRequest
}

func (s *stubPusher) SendPushNotification(_ context.Context, req request.SendPushRequest) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sent = append(s.sent, req)
	return true, nil
}

func newTestRewardUseCase(f *fakeRewardRepo, now time.Time) (*RewardUseCase, *stubPusher) {
	p := &stubPusher{}
	uc := NewRewardUseCase(f, p)
	uc.now = func() time.Time { return now }
	uc.dispatch = func(fn func()) { fn() }
	return uc, p
}
