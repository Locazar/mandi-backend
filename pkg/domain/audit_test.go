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
