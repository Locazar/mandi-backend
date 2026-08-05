package notification

import (
	"errors"
	"testing"
)

// TestIsInvalidTokenError_Strings guards the string-fallback classification so
// the real FCM "NotRegistered" response (and its variants) is treated as a
// permanently-invalid token — the gap that let logged-out sellers fail an
// admin push send.
func TestIsInvalidTokenError_Strings(t *testing.T) {
	invalid := []string{
		"NotRegistered", // the actual FCM multicast per-token error string
		"registration-token-not-registered",
		"UNREGISTERED",
		"the token is not registered",
		"Requested entity was not found",
	}
	for _, m := range invalid {
		if !isInvalidTokenError(errors.New(m)) {
			t.Errorf("isInvalidTokenError(%q) = false, want true", m)
		}
	}

	valid := []string{
		"context deadline exceeded",
		"INTERNAL",
		"quota exceeded",
		"messaging/server-unavailable",
	}
	for _, m := range valid {
		if isInvalidTokenError(errors.New(m)) {
			t.Errorf("isInvalidTokenError(%q) = true, want false (transient/other)", m)
		}
	}

	if isInvalidTokenError(nil) {
		t.Error("isInvalidTokenError(nil) = true, want false")
	}
}
