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

// Devanagari (Hindi/Marathi/Nepali/Sanskrit script) transliteration tables —
// ported verbatim from shopUrl.ts's transliterateDevanagari. Not a
// linguistically exact transliteration (no schwa deletion, no long/short
// vowel distinction), just enough that a Devanagari name maps to a distinct,
// readable set of Latin words instead of being stripped to nothing by
// nonAlphaNumRun (which only understands a-z0-9).
var devanagariIndependentVowels = map[rune]string{
	'अ': "a", 'आ': "a", 'इ': "i", 'ई': "i", 'उ': "u", 'ऊ': "u",
	'ऋ': "ri", 'ॠ': "ri", 'ऌ': "li", 'ॡ': "li",
	'ए': "e", 'ऐ': "ai", 'ओ': "o", 'औ': "au",
	'ॲ': "a", 'ऑ': "aa", 'ऎ': "e", 'ऒ': "o",
}

var devanagariConsonants = map[string]string{
	"क": "ka", "ख": "kha", "ग": "ga", "घ": "gha", "ङ": "nga",
	"च": "cha", "छ": "chha", "ज": "ja", "झ": "jha", "ञ": "nya",
	"ट": "ta", "ठ": "tha", "ड": "da", "ढ": "dha", "ण": "na",
	"त": "ta", "थ": "tha", "द": "da", "ध": "dha", "न": "na",
	"प": "pa", "फ": "pha", "ब": "ba", "भ": "bha", "म": "ma",
	"य": "ya", "र": "ra", "ल": "la", "व": "va",
	"श": "sha", "ष": "sha", "स": "sa", "ह": "ha", "ळ": "la",
	// Nukta forms — loanword sounds (Urdu/Persian/English origin).
	"क़": "qa", "ख़": "kha", "ग़": "ga", "ज़": "za",
	"ड़": "ra", "ढ़": "rha", "फ़": "fa", "य़": "ya",
}

var devanagariMatras = map[rune]string{
	'ा': "a", 'ि': "i", 'ी': "i", 'ु': "u", 'ू': "u",
	'ृ': "ri", 'ॄ': "ri", 'े': "e", 'ै': "ai", 'ो': "o", 'ौ': "au",
	'ॅ': "a", 'ॉ': "aa", 'ॆ': "e", 'ॊ': "o",
}

var devanagariDigits = map[rune]string{
	'०': "0", '१': "1", '२': "2", '३': "3", '४': "4",
	'५': "5", '६': "6", '७': "7", '८': "8", '९': "9",
}

const (
	devanagariVirama       = '्'
	devanagariNukta        = '़'
	devanagariAnusvara     = 'ं'
	devanagariChandrabindu = 'ँ'
	devanagariVisarga      = 'ः'
	devanagariAvagraha     = 'ऽ'
)

var devanagariDanda = regexp.MustCompile(`[।॥]`)

// transliterateDevanagari mirrors shopUrl.ts's function of the same name —
// keep the two in exact sync.
func transliterateDevanagari(value string) string {
	runesIn := []rune(value)
	var out strings.Builder
	pendingConsonant := false // last appended syllable still carries its inherent "a"

	dropLastRune := func() {
		s := out.String()
		if s == "" {
			return
		}
		r := []rune(s)
		out.Reset()
		out.WriteString(string(r[:len(r)-1]))
	}

	for i := 0; i < len(runesIn); i++ {
		ch := runesIn[i]
		var twoChar string
		if i+1 < len(runesIn) {
			twoChar = string(runesIn[i : i+2])
		}

		if twoChar != "" {
			if latin, ok := devanagariConsonants[twoChar]; ok {
				out.WriteString(latin)
				pendingConsonant = true
				i++ // also consumed the nukta mark
				continue
			}
		}
		if latin, ok := devanagariConsonants[string(ch)]; ok {
			out.WriteString(latin)
			pendingConsonant = true
			continue
		}
		if latin, ok := devanagariMatras[ch]; ok {
			if pendingConsonant {
				dropLastRune() // drop the consonant's inherent "a"
			}
			out.WriteString(latin)
			pendingConsonant = false
			continue
		}
		if ch == devanagariVirama {
			if pendingConsonant {
				dropLastRune() // no vowel follows this consonant at all
			}
			pendingConsonant = false
			continue
		}
		if ch == devanagariNukta {
			continue // already folded into the two-char consonant lookup above; stray mark is a no-op
		}
		if latin, ok := devanagariIndependentVowels[ch]; ok {
			out.WriteString(latin)
			pendingConsonant = false
			continue
		}
		if latin, ok := devanagariDigits[ch]; ok {
			out.WriteString(latin)
			pendingConsonant = false
			continue
		}
		if ch == devanagariAnusvara {
			out.WriteString("n")
			pendingConsonant = false
			continue
		}
		if ch == devanagariChandrabindu || ch == devanagariAvagraha {
			pendingConsonant = false
			continue // no distinct Latin letter worth keeping
		}
		if ch == devanagariVisarga {
			out.WriteString("h")
			pendingConsonant = false
			continue
		}
		if devanagariDanda.MatchString(string(ch)) {
			out.WriteString(" ")
			pendingConsonant = false
			continue
		}

		// Not Devanagari (Latin, digits, spaces, punctuation, other scripts) — pass through untouched.
		out.WriteRune(ch)
		pendingConsonant = false
	}

	return out.String()
}

// Slugify ports customer-web-app's src/lib/shopUrl.ts `slugify` verbatim:
// transliterate Devanagari, lowercase, NFKD-normalize, collapse any run of
// non a-z0-9 into one hyphen, trim leading/trailing hyphens, cap at 80
// characters. Keep this in exact sync with that file — mandi-backend and the
// customer app must derive the identical slug from the same name+city.
func Slugify(value string) string {
	transliterated := transliterateDevanagari(value)
	lower := strings.ToLower(transliterated)
	stripped := stripAccents(lower)
	hyphenated := nonAlphaNumRun.ReplaceAllString(stripped, "-")
	trimmed := strings.Trim(hyphenated, "-")
	if len(trimmed) > 80 {
		trimmed = trimmed[:80]
	}
	return trimmed
}

// ShopSlug ports shopUrl.ts `shopSlug` verbatim: the shop name alone, with
// the city appended only when the name doesn't already end with it. Returns
// "" when the name can't produce a usable slug (empty, or a script this
// can't render into Latin letters) — callers fall back to the raw id.
func ShopSlug(name, city string) string {
	base := Slugify(name)
	if base == "" {
		return ""
	}
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
// when there's no name to build a slug from, or the name can't produce a
// usable slug — customer-web-app's own resolver accepts a raw id too and
// redirects it to the canonical slug.
func ShopPublicURL(id, name, city string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return ShopSiteOrigin + "/shop/" + id
	}
	slug := ShopSlug(name, city)
	if slug == "" {
		return ShopSiteOrigin + "/shop/" + id
	}
	return ShopSiteOrigin + "/shop/" + slug
}
