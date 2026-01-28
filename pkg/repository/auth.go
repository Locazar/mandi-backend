package repository

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type authDatabase struct {
	DB *gorm.DB
}

func NewAuthRepository(db *gorm.DB) interfaces.AuthRepository {
	return &authDatabase{
		DB: db,
	}
}

func (c *authDatabase) SaveRefreshSession(ctx context.Context, refreshSession request.RefreshSession) error {
	if refreshSession.UserType == "admin" {
		query := `INSERT INTO admin_refresh_sessions (token_id, user_id, refresh_token, expire_at, user_type) 
VALUES ($1, $2, $3, $4, $5)`
		err := c.DB.Exec(query, refreshSession.TokenID, refreshSession.UserID, refreshSession.RefreshToken, refreshSession.ExpireAt, refreshSession.UserType).Error
		return err
	} else {
		query := `INSERT INTO user_refresh_sessions (token_id, user_id, refresh_token, expire_at, user_type) 
VALUES ($1, $2, $3, $4, $5)`
		err := c.DB.Exec(query, refreshSession.TokenID, refreshSession.UserID, refreshSession.RefreshToken, refreshSession.ExpireAt, refreshSession.UserType).Error
		return err
	}
}
func (c *authDatabase) FindRefreshSessionByTokenID(ctx context.Context, tokenID string, userType string) (refreshSession domain.RefreshSession, err error) {
	if userType == "admin" {
		query := `SELECT * FROM admin_refresh_sessions WHERE token_id = $1`
		err = c.DB.Raw(query, tokenID).Scan(&refreshSession).Error
	} else {
		query := `SELECT * FROM user_refresh_sessions WHERE token_id = $1`
		err = c.DB.Raw(query, tokenID).Scan(&refreshSession).Error
	}

	return
}

func (c *authDatabase) SaveOtpSession(ctx context.Context, otpSession domain.OtpSession) error {

	query := `INSERT INTO otp_sessions (otp_id, user_id, admin_id, user_type, phone ,expire_at) 
	VALUES ($1, $2, $3, $4, $5, $6)`
	err := c.DB.Exec(query, otpSession.OtpID, otpSession.UserID, otpSession.AdminID, otpSession.UserType, otpSession.Phone, otpSession.ExpireAt).Error
	return err
}

func (c *authDatabase) FindOtpSession(ctx context.Context, otpID string) (otpSession domain.OtpSession, err error) {

	query := `SELECT * FROM otp_sessions WHERE otp_id = $1`

	err = c.DB.Raw(query, otpID).Scan(&otpSession).Error

	return otpSession, err
}

func (c *authDatabase) SaveOtpSessionEmail(ctx context.Context, otpSession domain.OtpSessionEmail) error {

	query := `INSERT INTO otp_sessions_email (otp_id, user_id, email ,expire_at) 
	VALUES ($1, $2, $3, $4)`
	err := c.DB.Exec(query, otpSession.OtpID, otpSession.UserID, otpSession.Email, otpSession.ExpireAt).Error
	return err

}

func (c *authDatabase) FindOtpSessionEmail(ctx context.Context, otpID string) (otpSession domain.OtpSessionEmail, err error) {

	query := `SELECT * FROM otp_sessions_email WHERE otp_id = $1`

	err = c.DB.Raw(query, otpID).Scan(&otpSession).Error

	return otpSession, err
}
