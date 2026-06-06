// pkg/domain/consent_test.go
package domain

import (
	"strings"
	"testing"
)

func TestUserConsentHookAndType(t *testing.T) {
	c := &UserConsent{ConsentType: ConsentTypeTerms}
	_ = c.BeforeCreate(nil)
	if !strings.HasPrefix(c.ID, "cnst_") {
		t.Fatalf("consent id = %q", c.ID)
	}
	if !c.ConsentType.IsValid() {
		t.Fatal("terms must be a valid consent type")
	}
}

func TestUserConsentHookOverwrite(t *testing.T) {
	c := &UserConsent{ID: "preset_consent_id", ConsentType: ConsentTypeTerms}
	_ = c.BeforeCreate(nil)
	if c.ID == "preset_consent_id" {
		t.Fatal("expected ID to be overwritten, but it remained unchanged")
	}
	if !strings.HasPrefix(c.ID, "cnst_") {
		t.Fatalf("expected consent id to have cnst_ prefix, got %q", c.ID)
	}
}
