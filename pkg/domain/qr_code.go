package domain

import "time"

// QRCode is a dynamic QR redirect: an admin registers a target_url and gets a
// permanent short_code. The printed/shared QR encodes a link on our server
// (/r/<short_code>); scanning it hits us and we 302-redirect to target_url.
// Because target_url is editable, the same QR can be repointed at any time.
type QRCode struct {
	ID            string     `json:"id" gorm:"primaryKey;type:varchar(32)"`
	ShortCode     string     `json:"short_code" gorm:"type:varchar(16);not null;uniqueIndex"`
	Title         string     `json:"title" gorm:"size:150;not null;default:''"`
	TargetURL     string     `json:"target_url" gorm:"type:text;not null"`
	IsActive      bool       `json:"is_active" gorm:"not null;default:true"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	ScanCount     int64      `json:"scan_count" gorm:"not null;default:0"`
	LastScannedAt *time.Time `json:"last_scanned_at,omitempty"`
	CreatedBy     string     `json:"created_by,omitempty" gorm:"type:varchar(32)"`
	CreatedAt     time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     *time.Time `json:"-" gorm:"index"`

	// ShortURL and QRImageURL are computed on read (never persisted); they let
	// the admin-portal render/download without knowing the server host.
	ShortURL   string `json:"short_url,omitempty" gorm:"-"`
	QRImageURL string `json:"qr_image_url,omitempty" gorm:"-"`
}

// TableName pins the table so GORM doesn't pluralize to "qr_cods".
func (QRCode) TableName() string { return "qr_codes" }

// QRScanEvent is one recorded scan of a QRCode, for analytics. Written
// best-effort on redirect; a failure here must never block the redirect.
type QRScanEvent struct {
	ID        string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	QRCodeID  string    `json:"qr_code_id" gorm:"type:varchar(32);not null;index"`
	ScannedAt time.Time `json:"scanned_at" gorm:"autoCreateTime"`
	IPAddress string    `json:"ip_address" gorm:"type:varchar(64)"`
	UserAgent string    `json:"user_agent" gorm:"type:text"`
	Referer   string    `json:"referer" gorm:"type:text"`
}

func (QRScanEvent) TableName() string { return "qr_scan_events" }

// IsExpired reports whether the code has passed its optional hard expiry.
func (q QRCode) IsExpired(now time.Time) bool {
	return q.ExpiresAt != nil && now.After(*q.ExpiresAt)
}

// Resolvable reports whether a lookup should redirect (vs. return 410 Gone).
func (q QRCode) Resolvable(now time.Time) bool {
	return q.IsActive && !q.IsExpired(now)
}
