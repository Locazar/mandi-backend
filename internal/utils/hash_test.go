package utils

import "testing"

func TestHashCompare(t *testing.T) {
    pass := "s3cr3t"
    h, err := HashPassword(pass)
    if err != nil { t.Fatalf("hash err: %v", err) }
    if err := CompareHashAndPassword(h, pass); err != nil { t.Fatalf("compare failed: %v", err) }
}
