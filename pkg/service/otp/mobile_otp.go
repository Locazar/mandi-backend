package otp

import (
	"crypto/rand"
	"fmt"
	"regexp"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"golang.org/x/crypto/bcrypt"
)

// MobileOTPService generates, hashes, and validates OTPs. The validity window
// is per-instance so it can be configured (OTP_EXPIRY_SECONDS) rather than
// hard-coded, and is shared by every OTP flow that calls CalculateOTPExpiry.
type MobileOTPService struct {
	validity time.Duration
}

// NewMobileOTPService builds the service with the given OTP validity window.
// A non-positive duration falls back to domain.OTPValidityDuration rather than
// producing OTPs that are already expired the moment they are issued.
func NewMobileOTPService(validity time.Duration) *MobileOTPService {
	if validity <= 0 {
		validity = domain.OTPValidityDuration
	}
	return &MobileOTPService{validity: validity}
}

// ValidityDuration reports the configured OTP lifetime, so callers can tell the
// client how long its countdown should run instead of guessing.
func (m *MobileOTPService) ValidityDuration() time.Duration {
	if m == nil || m.validity <= 0 {
		return domain.OTPValidityDuration
	}
	return m.validity
}

func (m *MobileOTPService) ValidateIndianPhoneNumber(phone string) bool {
	// Simple validation: 10 digits starting with 6-9
	re := regexp.MustCompile(`^[6-9]\d{9}$`)
	return re.MatchString(phone)
}

func (m *MobileOTPService) GenerateOTP() (string, error) {
	b := make([]byte, 6)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	// convert bytes to digits
	otp := ""
	for _, by := range b {
		otp += fmt.Sprintf("%d", int(by)%10)
		if len(otp) >= domain.OTPLength {
			break
		}
	}
	if len(otp) < domain.OTPLength {
		otp = fmt.Sprintf("%0*d", domain.OTPLength, time.Now().UnixNano()%1000000)
	}
	return otp[:domain.OTPLength], nil
}

func (m *MobileOTPService) HashOTP(otp string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(otp), bcrypt.DefaultCost)
	return string(b), err
}

func (m *MobileOTPService) CalculateOTPExpiry() time.Time {
	return time.Now().Add(m.ValidityDuration())
}

func (m *MobileOTPService) IsOTPExpired(expiresAt time.Time) bool {
	return time.Now().After(expiresAt)
}

func (m *MobileOTPService) VerifyOTP(otp, otpHash string) error {
	return bcrypt.CompareHashAndPassword([]byte(otpHash), []byte(otp))
}
