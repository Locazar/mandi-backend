package utils

import "testing"

func TestHashRefreshToken(t *testing.T) {
	h1 := HashRefreshToken("abc123")
	h2 := HashRefreshToken("abc123")
	if h1 != h2 {
		t.Fatal("hash must be deterministic")
	}
	if h1 == "abc123" || len(h1) != 64 {
		t.Fatalf("expected 64-hex-char sha256, got %q", h1)
	}
	if HashRefreshToken("abc124") == h1 {
		t.Fatal("different inputs must hash differently")
	}
}
