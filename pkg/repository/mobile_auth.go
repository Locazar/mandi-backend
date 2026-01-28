package repository

import (
	"context"
	"database/sql"
	"errors"
	"log"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoInterface "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
)

type mobileAuthDatabase struct {
	db *sql.DB
}

// NewMobileAuthRepository creates a new instance of mobile auth repository
func NewMobileAuthRepository(db *sql.DB) repoInterface.MobileAuthRepository {
	return &mobileAuthDatabase{db: db}
}

// CreateUser creates a new user
func (m *mobileAuthDatabase) CreateUser(ctx context.Context, user *domain.MobileUser) error {
	query := `
		INSERT INTO users (phone, first_name, last_name, email, is_active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	err := m.db.QueryRowContext(ctx, query, user.Phone, user.FirstName, user.LastName, user.Email, user.IsActive).
		Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		log.Printf("Failed to create user: %v", err)
		return err
	}
	return nil
}

// GetUserByPhone retrieves a user by phone number
func (m *mobileAuthDatabase) GetUserByPhone(ctx context.Context, phone string) (*domain.MobileUser, error) {
	query := `
		SELECT id, phone, first_name, last_name, email, is_active, created_at, updated_at, deleted_at
		FROM users
		WHERE phone = $1 AND deleted_at IS NULL
		LIMIT 1
	`
	user := &domain.MobileUser{}
	err := m.db.QueryRowContext(ctx, query, phone).
		Scan(&user.ID, &user.Phone, &user.FirstName, &user.LastName, &user.Email, &user.IsActive, &user.CreatedAt, &user.UpdatedAt, &user.DeletedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // User not found (not an error)
		}
		log.Printf("Failed to get user by phone: %v", err)
		return nil, err
	}
	return user, nil
}

// UpdateUser updates a user
func (m *mobileAuthDatabase) UpdateUser(ctx context.Context, user *domain.MobileUser) error {
	query := `
		UPDATE users
		SET first_name = $1, last_name = $2, email = $3, is_active = $4, updated_at = NOW()
		WHERE id = $5
	`
	_, err := m.db.ExecContext(ctx, query, user.FirstName, user.LastName, user.Email, user.IsActive, user.ID)
	if err != nil {
		log.Printf("Failed to update user: %v", err)
		return err
	}
	return nil
}

// CreateOTPRequest creates a new OTP request
func (m *mobileAuthDatabase) CreateOTPRequest(ctx context.Context, otpReq *domain.OTPRequest) error {
	query := `
		INSERT INTO otp_requests (phone, otp_hash, expires_at, attempts, max_attempts, status, ip_address, user_agent, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW(), NOW())
		RETURNING id, created_at, updated_at
	`
	err := m.db.QueryRowContext(ctx, query,
		otpReq.Phone,
		otpReq.OTPHash,
		otpReq.ExpiresAt,
		otpReq.Attempts,
		otpReq.MaxAttempts,
		otpReq.Status,
		otpReq.IPAddress,
		otpReq.UserAgent,
	).Scan(&otpReq.ID, &otpReq.CreatedAt, &otpReq.UpdatedAt)
	if err != nil {
		log.Printf("Failed to create OTP request: %v", err)
		return err
	}
	return nil
}

// GetLatestOTPRequest retrieves the latest active OTP request for a phone
func (m *mobileAuthDatabase) GetLatestOTPRequest(ctx context.Context, phone string) (*domain.OTPRequest, error) {
	query := `
		SELECT id, phone, otp_hash, expires_at, attempts, max_attempts, status, ip_address, user_agent, created_at, updated_at
		FROM otp_requests
		WHERE phone = $1 AND status = $2
		ORDER BY created_at DESC
		LIMIT 1
	`
	otpReq := &domain.OTPRequest{}
	err := m.db.QueryRowContext(ctx, query, phone, domain.OTPStatusActive).
		Scan(&otpReq.ID, &otpReq.Phone, &otpReq.OTPHash, &otpReq.ExpiresAt, &otpReq.Attempts, &otpReq.MaxAttempts, &otpReq.Status, &otpReq.IPAddress, &otpReq.UserAgent, &otpReq.CreatedAt, &otpReq.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No OTP found
		}
		log.Printf("Failed to get latest OTP request: %v", err)
		return nil, err
	}
	return otpReq, nil
}

// GetOTPRequestByID retrieves an OTP request by ID
func (m *mobileAuthDatabase) GetOTPRequestByID(ctx context.Context, id int64) (*domain.OTPRequest, error) {
	query := `
		SELECT id, phone, otp_hash, expires_at, attempts, max_attempts, status, ip_address, user_agent, created_at, updated_at
		FROM otp_requests
		WHERE id = $1
		LIMIT 1
	`
	otpReq := &domain.OTPRequest{}
	err := m.db.QueryRowContext(ctx, query, id).
		Scan(&otpReq.ID, &otpReq.Phone, &otpReq.OTPHash, &otpReq.ExpiresAt, &otpReq.Attempts, &otpReq.MaxAttempts, &otpReq.Status, &otpReq.IPAddress, &otpReq.UserAgent, &otpReq.CreatedAt, &otpReq.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		log.Printf("Failed to get OTP request by ID: %v", err)
		return nil, err
	}
	return otpReq, nil
}

// UpdateOTPRequest updates an OTP request
func (m *mobileAuthDatabase) UpdateOTPRequest(ctx context.Context, otpReq *domain.OTPRequest) error {
	query := `
		UPDATE otp_requests
		SET status = $1, attempts = $2, updated_at = NOW()
		WHERE id = $3
	`
	_, err := m.db.ExecContext(ctx, query, otpReq.Status, otpReq.Attempts, otpReq.ID)
	if err != nil {
		log.Printf("Failed to update OTP request: %v", err)
		return err
	}
	return nil
}

// IncrementOTPAttempts increments the attempts counter for an OTP request
func (m *mobileAuthDatabase) IncrementOTPAttempts(ctx context.Context, id int64) error {
	query := `
		UPDATE otp_requests
		SET attempts = attempts + 1, updated_at = NOW()
		WHERE id = $1
	`
	_, err := m.db.ExecContext(ctx, query, id)
	if err != nil {
		log.Printf("Failed to increment OTP attempts: %v", err)
		return err
	}
	return nil
}

// CountOTPRequestsInLastHour counts OTP requests from a phone in the last hour
func (m *mobileAuthDatabase) CountOTPRequestsInLastHour(ctx context.Context, phone string) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM otp_requests
		WHERE phone = $1 AND created_at > NOW() - INTERVAL '1 hour'
	`
	var count int64
	err := m.db.QueryRowContext(ctx, query, phone).Scan(&count)
	if err != nil {
		log.Printf("Failed to count OTP requests: %v", err)
		return 0, err
	}
	return count, nil
}

// GetLastOTPRequestTime gets the Unix timestamp of the last OTP request
func (m *mobileAuthDatabase) GetLastOTPRequestTime(ctx context.Context, phone string) (int64, error) {
	query := `
		SELECT EXTRACT(EPOCH FROM created_at)::bigint
		FROM otp_requests
		WHERE phone = $1
		ORDER BY created_at DESC
		LIMIT 1
	`
	var timestamp int64
	err := m.db.QueryRowContext(ctx, query, phone).Scan(&timestamp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil // No requests found
		}
		log.Printf("Failed to get last OTP request time: %v", err)
		return 0, err
	}
	return timestamp, nil
}

// CreateAuditLog creates an audit log entry
func (m *mobileAuthDatabase) CreateAuditLog(ctx context.Context, auditLog *domain.LoginAuditLog) error {
	query := `
		INSERT INTO login_audit_logs (phone, event, ip_address, user_agent, details, created_at)
		VALUES ($1, $2, $3, $4, $5, NOW())
		RETURNING id, created_at
	`
	err := m.db.QueryRowContext(ctx, query,
		auditLog.Phone,
		auditLog.Event,
		auditLog.IPAddress,
		auditLog.UserAgent,
		auditLog.Details,
	).Scan(&auditLog.ID, &auditLog.CreatedAt)
	if err != nil {
		log.Printf("Failed to create audit log: %v", err)
		return err
	}
	return nil
}
