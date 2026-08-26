package otp

import (
	"errors"
	"fmt"

	"github.com/rohit221990/mandi-backend/pkg/config"
)

// ErrNoOtpProvider is returned by the unconfigured OTP provider. Callers treat
// a verification error as a failed verification (fail closed), so an OTP path
// backed by this provider rejects every code rather than accepting every code.
var ErrNoOtpProvider = errors.New("no external OTP provider is configured")

type twilioOtp struct{}

// NewOtpAuth returns the external OTP provider. Twilio Verify was removed and
// nothing replaced it, so this is an unconfigured provider that FAILS CLOSED.
//
// It previously returned (true, nil) from VerifyOtp unconditionally, which made
// every caller's verification a no-op while still reading like a real check —
// callers guarding with `if !valid { return ErrInvalidOtp }` could never reject
// anything. Any flow that needs real OTP verification should follow the pattern
// used by adminUseCase.issueOtpSession / AdminSignUpOtpVerify: generate a code,
// store only its bcrypt hash on the OTP session, send via TwoFactorSMSService,
// and compare with MobileOTPService.VerifyOTP.
func NewOtpAuth(cfg config.Config) OtpAuth {
	return &twilioOtp{}
}

func (c *twilioOtp) SentOtp(phoneNumber string) (string, error) {
	return "", fmt.Errorf("send otp to %s: %w", phoneNumber, ErrNoOtpProvider)
}

func (c *twilioOtp) VerifyOtp(phoneNumber string, code string) (valid bool, err error) {
	return false, ErrNoOtpProvider
}

func (c *twilioOtp) SentOtpEmail(email string) (string, error) {
	return "", fmt.Errorf("send email otp to %s: %w", email, ErrNoOtpProvider)
}

func (c *twilioOtp) VerifyOtpEmail(email string, code string) (valid bool, err error) {
	return false, ErrNoOtpProvider
}
