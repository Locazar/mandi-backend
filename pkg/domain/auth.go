package domain

import (
	"time"
)

type AdminRefreshSession struct {
	TokenID      string    `json:"token_id" gorm:"primaryKey;not null"`
	UserID       string    `json:"user_id" gorm:"type:varchar(32)"`
	AdminID      string    `json:"admin_id" gorm:"type:varchar(32)"`
	UserType     UserType  `json:"user_type"`
	RefreshToken string    `json:"-" gorm:"not null"`
	ExpireAt     time.Time `json:"expire_at" gorm:"not null"`
	IsBlocked    bool      `json:"is_blocked" gorm:"not null;default:false"`
}

type UserRefreshSession struct {
	TokenID      string    `json:"token_id" gorm:"primaryKey;not null"`
	UserID       string    `json:"user_id" gorm:"type:varchar(32)"`
	AdminID      string    `json:"admin_id" gorm:"type:varchar(32)"`
	UserType     UserType  `json:"user_type"`
	RefreshToken string    `json:"-" gorm:"not null"`
	ExpireAt     time.Time `json:"expire_at" gorm:"not null"`
	IsBlocked    bool      `json:"is_blocked" gorm:"not null;default:false"`
}

type OtpSession struct {
	ID       string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	OtpID    string    `json:"otp_id" gorm:"unique;not null"`
	UserID   string    `json:"user_id" gorm:"type:varchar(32)"`
	AdminID  string    `json:"admin_id" gorm:"type:varchar(32)"`
	UserType UserType  `json:"user_type"`
	Phone    string    `json:"phone"`
	ExpireAt time.Time `json:"expire_at"`
}

type OtpSessionEmail struct {
	ID       string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	OtpID    string    `json:"otp_id" gorm:"unique;not null"`
	UserID   string    `json:"user_id" gorm:"type:varchar(32);not null"`
	AdminID  string    `json:"admin_id" gorm:"type:varchar(32);not null"`
	UserType UserType  `json:"user_type" gorm:"not null"`
	Email    string    `json:"email" gorm:"not null"`
	ExpireAt time.Time `json:"expire_at" gorm:"not null"`
}

// RefreshSession represents a stored refresh token session used by the auth
// usecases and tests. It intentionally mirrors the request/response structures
// used elsewhere in the codebase so mocks and tests can operate against it.
type RefreshSession struct {
	TokenID      string    `json:"token_id"`
	UserID       string    `json:"user_id"`
	UserType     UserType  `json:"user_type"`
	RefreshToken string    `json:"-"`
	ExpireAt     time.Time `json:"expire_at"`
	IsBlocked    bool      `json:"is_blocked"`
}
