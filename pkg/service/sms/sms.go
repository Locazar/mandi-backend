package sms

import (
    "context"
    "errors"
    "fmt"
    "net/http"
    "net/url"
    "strings"
)

// TwilioSMSService is a minimal Twilio SMS sender used by mobile auth usecase.
// It provides a simple SendOTPSMS method that posts to Twilio REST API.
type TwilioSMSService struct{
    AccountSID string
    AuthToken  string
    FromNumber string
    Client     *http.Client
}

// NewTwilioSMSService creates a new TwilioSMSService.
func NewTwilioSMSService(sid, token, from string) *TwilioSMSService {
    return &TwilioSMSService{AccountSID: sid, AuthToken: token, FromNumber: from, Client: http.DefaultClient}
}

// SendOTPSMSWithCtx sends an OTP SMS to the provided phone number using Twilio API with a context.
func (t *TwilioSMSService) SendOTPSMSWithCtx(ctx context.Context, toPhone, otp string) error {
    if t == nil {
        return errors.New("twilio service not initialized")
    }
    if t.AccountSID == "" || t.AuthToken == "" {
        return errors.New("twilio credentials missing")
    }
    endpoint := fmt.Sprintf("https://api.twilio.com/2010-04-01/Accounts/%s/Messages.json", t.AccountSID)
    data := url.Values{}
    data.Set("To", toPhone)
    data.Set("From", t.FromNumber)
    data.Set("Body", fmt.Sprintf("Your OTP is %s. It will expire in 5 minutes.", otp))

    req, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(data.Encode()))
    if err != nil { return err }
    req.SetBasicAuth(t.AccountSID, t.AuthToken)
    req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

    resp, err := t.Client.Do(req)
    if err != nil { return err }
    defer resp.Body.Close()
    if resp.StatusCode >= 400 { return fmt.Errorf("twilio send failed: %s", resp.Status) }
    return nil
}

// SendOTPSMS is a compatibility wrapper without context used by existing code.
func (t *TwilioSMSService) SendOTPSMSNoCtx(toPhone, otp string) error {
    return t.SendOTPSMSWithCtx(context.Background(), toPhone, otp)
}

// SendOTPSMS matches older expected signature (phone, otp string) error
func (t *TwilioSMSService) SendOTPSMS(phone, otp string) error {
    return t.SendOTPSMSNoCtx(phone, otp)
}
