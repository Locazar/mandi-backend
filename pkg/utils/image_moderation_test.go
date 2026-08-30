package utils

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// tempImage writes a throwaway file for CheckNudity to open.
func tempImage(t *testing.T) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "img.jpg")
	if err := os.WriteFile(path, []byte("not-a-real-jpeg"), 0o600); err != nil {
		t.Fatalf("write temp image: %v", err)
	}
	return path
}

// withModeration restores the process-wide flag after the test, so these cases
// cannot leak state into each other.
func withModeration(t *testing.T, enabled bool) {
	t.Helper()
	prev := ImageModerationEnabled()
	SetImageModerationEnabled(enabled)
	t.Cleanup(func() { SetImageModerationEnabled(prev) })
}

// The bypass: with moderation off, CheckNudity must return "safe" without
// touching Sightengine at all. The credentials are deliberately left empty —
// any outbound call would fail, so a pass here proves no call was made.
func TestCheckNudityBypassedWhenDisabled(t *testing.T) {
	withModeration(t, false)
	SetSightengineCredentials("", "")

	isAdult, err := CheckNudity(tempImage(t))

	if err != nil {
		t.Fatalf("bypass must not error, got %v", err)
	}
	if isAdult {
		t.Fatal("bypass must report the image as safe")
	}
}

// The bypass short-circuits before the file is even opened, so a missing file
// is not an error either — nothing about the image is inspected.
func TestCheckNudityBypassIgnoresMissingFile(t *testing.T) {
	withModeration(t, false)

	isAdult, err := CheckNudity("/nonexistent/path/does-not-exist.jpg")

	if err != nil {
		t.Fatalf("bypass must not error on a missing file, got %v", err)
	}
	if isAdult {
		t.Fatal("bypass must report safe")
	}
}

// Default is ON, so an existing deployment that sets nothing keeps moderating.
func TestImageModerationDefaultsToEnabled(t *testing.T) {
	// The package-level default, before any startup wiring runs.
	if !imageModerationEnabled {
		t.Fatal("image moderation must default to enabled")
	}
}

// The toggle round-trips, since startup wiring and tests both drive it.
func TestSetImageModerationEnabledRoundTrips(t *testing.T) {
	withModeration(t, true)

	SetImageModerationEnabled(false)
	if ImageModerationEnabled() {
		t.Fatal("expected disabled")
	}
	SetImageModerationEnabled(true)
	if !ImageModerationEnabled() {
		t.Fatal("expected enabled")
	}
}

// A service-side failure must be classified as ErrModerationUnavailable, which
// is what lets handleUpload fail open on it while still blocking a real policy
// verdict. Empty credentials make Sightengine reject the request.
func TestServiceFailureIsClassifiedUnavailable(t *testing.T) {
	// This is the one case that must exercise the real request path, so it
	// makes an outbound call. Skipped under -short so the usual test run does
	// not depend on a third party being reachable.
	if testing.Short() {
		t.Skip("makes a real Sightengine request; skipped in short mode")
	}
	withModeration(t, true)
	SetSightengineCredentials("", "")

	_, err := CheckNudity(tempImage(t))
	if err == nil {
		t.Skip("Sightengine unexpectedly accepted an unauthenticated request; nothing to classify")
	}
	if !errors.Is(err, ErrModerationUnavailable) {
		t.Fatalf("a service-side failure must wrap ErrModerationUnavailable, got %v", err)
	}
}

// Credentials come from config now, not from literals in the source.
func TestSetSightengineCredentials(t *testing.T) {
	prevUser, prevSecret := sightengineAPIUser, sightengineAPISecret
	t.Cleanup(func() { SetSightengineCredentials(prevUser, prevSecret) })

	SetSightengineCredentials("user123", "secret456")
	if sightengineAPIUser != "user123" || sightengineAPISecret != "secret456" {
		t.Fatalf("credentials not stored: user=%q secret=%q", sightengineAPIUser, sightengineAPISecret)
	}
}
