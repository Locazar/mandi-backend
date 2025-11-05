package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// //go:generate mockgen -destination=../../mock/mockrepo/auth_mock.go -package=mockrepo . AuthRepository
type AuthRepository interface {
	SaveRefreshSession(ctx context.Context, refreshSession domain.RefreshSession) error
	FindRefreshSessionByTokenID(ctx context.Context, tokenID string) (domain.RefreshSession, error)

	SaveOtpSession(ctx context.Context, otpSession domain.OtpSession) error
	FindOtpSession(ctx context.Context, otpID string) (domain.OtpSession, error)
}
