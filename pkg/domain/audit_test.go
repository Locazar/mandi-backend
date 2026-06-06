// pkg/domain/audit_test.go
package domain

import (
	"strings"
	"testing"
)

func TestAuditLogHook(t *testing.T) {
	a := &AuditLog{Action: "shop.verify", EntityType: "shop", EntityID: "shp_1"}
	_ = a.BeforeCreate(nil)
	if !strings.HasPrefix(a.ID, "aud_") {
		t.Fatalf("audit id = %q", a.ID)
	}
}

func TestAuditLogHookOverwrite(t *testing.T) {
	a := &AuditLog{ID: "preset_audit_id", Action: "shop.verify", EntityType: "shop", EntityID: "shp_1"}
	_ = a.BeforeCreate(nil)
	if a.ID == "preset_audit_id" {
		t.Fatal("expected ID to be overwritten, but it remained unchanged")
	}
	if !strings.HasPrefix(a.ID, "aud_") {
		t.Fatalf("expected audit id to have aud_ prefix, got %q", a.ID)
	}
}
