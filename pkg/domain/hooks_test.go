package domain

import (
	"strings"
	"testing"

	"gorm.io/gorm"
)

// TestBeforeCreateAlwaysAssignsFreshID verifies that BeforeCreate unconditionally
// assigns a fresh prefixed ID, even when the struct already carries a value.
// This is the security backstop: the backend always owns ID generation.
func TestBeforeCreateAlwaysAssignsFreshID(t *testing.T) {
	// Empty struct gets a correctly-prefixed ID.
	u := &User{}
	_ = u.BeforeCreate(nil)
	if !strings.HasPrefix(u.ID, "usr_") {
		t.Fatalf("user id = %q, want usr_ prefix", u.ID)
	}

	// Pre-populated struct is overwritten (unconditional policy).
	preset := &Admin{ID: "adm_keepme"}
	_ = preset.BeforeCreate(nil)
	if preset.ID == "adm_keepme" {
		t.Fatal("BeforeCreate must overwrite a pre-set ID (unconditional policy)")
	}
	if !strings.HasPrefix(preset.ID, "adm_") {
		t.Fatalf("admin id after BeforeCreate = %q, want adm_ prefix", preset.ID)
	}
}

// TestUserSubscriptionBeforeCreateAssignsPrefixedID verifies that UserSubscription
// now uses the proper usub_ prefix instead of a raw Unix-nanosecond integer.
func TestUserSubscriptionBeforeCreateAssignsPrefixedID(t *testing.T) {
	sub := &UserSubscription{}
	_ = sub.BeforeCreate(nil)
	if !strings.HasPrefix(sub.ID, "usub_") {
		t.Fatalf("user subscription id = %q, want usub_ prefix", sub.ID)
	}

	// Pre-populated ID is also overwritten (unconditional policy).
	existing := &UserSubscription{ID: "42"}
	_ = existing.BeforeCreate(nil)
	if existing.ID == "42" {
		t.Fatal("BeforeCreate must overwrite a pre-set user subscription ID")
	}
	if !strings.HasPrefix(existing.ID, "usub_") {
		t.Fatalf("user subscription id after BeforeCreate = %q, want usub_ prefix", existing.ID)
	}
}

type jobEntity interface {
	BeforeCreate(*gorm.DB) error
}

// TestJobBeforeCreateUseDedicatedPrefixes verifies that all Job-domain hooks
// use their own dedicated prefixes rather than borrowing from other entities.
func TestJobBeforeCreateUseDedicatedPrefixes(t *testing.T) {
	type testCase struct {
		name string
		run  func() string
		want string
	}
	cases := []testCase{
		{"Job", func() string { v := &Job{}; _ = v.BeforeCreate(nil); return v.ID }, "job_"},
		{"JobCategory", func() string { v := &JobCategory{}; _ = v.BeforeCreate(nil); return v.ID }, "jcat_"},
		{"JobSubCategory", func() string { v := &JobSubCategory{}; _ = v.BeforeCreate(nil); return v.ID }, "jscat_"},
		{"JobLocation", func() string { v := &JobLocation{}; _ = v.BeforeCreate(nil); return v.ID }, "jloc_"},
		{"JobFilter", func() string { v := &JobFilter{}; _ = v.BeforeCreate(nil); return v.ID }, "jflt_"},
		{"JobCategoryFilter", func() string { v := &JobCategoryFilter{}; _ = v.BeforeCreate(nil); return v.ID }, "jcflt_"},
		{"JobCategoryLocation", func() string { v := &JobCategoryLocation{}; _ = v.BeforeCreate(nil); return v.ID }, "jcloc_"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			id := tc.run()
			if !strings.HasPrefix(id, tc.want) {
				t.Fatalf("%s id = %q, want %s prefix", tc.name, id, tc.want)
			}
		})
	}
}
