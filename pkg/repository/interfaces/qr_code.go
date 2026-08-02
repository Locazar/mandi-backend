package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// QRCodeRepository is the persistence contract for dynamic QR redirects:
// admin CRUD plus the public short-code lookup and best-effort scan logging.
// All reads exclude soft-deleted rows.
type QRCodeRepository interface {
	Create(ctx context.Context, qr domain.QRCode) (domain.QRCode, error)
	List(ctx context.Context) ([]domain.QRCode, error)
	GetByID(ctx context.Context, id string) (domain.QRCode, error)
	// GetByShortCode powers the public redirect; returns gorm.ErrRecordNotFound
	// when no live row matches.
	GetByShortCode(ctx context.Context, shortCode string) (domain.QRCode, error)
	// Update writes only the mutable fields (title, target_url, is_active,
	// expires_at) of an existing, non-deleted row.
	Update(ctx context.Context, qr domain.QRCode) (domain.QRCode, error)
	SoftDelete(ctx context.Context, id string) error

	// ShortCodeExists checks liveness of a candidate code during generation.
	ShortCodeExists(ctx context.Context, shortCode string) (bool, error)

	// RecordScan inserts a scan event and bumps scan_count/last_scanned_at in a
	// single transaction. Best-effort at the call site.
	RecordScan(ctx context.Context, event domain.QRScanEvent) error
	ListScanEvents(ctx context.Context, qrCodeID string, limit int) ([]domain.QRScanEvent, error)
}
