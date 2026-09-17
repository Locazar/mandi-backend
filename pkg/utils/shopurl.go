package utils

import (
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

// ShopSiteOrigin is the customer-facing site a shop's public link lives on.
const ShopSiteOrigin = "https://locazar.in"

var nonAlphaNumRun = regexp.MustCompile(`[^a-z0-9]+`)

// stripAccents mirrors JS's `.normalize("NFKD")` immediately followed by the
// `[^a-z0-9]+` strip below: decompose accented runes (é → e + combining
// mark), then let the regex drop the now-isolated combining marks along with
// every other non a-z0-9 rune.
func stripAccents(s string) string {
	t := transform.Chain(norm.NFKD, runes.Remove(runes.In(unicode.Mn)))
	out, _, err := transform.String(t, s)
	if err != nil {
		return s
	}
	return out
}

// Slugify ports customer-web-app's src/lib/shopUrl.ts `slugify` verbatim:
// lowercase, NFKD-normalize, collapse any run of non a-z0-9 into one hyphen,
// trim leading/trailing hyphens, cap at 80 characters. Keep this in exact
// sync with that file — mandi-backend and the customer app must derive the
// identical slug from the same name+city.
func Slugify(value string) string {
	lower := strings.ToLower(value)
	stripped := stripAccents(lower)
	hyphenated := nonAlphaNumRun.ReplaceAllString(stripped, "-")
	trimmed := strings.Trim(hyphenated, "-")
	if len(trimmed) > 80 {
		trimmed = trimmed[:80]
	}
	return trimmed
}

// ShopSlug ports shopUrl.ts `shopSlug` verbatim: the shop name alone, with
// the city appended only when the name doesn't already end with it.
func ShopSlug(name, city string) string {
	base := Slugify(name)
	if city == "" {
		return base
	}
	suffix := Slugify(city)
	if suffix == "" || strings.HasSuffix(base, suffix) {
		return base
	}
	return base + "-" + suffix
}

// ShopPublicURL is the full customer-facing link for a shop, e.g.
// "https://locazar.in/shop/fashion-foru-banglore". Falls back to the raw id
// when there's no name to build a slug from — customer-web-app's own
// resolver accepts a raw id too and redirects it to the canonical slug.
func ShopPublicURL(id, name, city string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ShopSiteOrigin + "/shop/" + id
	}
	return ShopSiteOrigin + "/shop/" + ShopSlug(name, city)
}
