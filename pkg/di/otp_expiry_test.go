package di

import (
	"testing"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// Proves the OTP_EXPIRY_SECONDS config value actually reaches the OTP service,
// which is the wiring most likely to break silently: a mis-wired provider still
// compiles and still issues OTPs, just with the wrong lifetime.
func TestProvideMobileOTPServiceUsesConfiguredExpiry(t *testing.T) {
	tests := []struct {
		name    string
		seconds int
		want    time.Duration
	}{
		{"configured 30s", 30, 30 * time.Second},
		{"configured 120s", 120, 2 * time.Minute},
		{"unset config falls back to the default", 0, domain.OTPValidityDuration},
		{"negative config falls back", -10, domain.OTPValidityDuration},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := provideMobileOTPService(config.Config{OTPExpirySeconds: tt.seconds})
			if got := svc.ValidityDuration(); got != tt.want {
				t.Fatalf("ValidityDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

// An unset OTP_EXPIRY_SECONDS must still yield 30s, so deployments that never
// set the variable get the intended default rather than a zero-length window.
func TestUnsetConfigYields30Seconds(t *testing.T) {
	svc := provideMobileOTPService(config.Config{}) // zero value: nothing set
	if got := svc.ValidityDuration(); got != 30*time.Second {
		t.Fatalf("unset config gives %v, want 30s", got)
	}
	if config.DefaultOTPExpirySeconds != 30 {
		t.Fatalf("config.DefaultOTPExpirySeconds = %d, want 30", config.DefaultOTPExpirySeconds)
	}
}
