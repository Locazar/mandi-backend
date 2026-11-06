package whatsapp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

const apiBase = "https://graph.facebook.com/v21.0"

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

// SendText sends a plain text message to `to` (E.164 without '+').
func (c *Client) SendText(to, body string) error {
	if !c.Enabled() {
		return nil
	}
	payload := sendTextReq{
		MessagingProduct: "whatsapp",
		To:               to,
		Type:             "text",
		Text:             textBody{Body: body},
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	url := fmt.Sprintf("%s/%s/messages", apiBase, c.phoneNumberID)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(b))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.accessToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		raw, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("whatsapp API %d: %s", resp.StatusCode, raw)
	}
	return nil
}
