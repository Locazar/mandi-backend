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
	mobileAuthRepo    repoInterface.MobileAuthRepository
	otpService        *otp.MobileOTPService
	smsService        *sms.TwoFactorSMSService
	tokenService      token.TokenService
	skipOTPValidation bool
}

// NewMobileAuthUseCase creates a new mobile auth usecase
func NewMobileAuthUseCase(
	mobileAuthRepo repoInterface.MobileAuthRepository,
	otpService *otp.MobileOTPService,
	smsService *sms.TwoFactorSMSService,
	tokenService token.TokenService,
	skipOTPValidation bool,
) usecaseInterface.MobileAuthUseCase {
	return &mobileAuthUseCase{
		mobileAuthRepo:    mobileAuthRepo,
		otpService:        otpService,
		smsService:        smsService,
		tokenService:      tokenService,
		skipOTPValidation: skipOTPValidation,
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
		Details:   fmt.Sprintf(`{"otp_request_id":%s,"validity_seconds":%d}`, otpRequest.ID, domain.OTPValidityDuration/time.Second),
	}
	m.mobileAuthRepo.CreateAuditLog(ctx, auditLog)

	// Return success response (without exposing OTP)
	// SessionID is the OTP request ID; the client sends it back on verify.
	return &response.SendOTPResponse{
		Message:            "OTP sent successfully",
		SessionID:          otpRequest.ID,
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

	// Session ID is required to identify which OTP request to verify against
	if req.SessionID == "" {
		return nil, fmt.Errorf("session id is required")
	}

	// Get OTP request by session ID (the ID returned during send)
	otpRequest, err := m.mobileAuthRepo.GetOTPRequestByID(ctx, req.SessionID)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve OTP request: %v", err)
	}

	// Ensure the session belongs to the phone number being verified
	if otpRequest != nil && otpRequest.Phone != req.Phone {
		auditLog := &domain.LoginAuditLog{
			Phone:     req.Phone,
			Event:     domain.AuditEventOTPFailed,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Details:   `{"reason":"session_phone_mismatch"}`,
		}
		m.mobileAuthRepo.CreateAuditLog(ctx, auditLog)
		return nil, fmt.Errorf("invalid session for phone number")
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

	// Reject sessions that are no longer active (already verified, expired or blocked)
	if otpRequest.Status != domain.OTPStatusActive {
		auditLog := &domain.LoginAuditLog{
			Phone:     req.Phone,
			Event:     domain.AuditEventOTPFailed,
			IPAddress: ipAddress,
			UserAgent: userAgent,
			Details:   fmt.Sprintf(`{"reason":"session_not_active","status":%q}`, otpRequest.Status),
		}
		m.mobileAuthRepo.CreateAuditLog(ctx, auditLog)
		return nil, fmt.Errorf("OTP session is no longer valid")
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

	// Verify OTP hash (skip when SKIP_OTP_VALIDATION=true)
	if !m.skipOTPValidation {
		err = m.otpService.VerifyOTP(req.OTP, otpRequest.OTPHash)
		if err != nil {
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
		UserID:   user.ID,
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
		Details:   fmt.Sprintf(`{"user_id":%q,"is_new_user":%v}`, user.ID, isNewUser),
	}
	m.mobileAuthRepo.CreateAuditLog(ctx, auditLog)

	// Log login success
	successAuditLog := &domain.LoginAuditLog{
		Phone:     req.Phone,
		Event:     domain.AuditEventLoginSuccess,
		IPAddress: ipAddress,
		UserAgent: userAgent,
		Details:   fmt.Sprintf(`{"user_id":%q,"token_issued":true}`, user.ID),
	}
	m.mobileAuthRepo.CreateAuditLog(ctx, successAuditLog)

	// Prepare response
	userDetails := response.VerifyOTPUserDetails{
		ID:    user.ID,
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
