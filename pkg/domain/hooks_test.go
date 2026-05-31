package domain

import (
	"strconv"
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

func TestUserSubscriptionBeforeCreateAssignsNumericID(t *testing.T) {
	sub := &UserSubscription{}
	_ = sub.BeforeCreate(nil)
	if _, err := strconv.ParseInt(sub.ID, 10, 64); err != nil {
		t.Fatalf("user subscription id = %q, want numeric string", sub.ID)
	}

	existing := &UserSubscription{ID: "42"}
	_ = existing.BeforeCreate(nil)
	if existing.ID != "42" {
		t.Fatal("BeforeCreate must not overwrite a non-empty user subscription ID")
	}
}
