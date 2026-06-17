package domain

import (
	"strings"
	"testing"

	"github.com/nrednav/cuid2"
)

func TestNewIDFormat(t *testing.T) {
	id := NewID(PrefixUser)

	if !strings.HasPrefix(id, "usr_") {
		t.Fatalf("expected id to start with usr_, got %q", id)
	}
	body := strings.TrimPrefix(id, "usr_")
	if !cuid2.IsCuid(body) {
		t.Fatalf("expected body %q to be a valid cuid2", body)
	}
	if len(id) > 32 {
		t.Fatalf("id %q exceeds VARCHAR(32) budget (%d chars)", id, len(id))
	}
}

func TestNewIDUniqueness(t *testing.T) {
	const n = 5000
	seen := make(map[string]struct{}, n)
	for i := 0; i < n; i++ {
		id := NewID(PrefixOrder)
		if _, dup := seen[id]; dup {
			t.Fatalf("collision after %d generations: %q", i, id)
		}
		seen[id] = struct{}{}
	}
}

func TestPrefixHelpers(t *testing.T) {
	id := NewID(PrefixShop)

	if got := PrefixOf(id); got != PrefixShop {
		t.Fatalf("PrefixOf(%q) = %q, want %q", id, got, PrefixShop)
	}
	if !HasPrefix(id, PrefixShop) {
		t.Fatalf("HasPrefix(%q, %q) = false, want true", id, PrefixShop)
	}
	if HasPrefix(id, PrefixUser) {
		t.Fatalf("HasPrefix(%q, %q) = true, want false", id, PrefixUser)
	}
	if PrefixOf("noseparator") != "" {
		t.Fatalf("PrefixOf with no separator should return empty prefix")
	}
}

func TestAllPrefixesUnique(t *testing.T) {
	seen := make(map[IDPrefix]struct{}, len(allPrefixes))
	for _, p := range allPrefixes {
		if p == "" {
			t.Fatal("empty prefix registered")
		}
		if len(p) > 5 {
			t.Fatalf("prefix %q exceeds 5 chars; breaks VARCHAR(32) budget", p)
		}
		if _, dup := seen[p]; dup {
			t.Fatalf("duplicate prefix registered: %q", p)
		}
		seen[p] = struct{}{}
	}
}

// TestNewPrefixStringValues verifies that the new Job entity prefixes and
// PrefixUserSubsc carry the exact string values agreed upon in the design.
// These tests are intentional duplicates of the constant definition: if
// anyone renames the string value the constant holds, this test will fail
// loudly instead of silently changing all generated IDs in production.
func TestNewPrefixStringValues(t *testing.T) {
	cases := []struct {
		name  string
		got   IDPrefix
		want  IDPrefix
	}{
		{"PrefixUserSubsc", PrefixUserSubsc, "usub"},
		{"PrefixJob", PrefixJob, "job"},
		{"PrefixJobCategory", PrefixJobCategory, "jcat"},
		{"PrefixJobSubCategory", PrefixJobSubCategory, "jscat"},
		{"PrefixJobLocation", PrefixJobLocation, "jloc"},
		{"PrefixJobFilter", PrefixJobFilter, "jflt"},
		{"PrefixJobCategoryFilter", PrefixJobCategoryFilter, "jcflt"},
		{"PrefixJobCategoryLocation", PrefixJobCategoryLocation, "jcloc"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.got != tc.want {
				t.Fatalf("%s = %q, want %q", tc.name, tc.got, tc.want)
			}
		})
	}
}

// TestNewPrefixesRegisteredInAllPrefixes verifies that PrefixUserSubsc and all
// Job-domain prefixes appear in allPrefixes so they are covered by the
// uniqueness and length guard in TestAllPrefixesUnique.
func TestNewPrefixesRegisteredInAllPrefixes(t *testing.T) {
	required := []IDPrefix{
		PrefixUserSubsc,
		PrefixJob,
		PrefixJobCategory,
		PrefixJobSubCategory,
		PrefixJobLocation,
		PrefixJobFilter,
		PrefixJobCategoryFilter,
		PrefixJobCategoryLocation,
	}

	registered := make(map[IDPrefix]struct{}, len(allPrefixes))
	for _, p := range allPrefixes {
		registered[p] = struct{}{}
	}

	for _, p := range required {
		if _, ok := registered[p]; !ok {
			t.Errorf("prefix %q missing from allPrefixes registry", p)
		}
	}
}
