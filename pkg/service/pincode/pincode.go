// Package pincode looks up city/district/state for an Indian PIN code via
// the India Post PIN Code API (https://api.postalpincode.in). It has no
// dependencies beyond a plain HTTP client, so callers construct it directly
// with NewClient() rather than through the wire DI graph.
package pincode

import (
	"context"
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
// upstream can never hang a request indefinitely.
func NewClient() *Client {
	return &Client{httpClient: &http.Client{Timeout: 8 * time.Second}}
}

// Lookup resolves city/district/state for pincode. It returns a *domain.AppError
// for every failure mode (invalid format, upstream failure, no matching PIN
// code) so handlers can surface it via response.ErrorResponse unchanged.
func (c *Client) Lookup(ctx context.Context, pincode string) (*Details, error) {
	if !pincodePattern.MatchString(pincode) {
		return nil, domain.ValidationError("pincode", "must be a 6-digit Indian PIN code")
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, externalAPIBaseURL+pincode, nil)
	if err != nil {
		return nil, domain.InternalError("failed to build pincode lookup request", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		applogger.L().Error("pincode lookup request failed",
			zap.String("pincode", pincode), zap.Error(err))
		return nil, domain.ExternalServiceError("India Post PIN Code API", "request to upstream failed", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		applogger.L().Error("pincode lookup non-200 response",
			zap.String("pincode", pincode), zap.Int("status_code", resp.StatusCode))
		return nil, domain.ExternalServiceError("India Post PIN Code API",
			fmt.Sprintf("unexpected status code %d", resp.StatusCode), nil)
	}

	var results []apiResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		applogger.L().Error("pincode lookup response decode failed",
			zap.String("pincode", pincode), zap.Error(err))
		return nil, domain.ExternalServiceError("India Post PIN Code API", "invalid response format", err)
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
