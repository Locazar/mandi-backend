package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// //go:generate mockgen -destination=../../mock/mockrepo/auth_mock.go -package=mockrepo . AuthRepository
type AuthRepository interface {
	SaveRefreshSession(ctx context.Context, refreshSession request.RefreshSession) error
	FindRefreshSessionByTokenID(ctx context.Context, tokenID string, userType string) (request.RefreshSession, error)

	SaveOtpSession(ctx context.Context, otpSession domain.OtpSession) error
	FindOtpSession(ctx context.Context, otpID string) (domain.OtpSession, error)

	SaveOtpSessionEmail(ctx context.Context, otpSession domain.OtpSessionEmail) error
	FindOtpSessionEmail(ctx context.Context, otpID string) (domain.OtpSessionEmail, error)
}
