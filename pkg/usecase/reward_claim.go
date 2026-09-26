package usecase

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

const maxRejectReasonLen = 200

// sellerShop resolves the seller's own shop, mapping "no shop" to ErrShopNotFound.
func (u *RewardUseCase) sellerShop(ctx context.Context, adminID string) (domain.RewardShop, error) {
	shop, err := u.repo.GetShopByAdminID(ctx, adminID)
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return shop, ErrShopNotFound
	}
	return shop, err
}

// lockOwnPendingPurchase loads a purchase under lock and checks it belongs to
// shopID and can still be decided. Another shop's purchase is reported as not
// found so ids can't be probed.
func (u *RewardUseCase) lockOwnPendingPurchase(ctx context.Context, r repo.RewardRepository, shopID, purchaseID string) (domain.ShopPurchase, error) {
	p, err := r.LockPurchase(ctx, purchaseID)
	if errors.Is(err, gorm.ErrRecordNotFound) || (err == nil && p.ShopID != shopID) {
		return p, ErrPurchaseNotFound
	}
	if err != nil {
		return p, err
	}
	if p.Status != domain.ShopPurchasePending || !u.now().Before(p.ExpiresAt) {
		return p, ErrPurchaseNotPending
	}
	return p, nil
}

// ClaimPurchase is the seller confirming the purchase really happened. It
// credits the seller's points exactly once.
func (u *RewardUseCase) ClaimPurchase(ctx context.Context, sellerAdminID, purchaseID string) (domain.ShopPurchase, error) {
	shop, err := u.sellerShop(ctx, sellerAdminID)
	if err != nil {
		return domain.ShopPurchase{}, err
	}
	now := u.now()
	var p domain.ShopPurchase
	err = u.repo.InTx(ctx, func(r repo.RewardRepository) error {
		if p, err = u.lockOwnPendingPurchase(ctx, r, shop.ID, purchaseID); err != nil {
			return err
		}
		cfg, err := r.GetConfig(ctx)
		if err != nil {
			return err
		}
		if !cfg.ProgramEnabled {
			return ErrRewardProgramDisabled
		}
		if cfg.MaxClaimsPerShopPerDay > 0 {
			n, err := r.CountClaimedSince(ctx, shop.ID, rewardDayStart(now))
			if err != nil {
				return err
			}
			if n >= int64(cfg.MaxClaimsPerShopPerDay) {
				return ErrShopDailyCapReached
			}
		}
		acct, err := r.GetOrCreateAccount(ctx, domain.RewardOwnerShop, shop.ID)
		if err != nil {
			return err
		}
		points := CalculatePoints(cfg.SellerFormula(), p.NetPaidPaise)
		expires := now.AddDate(0, cfg.PointsExpiryMonths, 0)
		if _, err := credit(ctx, r, creditInput{
			AccountID: acct.ID, Points: points, Type: domain.RewardEntryPurchaseEarn,
			RefType: "shop_purchase", RefID: p.ID, ExpiresAt: &expires, CreatedBy: sellerAdminID,
		}); err != nil {
			return err
		}
		p.Status = domain.ShopPurchaseClaimed
		p.SellerPoints = points
		p.ClaimedAt, p.DecidedAt = &now, &now
		p.DecidedBy = sellerAdminID
		return r.UpdatePurchaseDecision(ctx, p)
	})
	if err != nil {
		return domain.ShopPurchase{}, err
	}
	u.notifyCustomerOfDecision(p, shop.ShopName)
	return p, nil
}

// RejectPurchase is the seller saying the purchase did not happen.
func (u *RewardUseCase) RejectPurchase(ctx context.Context, sellerAdminID, purchaseID, reason string) (domain.ShopPurchase, error) {
	shop, err := u.sellerShop(ctx, sellerAdminID)
	if err != nil {
		return domain.ShopPurchase{}, err
	}
	reason = strings.TrimSpace(reason)
	if len(reason) > maxRejectReasonLen {
		reason = reason[:maxRejectReasonLen]
	}
	now := u.now()
	var p domain.ShopPurchase
	err = u.repo.InTx(ctx, func(r repo.RewardRepository) error {
		if p, err = u.lockOwnPendingPurchase(ctx, r, shop.ID, purchaseID); err != nil {
			return err
		}
		p.Status = domain.ShopPurchaseRejected
		p.DecidedAt = &now
		p.DecidedBy = sellerAdminID
		p.RejectReason = reason
		return r.UpdatePurchaseDecision(ctx, p)
	})
	if err != nil {
		return domain.ShopPurchase{}, err
	}
	u.notifyCustomerOfDecision(p, shop.ShopName)
	return p, nil
}

func (u *RewardUseCase) notifyCustomerOfDecision(p domain.ShopPurchase, shopName string) {
	title, body := "Purchase confirmed", fmt.Sprintf("%s confirmed your Locazar purchase of ₹%s.", shopName, rupees(p.NetPaidPaise))
	if p.Status == domain.ShopPurchaseRejected {
		title, body = "Purchase not confirmed", fmt.Sprintf("%s could not confirm your Locazar purchase.", shopName)
	}
	u.push(request.SendPushRequest{
		OwnerID: p.CustomerID, OwnerType: "user", Title: title, Body: body,
		EventType: "reward_purchase_decided",
		Data:      map[string]string{"purchase_id": p.ID, "status": string(p.Status)},
	})
}

func (u *RewardUseCase) ListSellerPurchases(ctx context.Context, sellerAdminID string, status domain.ShopPurchaseStatus, p request.Pagination) ([]domain.ShopPurchaseView, error) {
	shop, err := u.sellerShop(ctx, sellerAdminID)
	if err != nil {
		return nil, err
	}
	return u.repo.ListPurchases(ctx, domain.ShopPurchaseFilter{ShopID: shop.ID, Status: status}, p)
}

func (u *RewardUseCase) ListCustomerPurchases(ctx context.Context, customerID string, p request.Pagination) ([]domain.ShopPurchaseView, error) {
	if _, err := u.repo.GetCustomer(ctx, customerID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrRewardCustomerOnly
		}
		return nil, err
	}
	return u.repo.ListPurchases(ctx, domain.ShopPurchaseFilter{CustomerID: customerID}, p)
}

func (u *RewardUseCase) ListAllPurchases(ctx context.Context, f domain.ShopPurchaseFilter, p request.Pagination) ([]domain.ShopPurchaseView, error) {
	return u.repo.ListPurchases(ctx, f, p)
}
