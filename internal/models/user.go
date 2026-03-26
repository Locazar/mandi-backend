package models

import "time"

type User struct {
    ID           uint      `gorm:"primaryKey;autoIncrement" json:"id"`
    Email        string    `gorm:"uniqueIndex;size:255" json:"email"`
    Phone        string    `gorm:"uniqueIndex;size:30" json:"phone"`
    PasswordHash string    `gorm:"size:255" json:"-"`
    PINHash      string    `gorm:"size:255" json:"-"`
    IsActive     bool      `gorm:"default:true" json:"is_active"`
    CreatedAt    time.Time `json:"created_at"`
    UpdatedAt    time.Time `json:"updated_at"`
}
