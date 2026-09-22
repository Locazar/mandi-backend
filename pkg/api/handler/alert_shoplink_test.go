package handler

import (
	"context"
	"testing"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	usecaseinterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type stubShopLookupAdminUC struct {
	usecaseinterfaces.AdminUseCase
	shop domain.ShopDetails
	err  error
}

func (s *stubShopLookupAdminUC) GetShopByOwnerID(_ context.Context, _ string) (domain.ShopDetails, error) {
	return s.shop, s.err
}

func TestShopLinkFor_ReturnsPublicURL(t *testing.T) {
	h := &AlertHandler{adminUseCase: &stubShopLookupAdminUC{
		shop: domain.ShopDetails{ID: "shp_1", ShopName: "Fashion ForU", City: "Banglore"},
	}}
	got := h.shopLinkFor(context.Background(), "seller_1")
	want := "https://locazar.in/shop/fashion-foru-banglore"
	if got != want {
		t.Errorf("shopLinkFor = %q, want %q", got, want)
	}
}

func TestShopLinkFor_LookupFailsReturnsEmpty(t *testing.T) {
	h := &AlertHandler{adminUseCase: &stubShopLookupAdminUC{err: context.DeadlineExceeded}}
	if got := h.shopLinkFor(context.Background(), "seller_1"); got != "" {
		t.Errorf("expected empty string on lookup failure, got %q", got)
	}
}

func TestSubstitutePlaceholder_ReplacesInNestedContentAndActions(t *testing.T) {
	content := map[string]interface{}{
		"title": "Share {{shop_link}} on your status",
		"actions": []interface{}{
			map[string]interface{}{"label": "Share", "action_type": "share", "link": "{{shop_link}}"},
		},
	}
	link := "https://locazar.in/shop/fashion-foru-banglore"
	got := substitutePlaceholder(content, link)
	m := got.(map[string]interface{})
	if m["title"] != "Share "+link+" on your status" {
		t.Errorf("title = %q", m["title"])
	}
	action := m["actions"].([]interface{})[0].(map[string]interface{})
	if action["link"] != link {
		t.Errorf("action link = %q, want %q", action["link"], link)
	}
}

// This is the bug the fix addresses: an admin can type {{shop_link}} into the
// template's plain top-level Title/Description fields (not inside
// content_schema), which are separate DB columns GetSellerAlerts sends
// straight through — substitutePlaceholder alone never touched them.
func TestSubstitutePlaceholder_WorksOnPlainStringNotJustNestedContent(t *testing.T) {
	link := "https://locazar.in/shop/fashion-foru-banglore"
	got := substitutePlaceholder("Your shop is live! {{shop_link}}", link)
	want := "Your shop is live! " + link
	if got != want {
		t.Errorf("substitutePlaceholder(plain string) = %q, want %q", got, want)
	}
}
