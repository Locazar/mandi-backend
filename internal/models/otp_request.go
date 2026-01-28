package models

import "time"

type OTPStatus string

const (
    OTPStatusPending OTPStatus = "pending"
    OTPStatusUsed    OTPStatus = "used"
    OTPStatusFailed  OTPStatus = "failed"
)

type OTPRequest struct {
    ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Target     string    `gorm:"index;size:255" json:"target"` // phone or email
    OTPHash    string    `gorm:"size:255" json:"-"`
    ExpiresAt  time.Time `json:"expires_at"`
    Attempts   int       `gorm:"default:0" json:"attempts"`
    MaxAttempt int       `gorm:"default:3" json:"max_attempt"`
    Status     OTPStatus `gorm:"size:50;default:'pending'" json:"status"`
    IPAddress  string    `gorm:"size:50" json:"ip_address"`
    UserAgent  string    `gorm:"size:255" json:"user_agent"`
    CreatedAt  time.Time `json:"created_at"`
}
