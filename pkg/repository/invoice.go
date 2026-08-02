package repository

import (
	"context"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

// defaultInvoicePageSize is the fallback page size for invoice listings,
// matching request.GetPagination's own default. Both listing methods need it:
// GORM turns Limit(0) into a literal `LIMIT 0`, which returns no rows rather
// than every row.
const defaultInvoicePageSize = 25

type invoiceDatabase struct {
	db *gorm.DB
}

func NewInvoiceRepository(db *gorm.DB) interfaces.InvoiceRepository {
	return &invoiceDatabase{db: db}
}

func (r *invoiceDatabase) GetCompanyBillingProfile(ctx context.Context) (domain.CompanyBillingProfile, error) {
	var profile domain.CompanyBillingProfile
	err := r.db.WithContext(ctx).
		Where("id = ?", domain.CompanyBillingProfileID).
		First(&profile).Error
	return profile, err
}

func (r *invoiceDatabase) UpdateCompanyBillingProfile(ctx context.Context, profile domain.CompanyBillingProfile) (domain.CompanyBillingProfile, error) {
	profile.ID = domain.CompanyBillingProfileID
	// Select("*") so zero-valued fields (a cleared address line 2, say) are
	// actually written — GORM's default Updates skips zero values.
	//
	// Only "id" is omitted. Do NOT add "updated_at": naming an autoUpdateTime
	// column in Omit is GORM's idiom for disabling timestamp tracking, which
	// would freeze updated_at at its seed value forever and destroy the only
	// audit signal this singleton has.
	err := r.db.WithContext(ctx).
		Model(&domain.CompanyBillingProfile{}).
		Where("id = ?", domain.CompanyBillingProfileID).
		Select("*").
		Omit("id").
		Updates(&profile).Error
	if err != nil {
		return domain.CompanyBillingProfile{}, err
	}
	return r.GetCompanyBillingProfile(ctx)
}

// AllocateInvoiceSequence reserves the next sequence for a financial year in a
// single upsert-returning statement. Doing it in one statement is what makes it
// safe: Postgres takes a row lock for the duration, so concurrent callers
// serialise and each gets a distinct, gapless number.
func (r *invoiceDatabase) AllocateInvoiceSequence(ctx context.Context, financialYear string) (int, error) {
	var seq int
	err := r.db.WithContext(ctx).Raw(`
		INSERT INTO invoice_number_sequences (financial_year, last_sequence)
		VALUES (?, 1)
		ON CONFLICT (financial_year)
		DO UPDATE SET last_sequence = invoice_number_sequences.last_sequence + 1
		RETURNING last_sequence`, financialYear).Scan(&seq).Error
	if err != nil {
		return 0, err
	}
	return seq, nil
}

func (r *invoiceDatabase) CreateInvoice(ctx context.Context, inv domain.Invoice) (domain.Invoice, error) {
	err := r.db.WithContext(ctx).Create(&inv).Error
	return inv, err
}

func (r *invoiceDatabase) FindInvoiceByID(ctx context.Context, invoiceID string) (domain.Invoice, error) {
	var inv domain.Invoice
	err := r.db.WithContext(ctx).Where("id = ?", invoiceID).First(&inv).Error
	return inv, err
}

func (r *invoiceDatabase) FindInvoiceBySubscriptionOrderID(ctx context.Context, orderID string) (domain.Invoice, error) {
	var inv domain.Invoice
	err := r.db.WithContext(ctx).Where("subscription_order_id = ?", orderID).First(&inv).Error
	return inv, err
}

func (r *invoiceDatabase) FindInvoicesByUserID(ctx context.Context, userID string, pagination request.Pagination) ([]domain.Invoice, error) {
	invoices := []domain.Invoice{}

	// GORM emits a literal `LIMIT 0` for Limit(0) — that means "return nothing",
	// not "unlimited". request.GetPagination parses `?limit=0` successfully and
	// overwrites its own default of 25, so without this guard a seller passing
	// ?limit=0 would silently get an empty invoice list instead of their bills.
	limit := int(pagination.Limit)
	if limit <= 0 {
		limit = defaultInvoicePageSize
	}

	err := r.db.WithContext(ctx).
		Where("user_id = ?", userID).
		Order("invoice_date DESC").
		Limit(limit).
		Offset(int(pagination.Offset)).
		Find(&invoices).Error
	return invoices, err
}

func (r *invoiceDatabase) ListInvoices(ctx context.Context, filter domain.InvoiceFilter) ([]domain.Invoice, error) {
	invoices := []domain.Invoice{}
	q := r.db.WithContext(ctx).Model(&domain.Invoice{})

	if filter.UserID != "" {
		q = q.Where("user_id = ?", filter.UserID)
	}
	if filter.From != nil {
		q = q.Where("invoice_date >= ?", *filter.From)
	}
	if filter.To != nil {
		q = q.Where("invoice_date <= ?", *filter.To)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = defaultInvoicePageSize
	}

	err := q.Order("invoice_date DESC").
		Limit(limit).
		Offset(filter.Offset).
		Find(&invoices).Error
	return invoices, err
}

func (r *invoiceDatabase) SetInvoicePDF(ctx context.Context, invoiceID, objectKey string) error {
	now := time.Now()
	return r.db.WithContext(ctx).
		Model(&domain.Invoice{}).
		Where("id = ?", invoiceID).
		Updates(map[string]interface{}{
			"pdf_object_key":   objectKey,
			"pdf_generated_at": now,
		}).Error
}
