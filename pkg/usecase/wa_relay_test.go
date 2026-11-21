package usecase

import "testing"

func TestMaskContactDetails(t *testing.T) {
	cases := map[string]string{
		"call me on 98765 43210 please":       "call me on [number hidden] please",
		"+91-98765-43210":                     "[number hidden]",
		"9 8 7 6 5 4 3 2 1 0":                 "[number hidden]",
		"mail me at a.b@example.com":          "mail me at [email hidden]",
		"chat https://wa.me/919876543210 now": "chat [link hidden] now",
		// Prices, quantities, years and PIN codes are not phone numbers.
		"price 1500 for 20 pcs":     "price 1500 for 20 pcs",
		"deliver in 2026 to 311001": "deliver in 2026 to 311001",
	}
	for in, want := range cases {
		if got := maskContactDetails(in); got != want {
			t.Errorf("maskContactDetails(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestChatRef(t *testing.T) {
	if got := chatRef("3f2a9c1e-5b7d-4e8a-9c21-7d4e5f6a8b90"); got != "#3F2A" {
		t.Errorf("chatRef = %q, want #3F2A", got)
	}
}

func TestRoutingTokens(t *testing.T) {
	body := "Hi, is this available?\n\nSHOP:shp_abc123"
	if id := extractShopID(body); id != "shp_abc123" {
		t.Errorf("extractShopID = %q", id)
	}
	if got := stripShopToken(body); got != "Hi, is this available?" {
		t.Errorf("stripShopToken = %q", got)
	}
	if got := strip91("919876543210"); got != "9876543210" {
		t.Errorf("strip91 = %q", got)
	}
	if got := toE164IN("98765 43210"); got != "919876543210" {
		t.Errorf("toE164IN = %q", got)
	}
}
