package usecase

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

const minAdjustReasonLen = 5

// maxAdjustPoints bounds a single manual adjustment so a typo with an extra
// zero can't mint or burn an absurd number of points.
const maxAdjustPoints int64 = 100_000

// adminRecentLedgerEntries is how much history the admin shop-points view shows.
const adminRecentLedgerEntries = 20

// requirePlatformAdmin guards the program-wide admin operations. The route's
// RequirePermission middleware lets blank-role accounts through, and every
// seller account is blank-role or "seller", so the rewards slice rejects
// both here. Other roles pass; the middleware still enforces the specific
// permission.
func (u *RewardUseCase) requirePlatformAdmin(ctx context.Context, adminID string) error {
	role, err := u.repo.GetAdminRole(ctx, adminID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrRewardAdminOnly
		}
		return err
	}
	if role == "" || role == domain.AdminRoleSeller {
		return ErrRewardAdminOnly
	}
	return nil
}

func (u *RewardUseCase) GetProgramConfig(ctx context.Context, adminID string) (domain.RewardProgramConfig, error) {
	if err := u.requirePlatformAdmin(ctx, adminID); err != nil {
		return domain.RewardProgramConfig{}, err
	}
	return u.repo.GetConfig(ctx)
}

func invalidConfig(format string, args ...interface{}) error {
	return fmt.Errorf("%w: "+format, append([]interface{}{ErrInvalidRewardConfig}, args...)...)
}

func validateRewardConfig(c *domain.RewardProgramConfig) error {
	if len(c.AllowedDiscountPercents) == 0 {
		return invalidConfig("allowed_discount_percents must not be empty")
	}
	seen := map[int64]bool{}
	for _, p := range c.AllowedDiscountPercents {
		if p < 1 || p > 90 {
			return invalidConfig("discount options must be between 1 and 90")
		}
		if seen[p] {
			return invalidConfig("discount option %d is duplicated", p)
		}
		seen[p] = true
	}
	sort.Slice(c.AllowedDiscountPercents, func(i, j int) bool { return c.AllowedDiscountPercents[i] < c.AllowedDiscountPercents[j] })

	for name, v := range map[string]int64{
		"seller_base_points": c.SellerBasePoints, "seller_points_per_100_rupees": c.SellerPointsPer100Rupees,
		"customer_base_points": c.CustomerBasePoints, "customer_points_per_100_rupees": c.CustomerPointsPer100Rupees,
		"min_bill_paise": c.MinBillPaise, "max_customer_points_per_day": c.MaxCustomerPointsPerDay,
		"seller_min_redeem_balance": c.SellerMinRedeemBalance, "customer_min_redeem_balance": c.CustomerMinRedeemBalance,
		"referral_bonus_points": c.ReferralBonusPoints,
	} {
		if v < 0 {
			return invalidConfig("%s must not be negative", name)
		}
	}
	switch {
	case c.SellerMaxPoints <= 0 || c.CustomerMaxPoints <= 0:
		return invalidConfig("max points per purchase must be positive")
	case c.GPSRadiusM < 10 || c.GPSRadiusM > 5000:
		return invalidConfig("gps_radius_m must be between 10 and 5000")
	case c.RepeatWindowHours < 1:
		return invalidConfig("repeat_window_hours must be at least 1")
	case c.ClaimExpiryHours < 1:
		return invalidConfig("claim_expiry_hours must be at least 1")
	case c.MaxClaimsPerShopPerDay < 0 || c.MaxReferralsPerMonth < 0:
		return invalidConfig("caps must not be negative (0 means no cap)")
	case c.PointValuePaise < 1:
		return invalidConfig("point_value_paise must be at least 1")
	case c.CustomerMaxRedeemPctBP < 0 || c.CustomerMaxRedeemPctBP > 10000:
		return invalidConfig("customer_max_redeem_pct_bp must be between 0 and 10000")
	case c.PointsExpiryMonths < 1 || c.PointsExpiryMonths > 60:
		return invalidConfig("points_expiry_months must be between 1 and 60")
	}
	return nil
}

func (u *RewardUseCase) UpdateProgramConfig(ctx context.Context, adminID string, cfg domain.RewardProgramConfig) (domain.RewardProgramConfig, error) {
	if err := u.requirePlatformAdmin(ctx, adminID); err != nil {
		return domain.RewardProgramConfig{}, err
	}
	if err := validateRewardConfig(&cfg); err != nil {
		return domain.RewardProgramConfig{}, err
	}
	cfg.ID = domain.RewardProgramConfigID
	cfg.UpdatedBy = adminID
	return u.repo.UpdateConfig(ctx, cfg)
}

// AdjustAccount is a support tool: a signed manual correction with a reason,
// recorded in the ledger under the acting admin's id. A non-empty
// clientRequestID makes retries of the same submission no-ops.
func (u *RewardUseCase) AdjustAccount(ctx context.Context, adminID, accountID string, delta int64, reason, clientRequestID string) (domain.RewardAccount, error) {
	if err := u.requirePlatformAdmin(ctx, adminID); err != nil {
		return domain.RewardAccount{}, err
	}
	reason = strings.TrimSpace(reason)
	if delta == 0 || delta > maxAdjustPoints || delta < -maxAdjustPoints || len(reason) < minAdjustReasonLen {
		return domain.RewardAccount{}, ErrInvalidAdjustment
	}
	if _, err := u.repo.GetAccountByID(ctx, accountID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return domain.RewardAccount{}, ErrRewardAccountNotFound
		}
		return domain.RewardAccount{}, err
	}
	cfg, err := u.repo.GetConfig(ctx)
	if err != nil {
		return domain.RewardAccount{}, err
	}
	refID := domain.NewID(domain.PrefixRewardAdjust)
	if crid := strings.TrimSpace(clientRequestID); crid != "" {
		refID = "adj:" + crid
	}
	in := creditInput{
		AccountID: accountID, Type: domain.RewardEntryAdminAdjust, RefType: "admin_adjust",
		RefID: refID, Note: reason, CreatedBy: adminID,
	}
	var acct domain.RewardAccount
	err = u.repo.InTx(ctx, func(r repo.RewardRepository) error {
		if delta > 0 {
			expires := u.now().AddDate(0, cfg.PointsExpiryMonths, 0)
			in.Points, in.ExpiresAt = delta, &expires
			_, err = credit(ctx, r, in)
		} else {
			in.Points = -delta
			_, err = debit(ctx, r, in)
		}
		if err != nil {
			return err
		}
		acct, err = r.GetAccountByID(ctx, accountID)
		return err
	})
	return acct, err
}

// AdminShopRewardAccount is what the admin portal shows for one shop's points.
type AdminShopRewardAccount struct {
	ShopID        string                     `json:"shop_id"`
	ShopName      string                     `json:"shop_name"`
	Account       domain.RewardAccount       `json:"account"`
	RecentEntries []domain.RewardLedgerEntry `json:"recent_entries"`
}

// GetShopAccountForAdmin resolves a shop's reward account (creating an empty
// one if the shop never earned points) so an admin can inspect and adjust it.
func (u *RewardUseCase) GetShopAccountForAdmin(ctx context.Context, adminID, shopID string) (AdminShopRewardAccount, error) {
	if err := u.requirePlatformAdmin(ctx, adminID); err != nil {
		return AdminShopRewardAccount{}, err
	}
	shop, err := u.repo.GetShop(ctx, shopID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return AdminShopRewardAccount{}, ErrRewardShopNotFound
	}
	if err != nil {
		return AdminShopRewardAccount{}, err
	}
	acct, err := u.repo.GetOrCreateAccount(ctx, domain.RewardOwnerShop, shop.ID)
	if err != nil {
		return AdminShopRewardAccount{}, err
	}
	entries, err := u.repo.ListLedger(ctx, acct.ID, request.Pagination{Limit: adminRecentLedgerEntries})
	if err != nil {
		return AdminShopRewardAccount{}, err
	}
	return AdminShopRewardAccount{ShopID: shop.ID, ShopName: shop.ShopName, Account: acct, RecentEntries: entries}, nil
}
