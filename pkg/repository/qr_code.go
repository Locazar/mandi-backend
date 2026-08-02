package repository

import (
	"context"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repo "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

type qrCodeDatabase struct {
	DB *gorm.DB
}

func NewQRCodeRepository(db *gorm.DB) repo.QRCodeRepository {
	return &qrCodeDatabase{DB: db}
}

// live scopes a query to rows that haven't been soft-deleted. DeletedAt is a
// plain *time.Time here (not gorm.DeletedAt), so every read filters explicitly.
func qrLive(db *gorm.DB) *gorm.DB {
	return db.Where("deleted_at IS NULL")
}

func (c *qrCodeDatabase) Create(ctx context.Context, qr domain.QRCode) (domain.QRCode, error) {
	qr.ID = domain.NewID(domain.PrefixQRCode)
	err := c.DB.WithContext(ctx).Omit("ShortURL", "QRImageURL").Create(&qr).Error
	return qr, err
}

func (c *qrCodeDatabase) List(ctx context.Context) ([]domain.QRCode, error) {
	var codes []domain.QRCode
	err := qrLive(c.DB.WithContext(ctx)).Order("created_at DESC").Find(&codes).Error
	return codes, err
}

func (c *qrCodeDatabase) GetByID(ctx context.Context, id string) (domain.QRCode, error) {
	var qr domain.QRCode
	err := qrLive(c.DB.WithContext(ctx)).First(&qr, "id = ?", id).Error
	return qr, err
}

func (c *qrCodeDatabase) GetByShortCode(ctx context.Context, shortCode string) (domain.QRCode, error) {
	var qr domain.QRCode
	err := qrLive(c.DB.WithContext(ctx)).First(&qr, "short_code = ?", shortCode).Error
	return qr, err
}

func (c *qrCodeDatabase) Update(ctx context.Context, qr domain.QRCode) (domain.QRCode, error) {
	qr.UpdatedAt = time.Now()
	err := qrLive(c.DB.WithContext(ctx)).Model(&domain.QRCode{}).
		Where("id = ?", qr.ID).
		Updates(map[string]interface{}{
			"title":      qr.Title,
			"target_url": qr.TargetURL,
			"is_active":  qr.IsActive,
			"expires_at": qr.ExpiresAt,
			"updated_at": qr.UpdatedAt,
		}).Error
	if err != nil {
		return qr, err
	}
	return c.GetByID(ctx, qr.ID)
}

func (c *qrCodeDatabase) SoftDelete(ctx context.Context, id string) error {
	now := time.Now()
	res := qrLive(c.DB.WithContext(ctx)).Model(&domain.QRCode{}).
		Where("id = ?", id).
		Updates(map[string]interface{}{"deleted_at": now, "updated_at": now})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	return nil
}

func (c *qrCodeDatabase) ShortCodeExists(ctx context.Context, shortCode string) (bool, error) {
	var count int64
	// Uniqueness is enforced only among live rows (partial unique index), so
	// this check matches what the DB will actually allow on insert.
	err := qrLive(c.DB.WithContext(ctx)).Model(&domain.QRCode{}).
		Where("short_code = ?", shortCode).Count(&count).Error
	return count > 0, err
}

func (c *qrCodeDatabase) RecordScan(ctx context.Context, event domain.QRScanEvent) error {
	event.ID = domain.NewID(domain.PrefixQRScanEvent)
	if event.ScannedAt.IsZero() {
		event.ScannedAt = time.Now()
	}
	return c.DB.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&event).Error; err != nil {
			return err
		}
		return tx.Model(&domain.QRCode{}).
			Where("id = ?", event.QRCodeID).
			Updates(map[string]interface{}{
				"scan_count":      gorm.Expr("scan_count + 1"),
				"last_scanned_at": event.ScannedAt,
			}).Error
	})
}

func (c *qrCodeDatabase) ListScanEvents(ctx context.Context, qrCodeID string, limit int) ([]domain.QRScanEvent, error) {
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	var events []domain.QRScanEvent
	err := c.DB.WithContext(ctx).
		Where("qr_code_id = ?", qrCodeID).
		Order("scanned_at DESC").
		Limit(limit).
		Find(&events).Error
	return events, err
}
