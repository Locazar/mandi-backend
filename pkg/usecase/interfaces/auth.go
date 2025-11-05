package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/service/token"
)

//go:generate mockgen -destination=../../mock/mockusecase/auth_mock.go -package=mockusecase . AuthUseCase
type AuthUseCase interface {
	//user
	UserSignUp(ctx context.Context, signUpDetails domain.User) (otpID string, err error)
	SingUpOtpVerify(ctx context.Context, otpVerifyDetails request.OTPVerify) (userID uint, err error)
	GoogleLogin(ctx context.Context, user domain.User) (userID uint, err error)
	UserLogin(ctx context.Context, loginDetails request.Login) (userID uint, err error)
	UserLoginOtpSend(ctx context.Context, loginDetails request.OTPLogin) (otpID string, err error)
	LoginOtpVerify(ctx context.Context, otpVerifyDetails request.OTPVerify) (userID uint, err error)

	// admin
	AdminLogin(ctx context.Context, loginDetails request.Login) (adminID uint, err error)
	// token
	GenerateAccessToken(ctx context.Context, tokenParams GenerateTokenParams) (tokenString string, err error)
	GenerateRefreshToken(ctx context.Context, tokenParams GenerateTokenParams) (tokenString string, err error)
	VerifyAndGetRefreshTokenSession(ctx context.Context, refreshToken string, usedFor token.UserType) (domain.RefreshSession, error)
}

type GenerateTokenParams struct {
	UserID   uint
	UserType token.UserType
}
