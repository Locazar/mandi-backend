package utils

import (
	"strings"
	"testing"
)

func TestGeneratePrefixedUUID(t *testing.T) {
	id := GeneratePrefixedUUID("AD")
	if !strings.HasPrefix(id, "AD-") {
		t.Fatalf("expected id to start with AD-, got %s", id)
	}
	if len(id) <= len("AD-") {
		t.Fatalf("expected id length greater than %d, got %d", len("AD-"), len(id))
	}
}

func TestGenerateAdminID(t *testing.T) {
	id := GenerateAdminID()
	if !strings.HasPrefix(id, "AD-") {
		t.Fatalf("expected admin id to start with AD-, got %s", id)
	}
}

func TestGenerateShopID(t *testing.T) {
	id := GenerateShopID()
	if !strings.HasPrefix(id, "SH-") {
		t.Fatalf("expected shop id to start with SH-, got %s", id)
	}
}
