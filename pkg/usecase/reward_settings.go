package usecase

import (
	"context"
	"errors"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"gorm.io/gorm"
)

// expiringSoonWindow is how far ahead the seller Rewards page warns about expiry.
const expiringSoonWindow = 30 * 24 * time.Hour

type SellerRewardSettings struct {
	ProgramEnabled          bool       `json:"program_enabled"`
	DiscountEnabled         bool       `json:"discount_enabled"`
	DiscountPercent         int        `json:"discount_percent"`
	OptedInAt               *time.Time `json:"opted_in_at"`
	AllowedDiscountPercents []int64    `json:"allowed_discount_percents"`
}

type SellerRewardSummary struct {
	AccountID          string     `json:"account_id"`
	BalancePoints      int64      `json:"balance_points"`
	LifetimeEarned     int64      `json:"lifetime_earned"`
	LifetimeSpent      int64      `json:"lifetime_spent"`
	RedeemThreshold    int64      `json:"redeem_threshold"`
	PointsToThreshold  int64      `json:"points_to_threshold"`
	CanRedeem          bool       `json:"can_redeem"`
	PointValuePaise    int64      `json:"point_value_paise"`
	ExpiringSoonPoints int64      `json:"expiring_soon_points"`
	ExpiringSoonAt     *time.Time `json:"expiring_soon_at"`
}

func (u *RewardUseCase) GetSellerSettings(ctx context.Context, adminID string) (SellerRewardSettings, error) {
	shop, err := u.sellerShop(ctx, adminID)
	if err != nil {
		return SellerRewardSettings{}, err
	}
	cfg, err := u.repo.GetConfig(ctx)
	if err != nil {
		return SellerRewardSettings{}, err
	}
	s, err := u.repo.GetShopSettings(ctx, shop.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return SellerRewardSettings{}, err
	}
	return SellerRewardSettings{
		ProgramEnabled: cfg.ProgramEnabled, DiscountEnabled: s.DiscountEnabled,
		DiscountPercent: s.DiscountPercent, OptedInAt: s.OptedInAt,
		AllowedDiscountPercents: []int64(cfg.AllowedDiscountPercents),
	}, nil
}

// UpdateSellerSettings saves the seller's Yes/No + discount %. Opting in also
// opens the shop's points account so the Rewards page has something to show.
func (u *RewardUseCase) UpdateSellerSettings(ctx context.Context, adminID string, enabled bool, percent int) (SellerRewardSettings, error) {
	shop, err := u.sellerShop(ctx, adminID)
	if err != nil {
		return SellerRewardSettings{}, err
	}
	cfg, err := u.repo.GetConfig(ctx)
	if err != nil {
		return SellerRewardSettings{}, err
	}
	existing, err := u.repo.GetShopSettings(ctx, shop.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return SellerRewardSettings{}, err
	}
	next := domain.ShopRewardSettings{ShopID: shop.ID, AdminID: adminID, OptedInAt: existing.OptedInAt}
	if enabled {
		allowed := false
		for _, p := range cfg.AllowedDiscountPercents {
			if int(p) == percent {
				allowed = true
			}
		}
		if !allowed {
			return SellerRewardSettings{}, ErrInvalidDiscountPercent
		}
		next.DiscountEnabled, next.DiscountPercent = true, percent
		if next.OptedInAt == nil {
			now := u.now()
			next.OptedInAt = &now
		}
	}
	if _, err := u.repo.UpsertShopSettings(ctx, next); err != nil {
		return SellerRewardSettings{}, err
	}
	if enabled {
		if _, err := u.repo.GetOrCreateAccount(ctx, domain.RewardOwnerShop, shop.ID); err != nil {
			return SellerRewardSettings{}, err
		}
	}
	return u.GetSellerSettings(ctx, adminID)
}

func (u *RewardUseCase) GetSellerAccount(ctx context.Context, adminID string) (SellerRewardSummary, error) {
	shop, err := u.sellerShop(ctx, adminID)
	if err != nil {
		return SellerRewardSummary{}, err
	}
	cfg, err := u.repo.GetConfig(ctx)
	if err != nil {
		return SellerRewardSummary{}, err
	}
	acct, err := u.repo.GetOrCreateAccount(ctx, domain.RewardOwnerShop, shop.ID)
	if err != nil {
		return SellerRewardSummary{}, err
	}
	soon, soonAt, err := u.repo.SumExpiringPoints(ctx, acct.ID, u.now().Add(expiringSoonWindow))
	if err != nil {
		return SellerRewardSummary{}, err
	}
	toGo := cfg.SellerMinRedeemBalance - acct.BalancePoints
	if toGo < 0 {
		toGo = 0
	}
	return SellerRewardSummary{
		AccountID: acct.ID, BalancePoints: acct.BalancePoints, LifetimeEarned: acct.LifetimeEarned,
		LifetimeSpent: acct.LifetimeSpent, RedeemThreshold: cfg.SellerMinRedeemBalance,
		PointsToThreshold: toGo, CanRedeem: toGo == 0 && acct.BalancePoints > 0,
		PointValuePaise: cfg.PointValuePaise, ExpiringSoonPoints: soon, ExpiringSoonAt: soonAt,
	}, nil
}

func (u *RewardUseCase) ListSellerLedger(ctx context.Context, adminID string, p request.Pagination) ([]domain.RewardLedgerEntry, error) {
	shop, err := u.sellerShop(ctx, adminID)
	if err != nil {
		return nil, err
	}
	acct, err := u.repo.GetOrCreateAccount(ctx, domain.RewardOwnerShop, shop.ID)
	if err != nil {
		return nil, err
	}
	return u.repo.ListLedger(ctx, acct.ID, p)
}
