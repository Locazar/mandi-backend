package repository

import "strings"

// searchRankingEnabled is the process-wide toggle for best-effort relevance
// ranking on the customer product-search path. It is set once at startup from
// config (SEARCH_RANKING_ENABLED) and defaults to false so the feature ships
// dark: when off, search behaviour is byte-identical to the previous
// substring-match + created_at/geo ordering.
var searchRankingEnabled bool

// SetSearchRankingEnabled configures the global product-search relevance ranking.
// Called at startup from config; may be toggled by integration tests.
func SetSearchRankingEnabled(enabled bool) { searchRankingEnabled = enabled }

// SearchRankingEnabled reports whether relevance ranking is active.
func SearchRankingEnabled() bool { return searchRankingEnabled }

// Relevance-score tier weights (additive). Tunable constants, not sacred: a
// product's relevance_score is the SUM of every tier it satisfies, so a name
// match that also hits its category naturally outranks a category-only match.
// Higher weight = stronger relevance signal.
const (
	// scoreExactName: product display name equals the full query (no wildcards).
	scoreExactName = 100
	// scoreNamePrefix: product display name starts with the full query phrase.
	scoreNamePrefix = 60
	// scoreNameContains: product display name contains the full query phrase.
	scoreNameContains = 40
	// scoreCategoryContains: category OR sub_category name contains the phrase.
	scoreCategoryContains = 25
	// scoreDepartmentContains: department name contains the phrase.
	scoreDepartmentContains = 15
	// scoreLongTextContains: description / highlights / dynamic_fields contains
	// the phrase. Lowest weight so a stray attribute hit never outranks a name.
	scoreLongTextContains = 5
	// scorePerTokenNameBonus: small bonus per query token found in the product
	// name, so a multi-word query matching several words in the name outranks
	// one that only matches a single word plus a category.
	scorePerTokenNameBonus = 10
)

// tokenizeQuery splits a raw search query into whitespace-delimited tokens with
// empties removed. Used for AND-of-tokens candidacy and per-token scoring.
func tokenizeQuery(keyword string) []string {
	return strings.Fields(keyword)
}
