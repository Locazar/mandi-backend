package otp

import (
	"fmt"

	"github.com/rohit221990/mandi-backend/pkg/config"
)

type twilioOtp struct{}

// NewOtpAuth returns a stub OTP provider that skips Twilio and accepts any code.
func NewOtpAuth(cfg config.Config) OtpAuth {
	return &twilioOtp{}
}

func (c *twilioOtp) SentOtp(phoneNumber string) (string, error) {
	fmt.Println("[OTP stub] skipping Twilio send for", phoneNumber)
	return "stub-sid", nil
}

func (c *twilioOtp) VerifyOtp(phoneNumber string, code string) (valid bool, err error) {
	return true, nil
}

func (c *twilioOtp) SentOtpEmail(email string) (string, error) {
	return "", fmt.Errorf("not implemented")
}

func (c *twilioOtp) VerifyOtpEmail(email string, code string) (valid bool, err error) {
	return false, fmt.Errorf("not implemented")
}
