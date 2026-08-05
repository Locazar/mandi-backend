package usecase

import (
	"strings"
	"testing"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
)

// TestShopVerificationMessage asserts that every combination of the four
// verification checks routes to the correct seller-facing message: shop photo +
// shop address are the mandatory go-live pair, business + identity docs are
// trust documents that never by themselves make a shop live.
func TestShopVerificationMessage(t *testing.T) {
	mk := func(photo, addr, biz, ident bool) request.ShopVerification {
		return request.ShopVerification{
			ShopId:                     "shp_test",
			Photo_Shop_Verification:    photo,
			Address_Proof_Verification: addr,
			Business_Doc_Verification:  biz,
			Identity_Doc_Verification:  ident,
		}
	}

	cases := []struct {
		name             string
		v                request.ShopVerification
		wantTitleHas     string
		wantBodyContains []string
		wantBodyExcludes []string
	}{
		{
			name:             "all four verified -> full welcome, live",
			v:                mk(true, true, true, true),
			wantTitleHas:     "fully verified",
			wantBodyContains: []string{"LIVE", "Welcome aboard"},
		},
		{
			name:             "mandatory pass, no docs -> live with trust tip",
			v:                mk(true, true, false, false),
			wantTitleHas:     "verified and live",
			wantBodyContains: []string{"LIVE", "business and identity documents"},
		},
		{
			name:             "mandatory pass, business only -> live, identity pending",
			v:                mk(true, true, true, false),
			wantTitleHas:     "verified and live",
			wantBodyContains: []string{"LIVE", "identity document is still pending"},
		},
		{
			name:             "mandatory pass, identity only -> live, business pending",
			v:                mk(true, true, false, true),
			wantTitleHas:     "verified and live",
			wantBodyContains: []string{"LIVE", "business document is still pending"},
		},
		{
			name:             "photo only -> pending/under review",
			v:                mk(true, false, false, false),
			wantTitleHas:     "Action needed",
			wantBodyContains: []string{"under review", "Shop photo: verified", "Shop address: pending"},
			wantBodyExcludes: []string{"LIVE"},
		},
		{
			name:             "address only -> pending/under review",
			v:                mk(false, true, false, false),
			wantTitleHas:     "Action needed",
			wantBodyContains: []string{"under review", "Shop photo: pending", "Shop address: verified"},
			wantBodyExcludes: []string{"LIVE"},
		},
		{
			name:             "docs verified but shop pending -> acknowledges docs, still under review",
			v:                mk(false, false, true, true),
			wantTitleHas:     "Action needed",
			wantBodyContains: []string{"business document and identity document are verified", "under review", "Business document: verified", "Identity document: verified"},
			wantBodyExcludes: []string{"LIVE"},
		},
		{
			name:             "nothing verified -> all pending",
			v:                mk(false, false, false, false),
			wantTitleHas:     "Action needed",
			wantBodyContains: []string{"under review", "Shop photo: pending", "Shop address: pending", "Business document: pending", "Identity document: pending"},
			wantBodyExcludes: []string{"LIVE"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			title, body := shopVerificationMessage(tc.v)
			if !strings.Contains(title, tc.wantTitleHas) {
				t.Errorf("title = %q, want it to contain %q", title, tc.wantTitleHas)
			}
			for _, want := range tc.wantBodyContains {
				if !strings.Contains(body, want) {
					t.Errorf("body = %q\n  want it to contain %q", body, want)
				}
			}
			for _, no := range tc.wantBodyExcludes {
				if strings.Contains(body, no) {
					t.Errorf("body = %q\n  want it to NOT contain %q", body, no)
				}
			}
		})
	}
}

// TestShopVerificationMessage_AllCombinationsProduceMessage guards that every
// one of the 16 combinations returns a non-empty title and body — no branch
// leaves the seller without a notification.
func TestShopVerificationMessage_AllCombinationsProduceMessage(t *testing.T) {
	for i := 0; i < 16; i++ {
		v := request.ShopVerification{
			ShopId:                     "shp_test",
			Photo_Shop_Verification:    i&1 != 0,
			Address_Proof_Verification: i&2 != 0,
			Business_Doc_Verification:  i&4 != 0,
			Identity_Doc_Verification:  i&8 != 0,
		}
		title, body := shopVerificationMessage(v)
		if strings.TrimSpace(title) == "" || strings.TrimSpace(body) == "" {
			t.Errorf("combination %04b produced empty title/body: title=%q body=%q", i, title, body)
		}
	}
}
