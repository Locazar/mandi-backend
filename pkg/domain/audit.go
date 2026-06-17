// pkg/domain/audit.go
package domain

import (
	"time"

	"gorm.io/gorm"
)

// AuditLog is an append-only record of privileged actions / PII access (DPDP
// record-of-processing). Generalized from LoginAuditLog.
type AuditLog struct {
	ID         string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	ActorID    string    `json:"actor_id" gorm:"type:varchar(32);index"`
	ActorType  UserType  `json:"actor_type" gorm:"type:varchar(20)"`
	Action     string    `json:"action" gorm:"size:100;not null;index"`
	EntityType string    `json:"entity_type" gorm:"size:50;index"`
	EntityID   string    `json:"entity_id" gorm:"type:varchar(32);index"`
	IPAddress  string    `json:"ip_address" gorm:"size:50"`
	Metadata   string    `json:"metadata" gorm:"type:jsonb"`
	CreatedAt  time.Time `json:"created_at" gorm:"autoCreateTime;index:,sort:desc"`
}

func (m *AuditLog) BeforeCreate(*gorm.DB) error {
	m.ID = NewID(PrefixAuditLog)
	return nil
}
