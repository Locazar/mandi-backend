package utils

import (
    "testing"
    "time"
)

func TestTokenPair(t *testing.T) {
    a, r, err := GenerateTokenPair(1, "access-secret", "refresh-secret", time.Minute*5, time.Hour*24)
    if err != nil { t.Fatalf("gen tokens: %v", err) }
    if a=="" || r=="" { t.Fatalf("empty tokens") }
    if _, err := ParseToken(a, "access-secret"); err != nil { t.Fatalf("parse access: %v", err) }
    if _, err := ParseToken(r, "refresh-secret"); err != nil { t.Fatalf("parse refresh: %v", err) }
}
