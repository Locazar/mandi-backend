package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
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

// The scannable short link uses the QR_REDIRECT_BASE_URL override when set, so
// a scan shows the branded domain instead of the API host — regardless of the
// request host. The override carries its own path prefix.
func TestQRShortURL_UsesBrandedRedirectBase(t *testing.T) {
	h := NewQRCodeHandler(nil, "https://api.locazar.com", "https://locazar.com/r")
	got := h.shortURL(qrTestCtx("api.locazar.com", "https"), "h39nwn8b")
	if want := "https://locazar.com/r/h39nwn8b"; got != want {
		t.Fatalf("shortURL = %q, want %q", got, want)
	}
}

// A bare short domain with no path prefix yields the shortest possible link.
func TestQRShortURL_BareShortDomain(t *testing.T) {
	h := NewQRCodeHandler(nil, "https://api.locazar.com", "https://lzr.in")
	got := h.shortURL(qrTestCtx("api.locazar.com", "https"), "ab12cd")
	if want := "https://lzr.in/ab12cd"; got != want {
		t.Fatalf("shortURL = %q, want %q", got, want)
	}
}

// A trailing slash on the configured base is normalised — never a double slash.
func TestQRShortURL_TrimsTrailingSlash(t *testing.T) {
	h := NewQRCodeHandler(nil, "", "https://locazar.com/r/")
	got := h.shortURL(qrTestCtx("api.locazar.com", "https"), "abc")
	if want := "https://locazar.com/r/abc"; got != want {
		t.Fatalf("shortURL = %q, want %q", got, want)
	}
}

// The regression this guards: with QR_REDIRECT_BASE_URL unset, the short link
// used to fall through to PublicBaseURL and bake api.locazar.com into printed
// QR codes. The compiled-in default must stand in front of that fallback.
func TestQRShortURL_UnconfiguredNeverUsesAPIHost(t *testing.T) {
	h := NewQRCodeHandler(nil, "https://api.locazar.com", "")
	got := h.shortURL(qrTestCtx("api.locazar.com", "https"), "abc")
	if want := defaultQRRedirectBase + "/abc"; got != want {
		t.Fatalf("shortURL = %q, want %q", got, want)
	}
	if strings.Contains(got, "api.locazar.com") {
		t.Fatalf("shortURL = %q leaks the API host into the scannable link", got)
	}
}

// Same guarantee with nothing configured at all: the request host must not be
// able to substitute itself into the scannable link.
func TestQRShortURL_UnconfiguredIgnoresRequestHost(t *testing.T) {
	h := NewQRCodeHandler(nil, "", "")
	got := h.shortURL(qrTestCtx("api.locazar.com", "https"), "abc")
	if want := defaultQRRedirectBase + "/abc"; got != want {
		t.Fatalf("shortURL = %q, want %q", got, want)
	}
}

// Local development still points scans back at this API's own /r/:code resolver
// by setting the override explicitly, since the compiled-in default is remote.
func TestQRShortURL_LocalDevOverride(t *testing.T) {
	h := NewQRCodeHandler(nil, "", "http://localhost:8080/r")
	got := h.shortURL(qrTestCtx("localhost:8080", "http"), "abc")
	if want := "http://localhost:8080/r/abc"; got != want {
		t.Fatalf("shortURL = %q, want %q", got, want)
	}
}

// The compiled-in default must stay consistent with the resolver it forwards
// to: no trailing slash (shortURL adds the separator) and an absolute origin.
func TestDefaultQRRedirectBase_WellFormed(t *testing.T) {
	if defaultQRRedirectBase == "" {
		t.Fatal("defaultQRRedirectBase is empty; the API host would leak into printed QR codes")
	}
	if strings.HasSuffix(defaultQRRedirectBase, "/") {
		t.Fatalf("defaultQRRedirectBase = %q must not end in / (shortURL adds it)", defaultQRRedirectBase)
	}
	if !strings.HasPrefix(defaultQRRedirectBase, "https://") {
		t.Fatalf("defaultQRRedirectBase = %q must be an absolute https origin", defaultQRRedirectBase)
	}
	if strings.Contains(defaultQRRedirectBase, "api.locazar.com") {
		t.Fatalf("defaultQRRedirectBase = %q is the API host, not a branded short domain", defaultQRRedirectBase)
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
