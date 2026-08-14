package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// qrTestCtx builds a minimal gin.Context carrying a request with the given host
// and X-Forwarded-Proto, for exercising the QR URL builders.
func qrTestCtx(host, proto string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Host = host
	if proto != "" {
		req.Header.Set("X-Forwarded-Proto", proto)
	}
	c.Request = req
	return c
}

// The scannable short link uses the branded redirect base when configured, so a
// scan shows locazar.com instead of the API host — regardless of request host.
func TestQRShortURL_UsesBrandedRedirectBase(t *testing.T) {
	h := NewQRCodeHandler(nil, "https://api.locazar.com", "https://locazar.com")
	got := h.shortURL(qrTestCtx("api.locazar.com", "https"), "h39nwn8b")
	if want := "https://locazar.com/r/h39nwn8b"; got != want {
		t.Fatalf("shortURL = %q, want %q", got, want)
	}
}

// A trailing slash on the configured base is normalised.
func TestQRShortURL_TrimsTrailingSlash(t *testing.T) {
	h := NewQRCodeHandler(nil, "", "https://locazar.com/")
	got := h.shortURL(qrTestCtx("api.locazar.com", "https"), "abc")
	if want := "https://locazar.com/r/abc"; got != want {
		t.Fatalf("shortURL = %q, want %q", got, want)
	}
}

// Without a branded base, the short link falls back to the API base (publicBaseURL).
func TestQRShortURL_FallsBackToPublicBase(t *testing.T) {
	h := NewQRCodeHandler(nil, "https://api.locazar.com", "")
	got := h.shortURL(qrTestCtx("api.locazar.com", "https"), "abc")
	if want := "https://api.locazar.com/r/abc"; got != want {
		t.Fatalf("shortURL = %q, want %q", got, want)
	}
}

// With neither configured, it derives scheme+host from the request (local/dev).
func TestQRShortURL_FallsBackToRequestHost(t *testing.T) {
	h := NewQRCodeHandler(nil, "", "")
	got := h.shortURL(qrTestCtx("localhost:8080", "http"), "abc")
	if want := "http://localhost:8080/r/abc"; got != want {
		t.Fatalf("shortURL = %q, want %q", got, want)
	}
}

// The QR image URL (and other API links) must stay on the API host even when a
// branded redirect base is set for the scannable link.
func TestQRImageBase_StaysOnApiHost(t *testing.T) {
	h := NewQRCodeHandler(nil, "https://api.locazar.com", "https://locazar.com")
	got := h.base(qrTestCtx("api.locazar.com", "https"))
	if want := "https://api.locazar.com"; got != want {
		t.Fatalf("base = %q, want %q (image URL must not move to the branded domain)", got, want)
	}
}
