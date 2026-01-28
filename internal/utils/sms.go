package utils

import (
    "context"
    "errors"
    "fmt"
    "net/http"
    "net/url"
    "strings"
)

// SMSClient defines an SMS provider interface
type SMSClient interface {
    SendSMS(ctx context.Context, to string, message string) error
}

// MSG91 client (simple HTTP implementation)
type MSG91Client struct{
    APIKey string
}

func NewMSG91(apiKey string) *MSG91Client { return &MSG91Client{APIKey: apiKey} }

func (m *MSG91Client) SendSMS(ctx context.Context, to, message string) error {
    if m.APIKey == "" {
        return errors.New("msg91 api key missing")
    }
    // lightweight POST (MSG91 expects different params; this is illustrative)
    endpoint := "https://api.msg91.com/api/v5/flow/"
    data := url.Values{}
    data.Set("mobiles", to)
    data.Set("message", message)
    req, _ := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(data.Encode()))
    req.Header.Set("authkey", m.APIKey)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    resp, err := http.DefaultClient.Do(req)
    if err != nil {
        return err
    }
    defer resp.Body.Close()
    if resp.StatusCode >= 400 {
        return fmt.Errorf("msg91 failed: %s", resp.Status)
    }
    return nil
}

// Twilio client (fallback) - minimal implementation
type TwilioClient struct{
    SID string
    AuthToken string
}

func NewTwilio(sid, token string) *TwilioClient { return &TwilioClient{SID: sid, AuthToken: token} }

func (t *TwilioClient) SendSMS(ctx context.Context, to, message string) error {
    if t.SID=="" || t.AuthToken=="" {
        return errors.New("twilio creds missing")
    }
    endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.SID)
    data := url.Values{}
    data.Set("To", to)
    data.Set("From", "" )
    data.Set("Body", message)
    req, _ := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(data.Encode()))
    req.SetBasicAuth(t.SID, t.AuthToken)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
    resp, err := http.DefaultClient.Do(req)
    if err != nil { return err }
    defer resp.Body.Close()
    if resp.StatusCode>=400 { return fmt.Errorf("twilio err: %s", resp.Status) }
    return nil
}
