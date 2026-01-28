package models

import "time"

type LoginEvent string

const (
    EventLoginSuccess LoginEvent = "login_success"
    EventLoginFail    LoginEvent = "login_fail"
    EventOTPRequested LoginEvent = "otp_requested"
    EventOTPVerified  LoginEvent = "otp_verified"
    EventLogout       LoginEvent = "logout"
)

type LoginAuditLog struct {
    ID        uint       `gorm:"primaryKey;autoIncrement" json:"id"`
    UserID    *uint      `gorm:"index" json:"user_id"`
    EventType LoginEvent `gorm:"size:50" json:"event_type"`
    IPAddress string     `gorm:"size:50" json:"ip_address"`
    UserAgent string     `gorm:"size:255" json:"user_agent"`
    CreatedAt time.Time  `json:"created_at"`
}
