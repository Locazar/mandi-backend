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

func TestSubstituteShopLink_ReplacesPlaceholderInActionsAndTitle(t *testing.T) {
	h := &AlertHandler{adminUseCase: &stubShopLookupAdminUC{
		shop: domain.ShopDetails{ID: "shp_1", ShopName: "Fashion ForU", City: "Banglore"},
	}}
	content := map[string]interface{}{
		"title": "Share {{shop_link}} on your status",
		"actions": []interface{}{
			map[string]interface{}{"label": "Share", "action_type": "share", "link": "{{shop_link}}"},
		},
	}
	got := h.substituteShopLink(context.Background(), "seller_1", content)
	m := got.(map[string]interface{})
	wantLink := "https://locazar.in/shop/fashion-foru-banglore"
	if m["title"] != "Share "+wantLink+" on your status" {
		t.Errorf("title = %q", m["title"])
	}
	action := m["actions"].([]interface{})[0].(map[string]interface{})
	if action["link"] != wantLink {
		t.Errorf("action link = %q, want %q", action["link"], wantLink)
	}
}

func TestSubstituteShopLink_LookupFailsLeavesPlaceholderAsIs(t *testing.T) {
	h := &AlertHandler{adminUseCase: &stubShopLookupAdminUC{err: context.DeadlineExceeded}}
	content := map[string]interface{}{"title": "{{shop_link}}"}
	got := h.substituteShopLink(context.Background(), "seller_1", content)
	if got.(map[string]interface{})["title"] != "{{shop_link}}" {
		t.Errorf("expected placeholder left untouched on lookup failure, got %v", got)
	}
}
