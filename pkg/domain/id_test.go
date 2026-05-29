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
