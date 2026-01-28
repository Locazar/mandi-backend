package otp

import (
    "crypto/rand"
    "fmt"
    "regexp"
    "time"

    "golang.org/x/crypto/bcrypt"
    "github.com/rohit221990/mandi-backend/pkg/domain"
)

type MobileOTPService struct{}

func NewMobileOTPService() *MobileOTPService { return &MobileOTPService{} }

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
    return time.Now().Add(domain.OTPValidityDuration)
}

func (m *MobileOTPService) IsOTPExpired(expiresAt time.Time) bool {
    return time.Now().After(expiresAt)
}

func (m *MobileOTPService) VerifyOTP(otp, otpHash string) error {
    return bcrypt.CompareHashAndPassword([]byte(otpHash), []byte(otp))
}
