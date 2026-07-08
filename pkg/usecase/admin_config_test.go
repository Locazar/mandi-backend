package usecase

import (
	"strings"
	"testing"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

func TestValidateAppConfigRequiresKey(t *testing.T) {
	cfg := domain.AppConfig{Value: "enabled"}
	err := validateAppConfig(cfg)
	if err == nil {
		t.Fatal("expected validation error for missing config_key")
	}
	if !strings.Contains(err.Error(), "config_key") {
		t.Fatalf("expected config_key validation error, got %v", err)
	}
}

func TestNormalizeAppConfigKey(t *testing.T) {
	key := normalizeAppConfigKey("  Shop Banner  ")
	if key != "shop_banner" {
		t.Fatalf("expected normalized key to be shop_banner, got %q", key)
	}
}
