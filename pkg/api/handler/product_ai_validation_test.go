package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	service "github.com/rohit221990/mandi-backend/pkg/service/ai"
)

// aiStub serves the AI validate-product endpoint with a canned body/status, so
// the tests exercise the real HTTP client path including its error branches.
func aiStub(t *testing.T, status int, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func handlerWithAI(baseURL string) *ProductHandler {
	return &ProductHandler{aiClient: *service.NewClient(baseURL)}
}

// THE FIX: an unreachable AI service must not block the seller. This used to
// return 400 "Failed to validate product image", so one AI outage stopped every
// product upload on the platform.
func TestValidateProductImageAllowsUploadWhenAIServiceIsDown(t *testing.T) {
	// A server that is created then immediately closed gives a connection refused.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	url := srv.URL
	srv.Close()

	if got := handlerWithAI(url).validateProductImage("/tmp/x.jpg", "Vegetables"); got != "" {
		t.Fatalf("an AI outage must not reject the upload; got rejection %q", got)
	}
}

// A 5xx from the AI service is likewise the platform's problem, not the seller's.
func TestValidateProductImageAllowsUploadOnAIServerError(t *testing.T) {
	srv := aiStub(t, http.StatusInternalServerError, `{"success":false,"error":"model unavailable"}`)

	if got := handlerWithAI(srv.URL).validateProductImage("/tmp/x.jpg", "Vegetables"); got != "" {
		t.Fatalf("an AI 500 must not reject the upload; got %q", got)
	}
}

// Same for a response the client cannot parse.
func TestValidateProductImageAllowsUploadOnMalformedAIResponse(t *testing.T) {
	srv := aiStub(t, http.StatusOK, `not json at all`)

	if got := handlerWithAI(srv.URL).validateProductImage("/tmp/x.jpg", "Vegetables"); got != "" {
		t.Fatalf("a malformed AI response must not reject the upload; got %q", got)
	}
}

// success:false is the AI service reporting its own failure — still not the
// seller's fault, so the upload proceeds.
func TestValidateProductImageAllowsUploadWhenAIReportsFailure(t *testing.T) {
	srv := aiStub(t, http.StatusOK, `{"success":false,"error":"could not read image"}`)

	if got := handlerWithAI(srv.URL).validateProductImage("/tmp/x.jpg", "Vegetables"); got != "" {
		t.Fatalf("an AI-side failure must not reject the upload; got %q", got)
	}
}

// PRESERVED BEHAVIOUR: a genuine "wrong category" verdict still rejects, and the
// message still names the category and the reason.
func TestValidateProductImageRejectsGenuineMismatch(t *testing.T) {
	srv := aiStub(t, http.StatusOK,
		`{"success":true,"data":{"valid":false,"confidence":0.92,"reason":"image shows a bicycle"}}`)

	got := handlerWithAI(srv.URL).validateProductImage("/tmp/x.jpg", "Vegetables")
	if got == "" {
		t.Fatal("a confident mismatch verdict must still reject the upload")
	}
	if !strings.Contains(got, "Vegetables") || !strings.Contains(got, "image shows a bicycle") {
		t.Fatalf("rejection should name the category and reason, got %q", got)
	}
}

// A matching image passes.
func TestValidateProductImageAllowsMatchingImage(t *testing.T) {
	srv := aiStub(t, http.StatusOK, `{"success":true,"data":{"valid":true,"confidence":0.99}}`)

	if got := handlerWithAI(srv.URL).validateProductImage("/tmp/x.jpg", "Vegetables"); got != "" {
		t.Fatalf("a valid image must not be rejected; got %q", got)
	}
}

// The confidence threshold is unchanged: a low-confidence mismatch is not
// trusted enough to block a seller.
func TestValidateProductImageIgnoresLowConfidenceMismatch(t *testing.T) {
	srv := aiStub(t, http.StatusOK,
		`{"success":true,"data":{"valid":false,"confidence":0.05,"reason":"unsure"}}`)

	if got := handlerWithAI(srv.URL).validateProductImage("/tmp/x.jpg", "Vegetables"); got != "" {
		t.Fatalf("a low-confidence verdict must not reject the upload; got %q", got)
	}
}

// With no category supplied there is nothing to validate against, and the AI
// service must not be called at all.
func TestValidateProductImageSkipsWhenNoCategory(t *testing.T) {
	called := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	}))
	defer srv.Close()

	for _, category := range []string{"", "   "} {
		if got := handlerWithAI(srv.URL).validateProductImage("/tmp/x.jpg", category); got != "" {
			t.Fatalf("category %q should skip validation, got %q", category, got)
		}
	}
	if called {
		t.Fatal("the AI service must not be called when there is no category")
	}
}
