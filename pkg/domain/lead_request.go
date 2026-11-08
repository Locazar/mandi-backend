package domain

import "time"

type LeadRequest struct {
	ID          uint      `json:"id" gorm:"primaryKey;autoIncrement"`
	Phone       string    `json:"phone" gorm:"type:varchar(20);not null;index"`
	Requirement string    `json:"requirement" gorm:"type:text;not null"`
	City        string    `json:"city" gorm:"type:varchar(100);not null"`
	Source      string    `json:"source" gorm:"type:varchar(500)"`
	Status      string    `json:"status" gorm:"type:varchar(20);not null;default:'new'"` // new, contacted, converted
	AdminNote   string    `json:"admin_note" gorm:"type:text"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
