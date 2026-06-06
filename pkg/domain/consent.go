// pkg/domain/consent.go
package domain

import (
	"time"

	"gorm.io/gorm"
)

// UserConsent is a demonstrable, versioned record of a user's consent (DPDP).
// Replaces the single Admin.AgreeToTerms bool.
type UserConsent struct {
	ID           string      `json:"id" gorm:"primaryKey;type:varchar(32)"`
	UserID       string      `json:"user_id" gorm:"type:varchar(32);not null;index"`
	ConsentType  ConsentType `json:"consent_type" gorm:"type:varchar(20);not null"`
	TermsVersion string      `json:"terms_version" gorm:"size:50;not null"`
	Accepted     bool        `json:"accepted" gorm:"not null"`
	AcceptedAt   time.Time   `json:"accepted_at" gorm:"autoCreateTime"`
	IPAddress    string      `json:"ip_address" gorm:"size:50"`
	UserAgent    string      `json:"user_agent" gorm:"type:text"`
}

func (m *UserConsent) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixUserConsent)
	return nil
}
