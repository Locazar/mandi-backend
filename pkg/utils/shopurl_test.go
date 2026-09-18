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
		{"कुलदेवी ट्युरीजम कैब्स", "Jaipur", "kuladevi-tyurijama-kaibsa-jaipur"},
		{"शर्मा स्टोर", "जयपुर", "sharma-stora-jayapura"},
	}
	for _, c := range cases {
		got := ShopSlug(c.name, c.city)
		if got != c.want {
			t.Errorf("ShopSlug(%q, %q) = %q, want %q", c.name, c.city, got, c.want)
		}
	}
}

func TestShopSlug_DevanagariNeverEmpty(t *testing.T) {
	// The bug this guards against: a Devanagari-only name with no city used
	// to slugify to "" (every rune stripped by the a-z0-9 filter), leaving
	// the shop with no working profile link at all.
	got := ShopSlug("कुलदेवी ट्युरीजम कैब्स", "")
	if got == "" {
		t.Fatal("ShopSlug returned empty for a Devanagari name — shop would have no working link")
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
