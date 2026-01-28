package domain

import "time"

// MobileUser represents a user in the mobile authentication system
type MobileUser struct {
	ID        int64      `json:"id" gorm:"primaryKey"`
	Phone     string     `json:"phone" gorm:"uniqueIndex;not null;size:20"`
	FirstName string     `json:"first_name" gorm:"size:100"`
	LastName  string     `json:"last_name" gorm:"size:100"`
	Email     string     `json:"email" gorm:"index;size:255"`
	IsActive  bool       `json:"is_active" gorm:"default:true"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

// OTPRequest represents an OTP request for mobile authentication
type OTPRequest struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	Phone       string    `json:"phone" gorm:"index;not null;size:20"`
	OTPHash     string    `json:"otp_hash" gorm:"not null;size:255"` // bcrypt hash of OTP
	ExpiresAt   time.Time `json:"expires_at" gorm:"index;not null"`
	Attempts    int       `json:"attempts" gorm:"default:0"`
	MaxAttempts int       `json:"max_attempts" gorm:"default:3"`
	Status      string    `json:"status" gorm:"index;size:50;default:'active'"` // 'active', 'verified', 'expired', 'blocked'
	IPAddress   string    `json:"ip_address" gorm:"size:50"`
	UserAgent   string    `json:"user_agent" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at" gorm:"autoCreateTime;index:,sort:desc"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

// LoginAuditLog represents audit logs for compliance and security monitoring
type LoginAuditLog struct {
	ID        int64     `json:"id" gorm:"primaryKey"`
	Phone     string    `json:"phone" gorm:"index;not null;size:20"`
	Event     string    `json:"event" gorm:"index;size:50"` // 'OTP_REQUESTED', 'OTP_SENT', 'OTP_VERIFIED', 'OTP_FAILED', 'OTP_EXPIRED', 'LOGIN_SUCCESS'
	IPAddress string    `json:"ip_address" gorm:"size:50"`
	UserAgent string    `json:"user_agent" gorm:"type:text"`
	Details   string    `json:"details" gorm:"type:jsonb"`
	CreatedAt time.Time `json:"created_at" gorm:"autoCreateTime;index:,sort:desc"`
}

// OTP rate limit constants
const (
	OTPMaxAttemptsPerPhone = 3
	OTPValidityDuration    = 5 * time.Minute
	OTPMaxRequestsPerHour  = 3
	OTPCooldownSeconds     = 60
	OTPLength              = 6
	IndianPhoneMinLength   = 10
	IndianPhoneMaxLength   = 10
)

// OTP Status constants
const (
	OTPStatusActive   = "active"
	OTPStatusVerified = "verified"
	OTPStatusExpired  = "expired"
	OTPStatusBlocked  = "blocked"
)

// Audit event constants
const (
	AuditEventOTPRequested = "OTP_REQUESTED"
	AuditEventOTPSent      = "OTP_SENT"
	AuditEventOTPVerified  = "OTP_VERIFIED"
	AuditEventOTPFailed    = "OTP_FAILED"
	AuditEventOTPExpired   = "OTP_EXPIRED"
	AuditEventLoginSuccess = "LOGIN_SUCCESS"
)
