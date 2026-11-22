package usecase

import (
	"context"
	"testing"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoi "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

// roleRepo fakes only the lookups senderRole uses.
type roleRepo struct {
	repoi.WAChatRepository
	shopPhone string // 10-digit phone of the sender's own shop ("" = not a seller)
	byOutID   map[string]domain.WAChat
	custChat  *domain.WAChat
	sellChat  *domain.WAChat
}

func (r *roleRepo) FindShopByPhone(_ context.Context, p string) (domain.ShopDetails, error) {
	if r.shopPhone != "" && p == r.shopPhone {
		return domain.ShopDetails{ID: "own-shop"}, nil
	}
	return domain.ShopDetails{}, gorm.ErrRecordNotFound
}
func (r *roleRepo) FindChatByOutboundMsgID(_ context.Context, id string) (domain.WAChat, error) {
	if c, ok := r.byOutID[id]; ok {
		return c, nil
	}
	return domain.WAChat{}, gorm.ErrRecordNotFound
}
func (r *roleRepo) FindActiveChatByCustomerPhone(context.Context, string) (domain.WAChat, error) {
	if r.custChat == nil {
		return domain.WAChat{}, gorm.ErrRecordNotFound
	}
	return *r.custChat, nil
}
func (r *roleRepo) FindActiveChatBySellerPhone(context.Context, string) (domain.WAChat, error) {
	if r.sellChat == nil {
		return domain.WAChat{}, gorm.ErrRecordNotFound
	}
	return *r.sellChat, nil
}

func TestSenderRole(t *testing.T) {
	const me = "919876543210"
	ago := func(d time.Duration) *time.Time { t := time.Now().Add(-d); return &t }
	asCustomer := domain.WAChat{CustomerPhone: me, SellerPhone: "911111111111"}
	asSeller := domain.WAChat{CustomerPhone: "912222222222", SellerPhone: me}

	tests := []struct {
		name               string
		repo               *roleRepo
		body, quoted       string
		wantSeller, wantAm bool
	}{
		{"plain customer", &roleRepo{}, "hi", "", false, false},
		{"seller with SHOP token is a customer", &roleRepo{shopPhone: "9876543210"}, "Hello\nSHOP:abc", "", false, false},
		{"seller, only seller chats", &roleRepo{shopPhone: "9876543210", sellChat: &asSeller}, "yes we have", "", true, false},
		{"seller swipes on enquiry", &roleRepo{shopPhone: "9876543210", byOutID: map[string]domain.WAChat{"o1": asSeller}, custChat: &asCustomer, sellChat: &asSeller}, "yes", "o1", true, false},
		{"seller swipes on a shop's reply to them", &roleRepo{shopPhone: "9876543210", byOutID: map[string]domain.WAChat{"o2": asCustomer}, custChat: &asCustomer, sellChat: &asSeller}, "ok thanks", "o2", false, false},
		{"both active today, no swipe", &roleRepo{shopPhone: "9876543210",
			custChat: &domain.WAChat{LastSellerMsgAt: ago(time.Hour)},
			sellChat: &domain.WAChat{LastCustomerMsgAt: ago(2 * time.Hour)}}, "ok", "", false, true},
		{"customer side is old, seller side recent", &roleRepo{shopPhone: "9876543210",
			custChat: &domain.WAChat{LastSellerMsgAt: ago(72 * time.Hour)},
			sellChat: &domain.WAChat{LastCustomerMsgAt: ago(time.Hour)}}, "ok", "", true, false},
		{"seller side is old, customer side recent", &roleRepo{shopPhone: "9876543210",
			custChat: &domain.WAChat{LastSellerMsgAt: ago(time.Hour)},
			sellChat: &domain.WAChat{LastCustomerMsgAt: ago(72 * time.Hour)}}, "ok", "", false, false},
	}
	for _, tc := range tests {
		uc := &waChatUseCase{repo: tc.repo}
		gotSeller, _, gotAm := uc.senderRole(context.Background(), me, tc.body, tc.quoted)
		if gotSeller != tc.wantSeller || gotAm != tc.wantAm {
			t.Errorf("%s: seller=%v ambiguous=%v, want %v %v", tc.name, gotSeller, gotAm, tc.wantSeller, tc.wantAm)
		}
	}
}
