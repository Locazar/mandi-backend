package otp

import (
	"testing"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// The validity window is configurable (OTP_EXPIRY_SECONDS) rather than a
// hard-coded constant, and every OTP flow reads it through this one service.
func TestValidityDurationIsConfigurable(t *testing.T) {
	tests := []struct {
		name  string
		given time.Duration
		want  time.Duration
	}{
		{"configured 30s", 30 * time.Second, 30 * time.Second},
		{"configured 2m", 2 * time.Minute, 2 * time.Minute},
		{"zero falls back to the package default", 0, domain.OTPValidityDuration},
		{"negative falls back rather than issuing already-expired OTPs", -5 * time.Second, domain.OTPValidityDuration},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMobileOTPService(tt.given).ValidityDuration(); got != tt.want {
				t.Fatalf("ValidityDuration() = %v, want %v", got, tt.want)
			}
		})
	}
}

// The default is 30 seconds, matching config.DefaultOTPExpirySeconds. These two
// must stay in step: the config default is what production uses, this constant
// is the fallback when the config value is missing.
func TestDefaultValidityIs30Seconds(t *testing.T) {
	if domain.OTPValidityDuration != 30*time.Second {
		t.Fatalf("default OTP validity = %v, want 30s", domain.OTPValidityDuration)
	}
}

// CalculateOTPExpiry must reflect the configured window, since that stamp is
// what every verify path compares against.
func TestCalculateOTPExpiryUsesConfiguredWindow(t *testing.T) {
	svc := NewMobileOTPService(30 * time.Second)

	before := time.Now()
	expiry := svc.CalculateOTPExpiry()

	delta := expiry.Sub(before)
	if delta < 29*time.Second || delta > 31*time.Second {
		t.Fatalf("expiry is %v from now, want ~30s", delta)
	}
	if svc.IsOTPExpired(expiry) {
		t.Fatal("a freshly issued OTP must not already be expired")
	}
}

// A stamp from a longer-lived config must not be treated as expired by a
// service instance, and one past its window must be.
func TestIsOTPExpiredBoundaries(t *testing.T) {
	svc := NewMobileOTPService(30 * time.Second)

	if svc.IsOTPExpired(time.Now().Add(1 * time.Second)) {
		t.Error("an OTP expiring in 1s is still valid now")
	}
	if !svc.IsOTPExpired(time.Now().Add(-1 * time.Second)) {
		t.Error("an OTP that expired 1s ago must be rejected")
	}
}

// A nil service must not panic — ValidityDuration is called from request paths.
func TestValidityDurationNilSafe(t *testing.T) {
	var svc *MobileOTPService
	if got := svc.ValidityDuration(); got != domain.OTPValidityDuration {
		t.Fatalf("nil service = %v, want the default %v", got, domain.OTPValidityDuration)
	}
}
