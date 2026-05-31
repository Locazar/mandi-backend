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
