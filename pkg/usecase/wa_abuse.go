package usecase

import (
	"strings"
	"unicode"
)

// Abuse filter for the WhatsApp relay. Messages that contain abusive words in
// English, Hindi (Devanagari) or Hinglish are not forwarded; the sender gets a
// warning instead. The lists are deliberately conservative to avoid blocking
// normal words (e.g. "kutta" = dog, "bc" = because, "saala" = brother-in-law).
// Extra words can be added without a code change via
// WA_RELAY_EXTRA_BLOCKED_WORDS=word1,word2 in .env.

const waMsgAbuseWarning = "⚠️ Your message was not delivered because it contains abusive language. " +
	"Please keep the conversation respectful. – Team Locazar"

// waAbuseRoots match anywhere inside a word, so they also catch longer forms
// ("fucking", "bhenchodd", "chutiyapa", "चूतियापा").
var waAbuseRoots = []string{
	// English
	"fuck", "fuk", "fck", "fuq", "motherfuck", "bitch", "btch", "bastard", "asshole", "whore", "slut", "dickhead",
	// Hinglish (romanised Hindi)
	"madarchod", "maderchod", "behenchod", "bhenchod", "benchod", "bhanchod",
	"chutiy", "chutia", "chootiy", "bhosd", "bhosad", "gandu", "gaandu", "lavde", "lawde",
	"lauda", "lodu", "harami", "haramkhor", "kamine", "kamina", "jhaatu",
	"chinal", "tatti kha",
	// Hindi (Devanagari)
	"मादरचोद", "बहनचोद", "बहेनचोद", "भेनचोद", "चूतिया", "चुतिया", "भोसड़ी", "भोसडी", "भोसड",
	"गांडू", "गाँडू", "लौड़ा", "लौडा", "लवडे", "रंडी", "हरामी", "हरामखोर", "कमीना", "कमीने", "झाटू",
}

// waAbuseWords only match a whole word, because as parts of other words they
// are harmless ("dick" in "Dickens", "lund" in "Lundgren"). "chod" is left out
// on purpose: "chod do" means "leave it". "bc"/"mc" are left out too
// ("because", "MC"); add them via WA_RELAY_EXTRA_BLOCKED_WORDS if wanted.
var waAbuseWords = []string{
	"dick", "cock", "prick", "cunt", "cunts", "randi", "randiya", "twat", "wanker", "dumbass", "jackass",
	"bsdk", "bkl", "mkc", "bkc", "lund", "loda", "lode", "gaand", "chut",
	"लंड", "गांड", "चूत",
}

// waAbuseJoinedRoots are long enough to check against the whole message with
// spaces and punctuation removed, to catch "f u c k" or "m.a.d.a.r.c.h.o.d".
var waAbuseJoinedRoots = []string{
	"fuck", "motherfuck", "madarchod", "behenchod", "bhenchod", "chutiya", "bhosdi",
}

var waLeet = strings.NewReplacer(
	"@", "a", "4", "a", "3", "e", "1", "i", "!", "i", "0", "o", "$", "s", "5", "s", "7", "t", "*", "",
)

// normalizeAbuseToken lowercases, undoes common character swaps and collapses
// runs of the same letter to at most two ("fuuuuck" -> "fuuck", "ass" stays).
func normalizeAbuseToken(s string) string { return collapseRuns(waLeet.Replace(strings.ToLower(s)), 2) }

// collapseRuns keeps at most max consecutive copies of each character.
func collapseRuns(s string, max int) string {
	var b strings.Builder
	var prev rune
	run := 0
	for i, r := range s {
		if i > 0 && r == prev {
			run++
			if run >= max {
				continue
			}
		} else {
			run = 0
		}
		prev = r
		b.WriteRune(r)
	}
	return b.String()
}

// splitAbuseTokens splits on anything that is not a letter, digit, Devanagari
// mark or one of the leet characters normalised above.
func splitAbuseTokens(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Mc, r) {
			return false
		}
		return !strings.ContainsRune("@!$*", r)
	})
}

// containsAbuse reports whether text contains an abusive word. extra holds
// additional whole-word or root entries from configuration.
func containsAbuse(text string, extra []string) bool {
	if strings.TrimSpace(text) == "" {
		return false
	}
	words := make(map[string]bool, len(waAbuseWords)+len(extra))
	for _, w := range waAbuseWords {
		words[w] = true
	}
	roots := append([]string{}, waAbuseRoots...)
	for _, e := range extra {
		if e = normalizeAbuseToken(strings.TrimSpace(e)); e != "" {
			roots = append(roots, e)
		}
	}

	// Multi-word roots (e.g. "tatti kha") are checked on the normalised text.
	normText := normalizeAbuseToken(text)
	var joined strings.Builder
	for _, tok := range splitAbuseTokens(text) {
		n2 := normalizeAbuseToken(tok)
		joined.WriteString(n2)
		// Check both "at most two" and "at most one" repeats, so "fuuuuck"
		// and "chuuutiya" match while "asshole" still does too.
		for _, n := range []string{n2, collapseRuns(n2, 1)} {
			if words[n] {
				return true
			}
			for _, r := range roots {
				if !strings.Contains(r, " ") && strings.Contains(n, r) {
					return true
				}
			}
		}
	}
	for _, r := range roots {
		if strings.Contains(r, " ") && strings.Contains(normText, r) {
			return true
		}
	}
	j := joined.String()
	for _, jv := range []string{j, collapseRuns(j, 1)} {
		for _, r := range waAbuseJoinedRoots {
			if strings.Contains(jv, r) {
				return true
			}
		}
	}
	return false
}
