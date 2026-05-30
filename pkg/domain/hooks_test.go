package domain

import (
	"strings"
	"testing"
)

func TestBeforeCreateAssignsPrefixedID(t *testing.T) {
	u := &User{}
	_ = u.BeforeCreate(nil)
	if !strings.HasPrefix(u.ID, "usr_") {
		t.Fatalf("user id = %q, want usr_ prefix", u.ID)
	}
	existing := &Admin{ID: "adm_keepme"}
	_ = existing.BeforeCreate(nil)
	if existing.ID != "adm_keepme" {
		t.Fatal("BeforeCreate must not overwrite a non-empty ID")
	}
}
