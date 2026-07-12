// Package pincode looks up city/district/state for an Indian PIN code via
// the India Post PIN Code API (https://api.postalpincode.in). It has no
// dependencies beyond a plain HTTP client, so callers construct it directly
// with NewClient() rather than through the wire DI graph.
package pincode

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	applogger "github.com/rohit221990/mandi-backend/pkg/logger"
	"go.uber.org/zap"
)

const externalAPIBaseURL = "https://api.postalpincode.in/pincode/"

// pincodePattern matches a valid 6-digit Indian PIN code (never starts with 0).
var pincodePattern = regexp.MustCompile(`^[1-9][0-9]{5}$`)

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

// Client looks up PIN code details from the India Post API.
type Client struct {
	httpClient *http.Client
}

// NewClient builds a Client with a bounded timeout so a slow/unresponsive
// upstream can never hang a request indefinitely. Transport forces HTTP/1.1
// — api.postalpincode.in intermittently resets HTTP/2 streams ("stream
// error ... CANCEL"), which surfaces as a spurious request failure; HTTP/1.1
// doesn't hit that failure mode.
func NewClient() *Client {
	transport := &http.Transport{
		TLSNextProto: make(map[string]func(authority string, c *tls.Conn) http.RoundTripper),
	}
	return &Client{httpClient: &http.Client{Timeout: 8 * time.Second, Transport: transport}}
}

// maxAttempts bounds retries for transient upstream failures (network
// errors, 5xx) — api.postalpincode.in is a free, occasionally flaky service.
const maxAttempts = 3

// Lookup resolves city/district/state for pincode. It returns a *domain.AppError
// for every failure mode (invalid format, upstream failure, no matching PIN
// code) so handlers can surface it via response.ErrorResponse unchanged.
func (c *Client) Lookup(ctx context.Context, pincode string) (*Details, error) {
	if !pincodePattern.MatchString(pincode) {
		return nil, domain.ValidationError("pincode", "must be a 6-digit Indian PIN code")
	}

	var results []apiResult
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		results, lastErr = c.fetch(ctx, pincode)
		if lastErr == nil {
			break
		}
		applogger.L().Warn("pincode lookup attempt failed",
			zap.String("pincode", pincode), zap.Int("attempt", attempt), zap.Error(lastErr))
		if attempt < maxAttempts {
			time.Sleep(time.Duration(attempt) * 200 * time.Millisecond)
		}
	}
	if lastErr != nil {
		applogger.L().Error("pincode lookup failed after retries",
			zap.String("pincode", pincode), zap.Int("attempts", maxAttempts), zap.Error(lastErr))
		return nil, domain.ExternalServiceError("India Post PIN Code API", "request to upstream failed", lastErr)
	}

	if len(results) == 0 || results[0].Status != "Success" || len(results[0].PostOffice) == 0 {
		applogger.L().Warn("pincode not found", zap.String("pincode", pincode))
		return nil, domain.NotFoundError("PIN code")
	}

	selected := selectPostOffice(results[0].PostOffice)

	applogger.L().Info("pincode lookup succeeded",
		zap.String("pincode", pincode), zap.String("city", selected.Name), zap.String("state", selected.State))

	return &Details{
		City:     selected.Name,
		District: selected.District,
		State:    selected.State,
	}, nil
}

// fetch performs one HTTP round-trip and decodes the response. Non-2xx
// status and transport errors are both returned as plain errors so Lookup's
// retry loop treats them uniformly.
func (c *Client) fetch(ctx context.Context, pincode string) ([]apiResult, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, externalAPIBaseURL+pincode, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code %d", resp.StatusCode)
	}

	var results []apiResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, fmt.Errorf("invalid response format: %w", err)
	}
	return results, nil
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
