// Package pincode looks up city/district/state for an Indian PIN code via
// the India Post PIN Code API (https://api.postalpincode.in). It has no
// dependencies beyond a plain HTTP client, so callers construct it directly
// with NewClient() rather than through the wire DI graph.
package pincode

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"regexp"
	"sync"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	applogger "github.com/rohit221990/mandi-backend/pkg/logger"
	"go.uber.org/zap"
)

const externalAPIBaseURL = "https://api.postalpincode.in/pincode/"

// pincodePattern matches a valid 6-digit Indian PIN code (never starts with 0).
var pincodePattern = regexp.MustCompile(`^[1-9][0-9]{5}$`)

// cacheTTL is how long a successful lookup is served from cache before being
// refreshed. Post office boundaries essentially never change, so this is set
// long — it both spares the flaky upstream from repeat traffic (the same
// handful of shop PIN codes get looked up over and over from admin-portal)
// and makes the common case (already-seen PIN) immune to upstream outages.
const cacheTTL = 24 * time.Hour

// overallBudget bounds the total wall-clock time Lookup will spend retrying
// a single upstream call, independent of per-attempt timeouts, so a caller
// (an HTTP handler) always gets a bounded response time.
const overallBudget = 12 * time.Second

// maxAttempts bounds retries for transient upstream failures (network
// errors, 5xx) — api.postalpincode.in is a free, occasionally flaky service.
const maxAttempts = 4

// postOffice mirrors one entry of the India Post API's PostOffice array.
type postOffice struct {
	Name       string `json:"Name"`
	District   string `json:"District"`
	State      string `json:"State"`
	BranchType string `json:"BranchType"`
}

// apiResult mirrors one element of the India Post API's top-level array
// response (the API always wraps its response in a single-element array).
type apiResult struct {
	Message    string       `json:"Message"`
	Status     string       `json:"Status"`
	PostOffice []postOffice `json:"PostOffice"`
}

// Details is the city/district/state resolved for a PIN code — the only
// shape this package exposes to callers.
type Details struct {
	City     string `json:"city"`
	District string `json:"district"`
	State    string `json:"state"`
}

type cacheEntry struct {
	details   Details
	fetchedAt time.Time
}

// retryableError wraps an error that's worth retrying (network failure or a
// 5xx/429 upstream response) so the retry loop can distinguish it from a
// deterministic failure (bad JSON shape, unexpected 4xx) that retrying can
// never fix.
type retryableError struct{ err error }

func (r retryableError) Error() string { return r.err.Error() }
func (r retryableError) Unwrap() error { return r.err }

// Client looks up PIN code details from the India Post API, with an
// in-process cache and retrying HTTP transport in front of it.
type Client struct {
	httpClient *http.Client

	mu    sync.RWMutex
	cache map[string]cacheEntry
}

// NewClient builds a Client with a bounded timeout so a slow/unresponsive
// upstream can never hang a request indefinitely. api.postalpincode.in is a
// free, occasionally flaky service, so the transport is tuned defensively:
//   - HTTP/1.1 forced (TLSNextProto disabled) — it intermittently resets
//     HTTP/2 streams ("stream error ... CANCEL").
//   - Keep-alives disabled — it also occasionally closes/resets idle
//     persistent connections out from under a reused connection ("read:
//     connection reset by peer"); a fresh connection per attempt avoids that.
func NewClient() *Client {
	transport := &http.Transport{
		TLSNextProto:      make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
		DisableKeepAlives: true,
	}
	return &Client{
		httpClient: &http.Client{Timeout: 5 * time.Second, Transport: transport},
		cache:      make(map[string]cacheEntry),
	}
}

// Lookup resolves city/district/state for pincode. It returns a *domain.AppError
// for every failure mode (invalid format, upstream failure, no matching PIN
// code) so handlers can surface it via response.ErrorResponse unchanged.
func (c *Client) Lookup(ctx context.Context, pincode string) (*Details, error) {
	if !pincodePattern.MatchString(pincode) {
		return nil, domain.ValidationError("pincode", "must be a 6-digit Indian PIN code")
	}

	if entry, fresh := c.cacheGet(pincode, cacheTTL); fresh {
		return &entry.details, nil
	}

	ctx, cancel := context.WithTimeout(ctx, overallBudget)
	defer cancel()

	results, err := c.fetchWithRetry(ctx, pincode)
	if err != nil {
		// Upstream is down for this request — serve a stale cache entry
		// rather than fail outright, if we have one. Correct-but-stale
		// beats a 502 for data that essentially never changes.
		if entry, ok := c.cacheGetAny(pincode); ok {
			applogger.L().Warn("pincode lookup failed, serving stale cache",
				zap.String("pincode", pincode), zap.Error(err))
			return &entry.details, nil
		}
		applogger.L().Error("pincode lookup failed after retries",
			zap.String("pincode", pincode), zap.Error(err))
		return nil, domain.ExternalServiceError("India Post PIN Code API", "request to upstream failed", err)
	}

	if len(results) == 0 || results[0].Status != "Success" || len(results[0].PostOffice) == 0 {
		applogger.L().Warn("pincode not found", zap.String("pincode", pincode))
		return nil, domain.NotFoundError("PIN code")
	}

	selected := selectPostOffice(results[0].PostOffice)
	details := Details{City: selected.Name, District: selected.District, State: selected.State}
	c.cacheSet(pincode, details)

	applogger.L().Info("pincode lookup succeeded",
		zap.String("pincode", pincode), zap.String("city", details.City), zap.String("state", details.State))

	return &details, nil
}

// fetchWithRetry retries fetch on retryableError with jittered exponential
// backoff (200ms, 400ms, 800ms, ...). Non-retryable errors (malformed
// response, unexpected 4xx) fail fast on the first attempt — retrying a
// deterministic failure just wastes the overall time budget.
func (c *Client) fetchWithRetry(ctx context.Context, pincode string) ([]apiResult, error) {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		results, err := c.fetch(ctx, pincode)
		if err == nil {
			return results, nil
		}
		lastErr = err

		var re retryableError
		if !errors.As(err, &re) {
			return nil, err
		}
		applogger.L().Warn("pincode lookup attempt failed, retrying",
			zap.String("pincode", pincode), zap.Int("attempt", attempt), zap.Error(err))

		if attempt < maxAttempts {
			backoff := time.Duration(1<<uint(attempt-1)) * 200 * time.Millisecond
			jitter := time.Duration(rand.Int63n(int64(backoff) / 2))
			select {
			case <-time.After(backoff + jitter):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}
	}
	return nil, lastErr
}

// fetch performs one HTTP round-trip and decodes the response. Network
// errors and 5xx/429 are wrapped as retryableError; anything else (bad
// JSON, unexpected 4xx) is returned as a plain error so the caller doesn't
// retry a failure that retrying can't fix.
func (c *Client) fetch(ctx context.Context, pincode string) ([]apiResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, externalAPIBaseURL+pincode, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}
	// Go's default "Go-http-client/1.1" User-Agent gets connection-reset by
	// this upstream's bot mitigation — confirmed by reproducing the exact
	// failure with `curl -A "Go-http-client/1.1"` while a normal curl UA
	// succeeds every time. A browser-like UA is required, not optional.
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, retryableError{fmt.Errorf("transport error: %w", err)}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, retryableError{fmt.Errorf("unexpected status code %d", resp.StatusCode)}
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var results []apiResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("invalid response format: %w", err)
	}
	return results, nil
}

func (c *Client) cacheGet(pincode string, ttl time.Duration) (cacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.cache[pincode]
	if !ok || time.Since(entry.fetchedAt) > ttl {
		return cacheEntry{}, false
	}
	return entry, true
}

// cacheGetAny returns a cache entry regardless of age — used only as a
// last-resort fallback when the upstream is unreachable.
func (c *Client) cacheGetAny(pincode string) (cacheEntry, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	entry, ok := c.cache[pincode]
	return entry, ok
}

func (c *Client) cacheSet(pincode string, details Details) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache[pincode] = cacheEntry{details: details, fetchedAt: time.Now()}
}

// selectPostOffice picks the best match when a PIN code covers multiple post
// offices: Head Post Office first, then Sub Post Office, else the first entry.
func selectPostOffice(offices []postOffice) postOffice {
	if len(offices) == 1 {
		return offices[0]
	}
	for _, o := range offices {
		if o.BranchType == "Head Post Office" {
			return o
		}
	}
	for _, o := range offices {
		if o.BranchType == "Sub Post Office" {
			return o
		}
	}
	return offices[0]
}
