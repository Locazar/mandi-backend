package notification

import (
	"os"
	"testing"
)

// The radius is 1km unless an operator overrides it. A garbage or non-positive
// override must fall back to the default rather than disabling the broadcast
// (0) or blasting an unbounded area (negative parsed as a huge radius).
func TestNearbyShopRadiusKm(t *testing.T) {
	tests := []struct {
		name string
		env  string
		want float64
	}{
		{"unset uses the 1km default", "", 1.0},
		{"blank uses the default", "   ", 1.0},
		{"valid override is honoured", "2.5", 2.5},
		{"integer override is honoured", "3", 3.0},
		{"zero falls back", "0", 1.0},
		{"negative falls back", "-5", 1.0},
		{"garbage falls back", "abc", 1.0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("NEARBY_SHOP_RADIUS_KM", tt.env)
			if tt.env == "" {
				os.Unsetenv("NEARBY_SHOP_RADIUS_KM")
			}
			if got := nearbyShopRadiusKm(); got != tt.want {
				t.Fatalf("nearbyShopRadiusKm() = %v, want %v", got, tt.want)
			}
		})
	}
}

// The address line is what the customer reads in the push body, so blank
// columns must not produce ", ," runs or a leading/trailing comma.
func TestNearbyShopAddress(t *testing.T) {
	tests := []struct {
		name string
		shop nearbyShop
		want string
	}{
		{
			"all parts present",
			nearbyShop{AddressLine1: "12 MG Road", AddressLine2: "Near Metro", City: "Jaipur", Pincode: "302001"},
			"12 MG Road, Near Metro, Jaipur, 302001",
		},
		{
			"blank middle part is skipped",
			nearbyShop{AddressLine1: "12 MG Road", AddressLine2: "  ", City: "Jaipur", Pincode: "302001"},
			"12 MG Road, Jaipur, 302001",
		},
		{"only city", nearbyShop{City: "Jaipur"}, "Jaipur"},
		{"nothing at all", nearbyShop{}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := nearbyShopAddress(tt.shop); got != tt.want {
				t.Fatalf("nearbyShopAddress() = %q, want %q", got, tt.want)
			}
		})
	}
}

// A shop with no address on file must still produce a readable body rather
// than a push that opens with ". Tap to explore."
func TestNearbyShopBodyNeverStartsWithPunctuation(t *testing.T) {
	withAddress := nearbyShopBody(nearbyShop{AddressLine1: "12 MG Road", City: "Jaipur"})
	if withAddress != "12 MG Road, Jaipur. Tap to explore." {
		t.Errorf("body with address = %q", withAddress)
	}

	withoutAddress := nearbyShopBody(nearbyShop{})
	if withoutAddress != "A new shop just opened near you. Tap to explore." {
		t.Errorf("body without address = %q", withoutAddress)
	}
}

// The shop image travels as data["image_url"], which fcm_service promotes onto
// the platform notification blocks. Relative upload paths must be absolutised
// or the image silently fails to render.
func TestNearbyShopImageURLResolution(t *testing.T) {
	t.Setenv("NOTIFICATION_PUBLIC_BASE_URL", "https://cdn.example.com")

	if got := resolvePublicNotificationImageURL("uploads/shops/a.jpg"); got != "https://cdn.example.com/uploads/shops/a.jpg" {
		t.Errorf("relative path = %q", got)
	}
	if got := resolvePublicNotificationImageURL("https://x.com/a.jpg"); got != "https://x.com/a.jpg" {
		t.Errorf("absolute path should pass through, got %q", got)
	}
	if got := resolvePublicNotificationImageURL(""); got != "" {
		t.Errorf("empty should stay empty, got %q", got)
	}
}

// A nil notifier / nil dependencies must be inert rather than panicking: the
// callers invoke this fire-and-forget from the onboarding path.
func TestNotifyNewShopIsInertWhenUnconfigured(t *testing.T) {
	var nilNotifier *NearbyShopNotifier
	nilNotifier.NotifyNewShop(t.Context(), "shop1") // must not panic

	NewNearbyShopNotifier(nil, nil).NotifyNewShop(t.Context(), "shop1")
	NewNearbyShopNotifier(nil, nil).NotifyNewShop(t.Context(), "")
}
