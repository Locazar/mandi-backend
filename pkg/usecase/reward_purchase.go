package usecase

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type RewardEligibility struct {
	ShopID          string     `json:"shop_id"`
	ShopName        string     `json:"shop_name"`
	DiscountPercent int        `json:"discount_percent"`
	MinBillPaise    int64      `json:"min_bill_paise"`
	Eligible        bool       `json:"eligible"`
	Reason          string     `json:"reason,omitempty"`
	NextAvailableAt *time.Time `json:"next_available_at,omitempty"`
	DistanceM       int        `json:"distance_m"`
	GPSRadiusM      int        `json:"gps_radius_m"`
}

type SubmitPurchaseInput struct {
	CustomerID      string
	ShopID          string
	ClientRequestID string
	BillAmountPaise int64
	Lat             float64
	Lng             float64
}

// rewardReasonCodes are the machine-readable reasons the customer app switches
// on to show a message instead of the Buy button.
var rewardReasonCodes = []struct {
	err  error
	code string
}{
	{ErrRewardProgramDisabled, "program_disabled"},
	{ErrShopNotOptedIn, "shop_not_opted_in"},
	{ErrSelfPurchase, "self_purchase"},
	{ErrShopLocationMissing, "shop_location_missing"},
	{ErrInvalidLocation, "invalid_location"},
	{ErrTooFarFromShop, "too_far"},
	{ErrRepeatPurchaseWindow, "repeat_window"},
	{ErrShopDailyCapReached, "shop_daily_cap"},
}

func rewardReasonCode(err error) string {
	for _, rc := range rewardReasonCodes {
		if errors.Is(err, rc.err) {
			return rc.code
		}
	}
	return ""
}

// purchaseContext is everything evaluatePurchase learned, filled in as far as
// evaluation got before a rule failed.
type purchaseContext struct {
	cfg             domain.RewardProgramConfig
	customer        domain.RewardCustomer
	shop            domain.RewardShop
	settings        domain.ShopRewardSettings
	distanceM       int
	nextAvailableAt *time.Time
}

// evaluatePurchase applies every "may this customer buy here now?" rule in a
// fixed order. Pass the tx-bound repo when called from SubmitPurchase so the
// repeat-window read happens under the pair lock.
func (u *RewardUseCase) evaluatePurchase(ctx context.Context, r repo.RewardRepository, customerID, shopID string, lat, lng float64, now time.Time) (purchaseContext, error) {
	var pc purchaseContext
	if !validCoordinates(lat, lng) {
		return pc, ErrInvalidLocation
	}
	var err error
	if pc.cfg, err = r.GetConfig(ctx); err != nil {
		return pc, err
	}
	if pc.customer, err = r.GetCustomer(ctx, customerID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pc, ErrRewardCustomerOnly
		}
		return pc, err
	}
	if pc.shop, err = r.GetShop(ctx, shopID); err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return pc, ErrRewardShopNotFound
		}
		return pc, err
	}
	if !pc.cfg.ProgramEnabled {
		return pc, ErrRewardProgramDisabled
	}
	settings, err := r.GetShopSettings(ctx, shopID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return pc, err
	}
	pc.settings = settings
	if pc.shop.ShopStatus != domain.ShopStatusActive || !settings.DiscountEnabled {
		return pc, ErrShopNotOptedIn
	}
	if cp := normalizePhone(pc.customer.Phone); cp != "" && cp == normalizePhone(pc.shop.OwnerMobile) {
		return pc, ErrSelfPurchase
	}
	if pc.shop.Latitude == 0 && pc.shop.Longitude == 0 {
		return pc, ErrShopLocationMissing
	}
	pc.distanceM = int(math.Round(distanceMeters(lat, lng, pc.shop.Latitude, pc.shop.Longitude)))
	if pc.distanceM > pc.cfg.GPSRadiusM {
		return pc, ErrTooFarFromShop
	}
	window := time.Duration(pc.cfg.RepeatWindowHours) * time.Hour
	blocking, err := r.FindBlockingPurchase(ctx, customerID, shopID, now.Add(-window), now)
	if err != nil {
		return pc, err
	}
	if blocking != nil {
		next := blocking.CreatedAt.Add(window)
		pc.nextAvailableAt = &next
		return pc, ErrRepeatPurchaseWindow
	}
	if pc.cfg.MaxClaimsPerShopPerDay > 0 {
		n, err := r.CountOpenOrClaimedCreatedSince(ctx, shopID, rewardDayStart(now), now)
		if err != nil {
			return pc, err
		}
		if n >= int64(pc.cfg.MaxClaimsPerShopPerDay) {
			return pc, ErrShopDailyCapReached
		}
	}
	return pc, nil
}

// GetEligibility tells the customer app what to show after a scan. Rule
// failures come back as Eligible=false + Reason; only identity/lookup failures
// and infrastructure errors are returned as errors.
func (u *RewardUseCase) GetEligibility(ctx context.Context, customerID, shopID string, lat, lng float64) (RewardEligibility, error) {
	pc, err := u.evaluatePurchase(ctx, u.repo, customerID, shopID, lat, lng, u.now())
	e := RewardEligibility{
		ShopID: shopID, ShopName: pc.shop.ShopName, DiscountPercent: pc.settings.DiscountPercent,
		MinBillPaise: pc.cfg.MinBillPaise, DistanceM: pc.distanceM, GPSRadiusM: pc.cfg.GPSRadiusM,
		NextAvailableAt: pc.nextAvailableAt,
	}
	if err == nil {
		e.Eligible = true
		return e, nil
	}
	if code := rewardReasonCode(err); code != "" {
		e.Reason = code
		return e, nil
	}
	return RewardEligibility{}, err
}

// SubmitPurchase records a pending purchase and notifies the seller. A repeated
// ClientRequestID returns the original purchase without side effects.
func (u *RewardUseCase) SubmitPurchase(ctx context.Context, in SubmitPurchaseInput) (domain.ShopPurchase, error) {
	if in.BillAmountPaise <= 0 || in.BillAmountPaise > maxBillPaise {
		return domain.ShopPurchase{}, ErrInvalidBillAmount
	}
	now := u.now()
	var (
		result  domain.ShopPurchase
		created bool
		pc      purchaseContext
	)
	err := u.repo.InTx(ctx, func(r repo.RewardRepository) error {
		if err := r.LockCustomerShopPair(ctx, in.CustomerID, in.ShopID); err != nil {
			return err
		}
		existing, err := r.FindPurchaseByClientRequest(ctx, in.CustomerID, in.ClientRequestID)
		if err != nil {
			return err
		}
		if existing != nil {
			result = *existing
			return nil
		}
		if pc, err = u.evaluatePurchase(ctx, r, in.CustomerID, in.ShopID, in.Lat, in.Lng, now); err != nil {
			return err
		}
		if in.BillAmountPaise < pc.cfg.MinBillPaise {
			return ErrBelowMinBill
		}
		discount := discountPaise(in.BillAmountPaise, pc.settings.DiscountPercent)
		result, err = r.CreatePurchase(ctx, domain.ShopPurchase{
			ShopID: in.ShopID, CustomerID: in.CustomerID, ClientRequestID: in.ClientRequestID,
			BillAmountPaise: in.BillAmountPaise, DiscountPercent: pc.settings.DiscountPercent,
			DiscountPaise: discount, NetPaidPaise: in.BillAmountPaise - discount,
			CustomerLat: in.Lat, CustomerLng: in.Lng, DistanceM: pc.distanceM,
			Status:    domain.ShopPurchasePending,
			ExpiresAt: now.Add(time.Duration(pc.cfg.ClaimExpiryHours) * time.Hour),
			CreatedAt: now,
		})
		created = err == nil
		return err
	})
	if err != nil {
		return domain.ShopPurchase{}, err
	}
	if created {
		u.notifySellerOfPurchase(ctx, pc, result)
	}
	return result, nil
}

func (u *RewardUseCase) notifySellerOfPurchase(ctx context.Context, pc purchaseContext, p domain.ShopPurchase) {
	visits, last, err := u.repo.CustomerShopHistory(ctx, p.CustomerID, p.ShopID, p.ID)
	if err != nil {
		visits, last = 0, nil
	}
	name := maskedCustomerName(pc.customer.FirstName, pc.customer.LastName)
	data := map[string]string{
		"purchase_id":       p.ID,
		"shop_id":           p.ShopID,
		"bill_amount_paise": strconv.FormatInt(p.BillAmountPaise, 10),
		"discount_percent":  strconv.Itoa(p.DiscountPercent),
		"net_paid_paise":    strconv.FormatInt(p.NetPaidPaise, 10),
		"customer_name":     name,
		"visit_count":       strconv.FormatInt(visits, 10),
	}
	if last != nil {
		data["last_purchase_paise"] = strconv.FormatInt(last.BillAmountPaise, 10)
		data["last_purchase_at"] = last.CreatedAt.UTC().Format(time.RFC3339)
	}
	u.push(request.SendPushRequest{
		OwnerID:   pc.shop.AdminID,
		OwnerType: "seller",
		Title:     "New Locazar purchase",
		Body:      fmt.Sprintf("%s paid ₹%s with your %d%% Locazar discount. Tap to claim your reward points.", name, rupees(p.NetPaidPaise), p.DiscountPercent),
		EventType: "reward_purchase",
		Data:      data,
	})
}

// maskedCustomerName renders a customer's display name for the seller push:
// "Ravi K." (first name + first letter of last name) when both are known,
// just the first name when there's no last name, and a generic fallback when
// neither is set. The full name is never shown — the spec requires the
// customer be identified to the seller only in masked form.
func maskedCustomerName(first, last string) string {
	first = strings.TrimSpace(first)
	last = strings.TrimSpace(last)
	if first == "" {
		return "A Locazar customer"
	}
	if last == "" {
		return first
	}
	initial := []rune(last)[0]
	return first + " " + strings.ToUpper(string(initial)) + "."
}

// rupees renders paise as a rupee amount, dropping ".00".
func rupees(paise int64) string {
	if paise%100 == 0 {
		return strconv.FormatInt(paise/100, 10)
	}
	return fmt.Sprintf("%d.%02d", paise/100, paise%100)
}
