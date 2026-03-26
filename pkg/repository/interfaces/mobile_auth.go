package interfaces

import (
    "context"

    "github.com/rohit221990/mandi-backend/pkg/domain"
)

// MobileAuthRepository defines persistence operations required by mobile auth usecase
type MobileAuthRepository interface {
    CreateUser(ctx context.Context, user *domain.MobileUser) error
    GetUserByPhone(ctx context.Context, phone string) (*domain.MobileUser, error)
    UpdateUser(ctx context.Context, user *domain.MobileUser) error

    CreateOTPRequest(ctx context.Context, otpReq *domain.OTPRequest) error
    GetLatestOTPRequest(ctx context.Context, phone string) (*domain.OTPRequest, error)
    GetOTPRequestByID(ctx context.Context, id int64) (*domain.OTPRequest, error)
    UpdateOTPRequest(ctx context.Context, otpReq *domain.OTPRequest) error
    IncrementOTPAttempts(ctx context.Context, id int64) error
    CountOTPRequestsInLastHour(ctx context.Context, phone string) (int64, error)
    GetLastOTPRequestTime(ctx context.Context, phone string) (int64, error)

    CreateAuditLog(ctx context.Context, auditLog *domain.LoginAuditLog) error
}
