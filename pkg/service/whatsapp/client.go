package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const apiBase = "https://graph.facebook.com/v21.0"

// ErrCodeOutsideWindow is Meta's error code for a free-form message sent to
// someone who has not messaged the business in the last 24 hours. Such
// messages must be sent as an approved template instead.
const ErrCodeOutsideWindow = 131047

// Client sends messages via the Meta WhatsApp Business Cloud API.
type Client struct {
	phoneNumberID string
	accessToken   string
	http          *http.Client
}

func NewClient(phoneNumberID, accessToken string) *Client {
	return &Client{
		phoneNumberID: phoneNumberID,
		accessToken:   accessToken,
		http:          &http.Client{Timeout: 10 * time.Second},
	}
}

// Enabled returns false when the credentials are not configured, so callers
// can skip the send gracefully in environments without WA credentials.
func (c *Client) Enabled() bool {
	return c.phoneNumberID != "" && c.accessToken != ""
}

// APIError is a non-2xx response from the Cloud API.
type APIError struct {
	Status int
	Code   int
	Body   string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("whatsapp API %d (code %d): %s", e.Status, e.Code, e.Body)
}

// IsOutsideWindow reports whether err is Meta's "outside the 24-hour customer
// service window" error.
func IsOutsideWindow(err error) bool {
	apiErr, ok := err.(*APIError)
	return ok && apiErr.Code == ErrCodeOutsideWindow
}

type sendTextReq struct {
	MessagingProduct string   `json:"messaging_product"`
	To               string   `json:"to"`
	Type             string   `json:"type"`
	Text             textBody `json:"text"`
}

type textBody struct {
	PreviewURL bool   `json:"preview_url"`
	Body       string `json:"body"`
}

type sendTemplateReq struct {
	MessagingProduct string       `json:"messaging_product"`
	To               string       `json:"to"`
	Type             string       `json:"type"`
	Template         templateBody `json:"template"`
}

type templateBody struct {
	Name       string              `json:"name"`
	Language   templateLanguage    `json:"language"`
	Components []templateComponent `json:"components,omitempty"`
}

type templateLanguage struct {
	Code string `json:"code"`
}

type templateComponent struct {
	Type       string              `json:"type"`
	Parameters []templateParameter `json:"parameters"`
}

type templateParameter struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type sendResp struct {
	Messages []struct {
		ID string `json:"id"`
	} `json:"messages"`
}

type errResp struct {
	Error struct {
		Code int `json:"code"`
	} `json:"error"`
}

// SendText sends a plain text message to `to` (E.164 without '+').
func (c *Client) SendText(to, body string) error {
	_, err := c.SendTextWithID(to, body)
	return err
}

// SendTextWithID sends a plain text message and returns Meta's message ID,
// which the relay stores so swipe-replies can be routed back.
func (c *Client) SendTextWithID(to, body string) (string, error) {
	return c.post(sendTextReq{
		MessagingProduct: "whatsapp",
		To:               to,
		Type:             "text",
		Text:             textBody{Body: body},
	})
}

// SendTemplate sends an approved template with positional body parameters
// ({{1}}, {{2}}, ...) and returns Meta's message ID. Templates are the only
// way to message someone outside the 24-hour service window.
func (c *Client) SendTemplate(to, name, lang string, bodyParams ...string) (string, error) {
	params := make([]templateParameter, 0, len(bodyParams))
	for _, p := range bodyParams {
		params = append(params, templateParameter{Type: "text", Text: templateSafe(p)})
	}
	tpl := templateBody{Name: name, Language: templateLanguage{Code: lang}}
	if len(params) > 0 {
		tpl.Components = []templateComponent{{Type: "body", Parameters: params}}
	}
	return c.post(sendTemplateReq{
		MessagingProduct: "whatsapp",
		To:               to,
		Type:             "template",
		Template:         tpl,
	})
}

func (c *Client) post(payload any) (string, error) {
	if !c.Enabled() {
		return "", nil
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	url := fmt.Sprintf("%s/%s/messages", apiBase, c.phoneNumberID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 300 {
		var e errResp
		_ = json.Unmarshal(raw, &e)
		return "", &APIError{Status: resp.StatusCode, Code: e.Error.Code, Body: string(raw)}
	}
	var out sendResp
	if err := json.Unmarshal(raw, &out); err == nil && len(out.Messages) > 0 {
		return out.Messages[0].ID, nil
	}
	return "", nil
}

// templateSafe makes a value acceptable as a template parameter: Meta rejects
// newlines, tabs and runs of more than four spaces, and caps length.
func templateSafe(s string) string {
	s = strings.Join(strings.Fields(s), " ")
	const max = 900
	if r := []rune(s); len(r) > max {
		s = string(r[:max]) + "…"
	}
	if s == "" {
		s = "-"
	}
	return s
}
