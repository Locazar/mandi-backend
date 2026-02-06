package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoInterface "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/service/otp"
	"github.com/rohit221990/mandi-backend/pkg/service/sms"
	"github.com/rohit221990/mandi-backend/pkg/service/token"
	usecaseInterface "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type mobileAuthUseCase struct {
	mobileAuthRepo repoInterface.MobileAuthRepository
	otpService     *otp.MobileOTPService
	smsService     *sms.TwilioSMSService
	tokenService   token.TokenService
}

// NewMobileAuthUseCase creates a new mobile auth usecase
func NewMobileAuthUseCase(
	mobileAuthRepo repoInterface.MobileAuthRepository,
	otpService *otp.MobileOTPService,
	smsService *sms.TwilioSMSService,
	tokenService token.TokenService,
) usecaseInterface.MobileAuthUseCase {
	return &mobileAuthUseCase{
		mobileAuthRepo: mobileAuthRepo,
		otpService:     otpService,
		smsService:     smsService,
		tokenService:   tokenService,
	}
}

// SendOTP generates an OTP, sends it via SMS, and stores the request
// Compliant with TRAI DLT guidelines for OTP delivery
func (m *mobileAuthUseCase) SendOTP(ctx context.Context, phone, ipAddress, userAgent string) (*response.SendOTPResponse, error) {
	// Validate phone number (10 digits, starts with 6-9)
	if phone == "" || phone == "null" {
		return nil, fmt.Errorf("phone number is required")
	}
	if !m.otpService.ValidateIndianPhoneNumber(phone) {
		auditLog := &domain.LoginAuditLog{
			Phone:     phone,
			Event:     domain.AuditEventOTPRequested,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Details:   `{"status":"failed","reason":"invalid_phone_format"}`,
		}
		m.mobileAuthRepo.CreateAuditLog(ctx, auditLog)
		return nil, fmt.Errorf("invalid phone number format")
	}

	// Generate OTP (6-digit numeric)
	generatedOTP, err := m.otpService.GenerateOTP()
	if err != nil {
		return nil, fmt.Errorf("failed to generate OTP: %v", err)
	}

	// Hash OTP for storage (never store plaintext)
	otpHash, err := m.otpService.HashOTP(generatedOTP)
	if err != nil {
		return nil, fmt.Errorf("failed to hash OTP: %v", err)
	}

	// Create OTP request record
	otpRequest := &domain.OTPRequest{
		Phone:       phone,
		OTPHash:     otpHash,
		ExpiresAt:   m.otpService.CalculateOTPExpiry(),
		Attempts:    0,
		MaxAttempts: domain.OTPMaxAttemptsPerPhone,
		Status:      domain.OTPStatusActive,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
	}

	err = m.mobileAuthRepo.CreateOTPRequest(ctx, otpRequest)
	if err != nil {
		auditLog := &domain.LoginAuditLog{
			Phone:     phone,
			Event:     domain.AuditEventOTPRequested,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Details:   fmt.Sprintf(`{"status":"failed","reason":"db_error","error":"%s"}`, err.Error()),
		}
		m.mobileAuthRepo.CreateAuditLog(ctx, auditLog)
		return nil, fmt.Errorf("failed to create OTP request: %v", err)
	}

	// Send OTP via Twilio SMS (DLT-approved template)
	fmt.Printf("Sending OTP to phone: %s\n", phone)
	err = m.smsService.SendOTPSMS(phone, generatedOTP)
	if err != nil {
		// Mark OTP as failed and log audit event
		otpRequest.Status = domain.OTPStatusExpired
		m.mobileAuthRepo.UpdateOTPRequest(ctx, otpRequest)

		auditLog := &domain.LoginAuditLog{
			Phone:     phone,
			Event:     domain.AuditEventOTPRequested,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Details:   fmt.Sprintf(`{"status":"failed","reason":"sms_send_failed","error":"%s"}`, err.Error()),
		}
		m.mobileAuthRepo.CreateAuditLog(ctx, auditLog)
		return nil, fmt.Errorf("failed to send OTP: %v", err)
	}

	// Log successful OTP send event (audit/compliance)
	auditLog := &domain.LoginAuditLog{
		Phone:     phone,
		Event:     domain.AuditEventOTPSent,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Details:   fmt.Sprintf(`{"otp_request_id":%d,"validity_seconds":%d}`, otpRequest.ID, domain.OTPValidityDuration/time.Second),
	}
	m.mobileAuthRepo.CreateAuditLog(ctx, auditLog)

	// Return success response (without exposing OTP)
	return &response.SendOTPResponse{
		Message:            "OTP sent successfully",
		Phone:              phone,
		OTPValiditySeconds: int(domain.OTPValidityDuration.Seconds()),
		ConsentMessage:     "By proceeding, you consent to receive SMS OTP for authentication. This is as per TRAI DLT guidelines.",
	}, nil
}

// VerifyOTP validates the OTP and issues JWT token if valid
// Handles OTP expiry, attempt limits, and creates user if not exists
func (m *mobileAuthUseCase) VerifyOTP(ctx context.Context, req *request.VerifyOTPRequest, ipAddress, userAgent string) (*response.VerifyOTPResponse, error) {
	// Validate phone number
	if req.Phone == "" || req.Phone == "null" {
		return nil, fmt.Errorf("phone number is required")
	}

	// Validate phone number format
	if !m.otpService.ValidateIndianPhoneNumber(req.Phone) {
		return nil, fmt.Errorf("invalid phone number format")
	}

	// Get latest OTP request for this phone
	otpRequest, err := m.mobileAuthRepo.GetLatestOTPRequest(ctx, req.Phone)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve OTP request: %v", err)
	}

	if otpRequest == nil {
		auditLog := &domain.LoginAuditLog{
			Phone:     req.Phone,
			Event:     domain.AuditEventOTPFailed,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Details:   `{"reason":"no_otp_request_found"}`,
		}
		m.mobileAuthRepo.CreateAuditLog(ctx, auditLog)
		return nil, fmt.Errorf("no OTP request found")
	}

	// Check if OTP is expired
	if m.otpService.IsOTPExpired(otpRequest.ExpiresAt) {
		otpRequest.Status = domain.OTPStatusExpired
		m.mobileAuthRepo.UpdateOTPRequest(ctx, otpRequest)

		auditLog := &domain.LoginAuditLog{
			Phone:     req.Phone,
			Event:     domain.AuditEventOTPExpired,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Details:   `{"reason":"otp_expired"}`,
		}
		m.mobileAuthRepo.CreateAuditLog(ctx, auditLog)
		return nil, fmt.Errorf("OTP has expired")
	}

	// Check if max attempts exceeded
	if otpRequest.Attempts >= otpRequest.MaxAttempts {
		otpRequest.Status = domain.OTPStatusBlocked
		m.mobileAuthRepo.UpdateOTPRequest(ctx, otpRequest)

		auditLog := &domain.LoginAuditLog{
			Phone:     req.Phone,
			Event:     domain.AuditEventOTPFailed,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Details:   `{"reason":"max_attempts_exceeded"}`,
		}
		m.mobileAuthRepo.CreateAuditLog(ctx, auditLog)
		return nil, fmt.Errorf("maximum OTP attempts exceeded")
	}

	// Verify OTP hash
	err = m.otpService.VerifyOTP(req.OTP, otpRequest.OTPHash)
	if err != nil {
		// Increment attempts on failure
		m.mobileAuthRepo.IncrementOTPAttempts(ctx, otpRequest.ID)

		auditLog := &domain.LoginAuditLog{
			Phone:     req.Phone,
			Event:     domain.AuditEventOTPFailed,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Details:   fmt.Sprintf(`{"attempts":%d,"reason":"invalid_otp"}`, otpRequest.Attempts+1),
		}
		m.mobileAuthRepo.CreateAuditLog(ctx, auditLog)
		return nil, fmt.Errorf("invalid OTP")
	}

	// OTP verified! Mark as verified
	otpRequest.Status = domain.OTPStatusVerified
	m.mobileAuthRepo.UpdateOTPRequest(ctx, otpRequest)

	// Get or create user
	user, err := m.mobileAuthRepo.GetUserByPhone(ctx, req.Phone)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve user: %v", err)
	}

	isNewUser := false
	if user == nil {
		// Create new user
		user = &domain.MobileUser{
			Phone:    req.Phone,
			IsActive: true,
		}
		err = m.mobileAuthRepo.CreateUser(ctx, user)
		if err != nil {
			return nil, fmt.Errorf("failed to create user: %v", err)
		}
		isNewUser = true
	}

	// Generate JWT token
	tokenReq := token.GenerateTokenRequest{
		UserID:   uint(user.ID),
		UsedFor:  token.User,
		ExpireAt: time.Now().Add(24 * time.Hour),
	}
	tokenResp, err := m.tokenService.GenerateToken(tokenReq)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %v", err)
	}
	accessToken := tokenResp.TokenString

	// Log successful OTP verification
	auditLog := &domain.LoginAuditLog{
		Phone:     req.Phone,
		Event:     domain.AuditEventOTPVerified,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Details:   fmt.Sprintf(`{"user_id":%d,"is_new_user":%v}`, user.ID, isNewUser),
	}
	m.mobileAuthRepo.CreateAuditLog(ctx, auditLog)

	// Log login success
	successAuditLog := &domain.LoginAuditLog{
		Phone:     req.Phone,
		Event:     domain.AuditEventLoginSuccess,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Details:   fmt.Sprintf(`{"user_id":%d,"token_issued":true}`, user.ID),
	}
	m.mobileAuthRepo.CreateAuditLog(ctx, successAuditLog)

	// Prepare response
	userDetails := response.VerifyOTPUserDetails{
		ID:    uint(user.ID),
		Phone: user.Phone,
		Email: user.Email,
		Name:  user.FirstName + " " + user.LastName,
	}

	return &response.VerifyOTPResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   24 * 3600, // 24 hours
		User:        userDetails,
		IsNewUser:   isNewUser,
		ConsentInfo: "You have successfully authenticated with your mobile number. Your data is secured and processed as per TRAI guidelines.",
	}, nil
}
