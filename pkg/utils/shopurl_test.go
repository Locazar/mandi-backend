package utils

import "testing"

func TestShopSlug(t *testing.T) {
	cases := []struct {
		name, city, want string
	}{
		{"Fashion ForU", "Banglore", "fashion-foru-banglore"},
		{"Sumona Florist", "", "sumona-florist"},
		{"Café Déjà Vu", "Jaipur", "cafe-deja-vu-jaipur"},
		{"  Sharma  Store!!  ", "Jaipur", "sharma-store-jaipur"},
		{"Jaipur Sweets Jaipur", "Jaipur", "jaipur-sweets-jaipur"},
	}
	for _, c := range cases {
		got := ShopSlug(c.name, c.city)
		if got != c.want {
			t.Errorf("ShopSlug(%q, %q) = %q, want %q", c.name, c.city, got, c.want)
		}
	}
}

func TestShopPublicURL(t *testing.T) {
	got := ShopPublicURL("shp_abc123", "Fashion ForU", "Banglore")
	want := "https://locazar.in/shop/fashion-foru-banglore"
	if got != want {
		t.Errorf("ShopPublicURL = %q, want %q", got, want)
	}

	// No name → falls back to the raw id, same as the TS resolver accepts.
	got = ShopPublicURL("shp_abc123", "", "")
	want = "https://locazar.in/shop/shp_abc123"
	if got != want {
		t.Errorf("ShopPublicURL fallback = %q, want %q", got, want)
	}
}
