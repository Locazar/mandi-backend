package usecase

import (
	"strings"
	"testing"
)

// TestVerificationTemplateKeyFor asserts that every combination of the four
// verification checks routes to the correct template key: shop photo + shop
// address are the mandatory go-live pair, business + identity docs are trust
// documents that never by themselves make a shop live.
func TestVerificationTemplateKeyFor(t *testing.T) {
	cases := []struct {
		name                    string
		photo, addr, biz, ident bool
		want                    string
	}{
		{"all four verified -> fully verified", true, true, true, true, "shop_fully_verified"},
		{"mandatory pass, no docs -> live partial", true, true, false, false, "shop_live_partial"},
		{"mandatory pass, business only -> live partial", true, true, true, false, "shop_live_partial"},
		{"mandatory pass, identity only -> live partial", true, true, false, true, "shop_live_partial"},
		{"photo only -> pending review", true, false, false, false, "shop_pending_review"},
		{"address only -> pending review", false, true, false, false, "shop_pending_review"},
		{"docs verified but shop pending -> pending review", false, false, true, true, "shop_pending_review"},
		{"nothing verified -> pending review", false, false, false, false, "shop_pending_review"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := verificationTemplateKeyFor(tc.photo, tc.addr, tc.biz, tc.ident)
			if got != tc.want {
				t.Errorf("verificationTemplateKeyFor(%v,%v,%v,%v) = %q, want %q",
					tc.photo, tc.addr, tc.biz, tc.ident, got, tc.want)
			}
		})
	}
}

// TestVerificationTemplateKeyFor_AllCombinationsProduceAKey guards that every
// one of the 16 combinations resolves to one of the three known keys — no
// branch leaves the seller without a notification template to send.
func TestVerificationTemplateKeyFor_AllCombinationsProduceAKey(t *testing.T) {
	known := map[string]bool{
		"shop_fully_verified": true,
		"shop_live_partial":   true,
		"shop_pending_review": true,
	}
	for i := 0; i < 16; i++ {
		key := verificationTemplateKeyFor(i&1 != 0, i&2 != 0, i&4 != 0, i&8 != 0)
		if !known[key] {
			t.Errorf("combination %04b produced unknown key %q", i, key)
		}
	}
}

// TestSubstituteVerificationPlaceholders covers the placeholder substitution
// every verification template body goes through before sending — shop name,
// the four verified/pending status placeholders, and the remark placeholder
// both with and without an actual remark.
func TestSubstituteVerificationPlaceholders(t *testing.T) {
	t.Run("shop_name substitutes", func(t *testing.T) {
		got := substituteVerificationPlaceholders("Hello {{shop_name}}!", "Raju Furniture", false, false, false, false, "")
		if got != "Hello Raju Furniture!" {
			t.Errorf("got %q", got)
		}
	})

	t.Run("status placeholders reflect each flag independently", func(t *testing.T) {
		got := substituteVerificationPlaceholders(
			"{{photo_status}}/{{address_status}}/{{business_doc_status}}/{{identity_doc_status}}",
			"", true, false, true, false, "",
		)
		want := "verified ✓/pending ✗/verified ✓/pending ✗"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})

	t.Run("remark present is appended with trailing punctuation", func(t *testing.T) {
		got := substituteVerificationPlaceholders("Reason: {{remark}}Please fix.", "", false, false, false, false, "blurry photo")
		if !strings.Contains(got, "blurry photo") {
			t.Errorf("got %q, want it to contain the remark", got)
		}
		if strings.Contains(got, "{{remark}}") {
			t.Errorf("got %q, placeholder was not substituted", got)
		}
	})

	t.Run("remark absent substitutes to nothing, not a placeholder-shaped gap", func(t *testing.T) {
		got := substituteVerificationPlaceholders("Reason: {{remark}}Please fix.", "", false, false, false, false, "")
		if got != "Reason: Please fix." {
			t.Errorf("got %q", got)
		}
	})

	t.Run("unrecognized placeholder-like text is left untouched", func(t *testing.T) {
		got := substituteVerificationPlaceholders("{{not_a_real_placeholder}}", "x", false, false, false, false, "")
		if got != "{{not_a_real_placeholder}}" {
			t.Errorf("got %q", got)
		}
	})
}
