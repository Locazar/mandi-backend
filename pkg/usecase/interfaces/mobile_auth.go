package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
)

// MobileAuthUseCase defines business logic for mobile authentication
type MobileAuthUseCase interface {
	// SendOTP sends an OTP to the given phone number
	SendOTP(ctx context.Context, phone, ipAddress, userAgent string) (*response.SendOTPResponse, error)

	// VerifyOTP verifies the OTP and returns a JWT token if valid
	VerifyOTP(ctx context.Context, req *request.VerifyOTPRequest, ipAddress, userAgent string) (*response.VerifyOTPResponse, error)
}
