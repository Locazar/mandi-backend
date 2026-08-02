package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"gorm.io/gorm"
)

type InvoiceRepository interface {
	// GetCompanyBillingProfile returns the singleton issuer profile.
	GetCompanyBillingProfile(ctx context.Context) (domain.CompanyBillingProfile, error)
	// UpdateCompanyBillingProfile overwrites the singleton issuer profile.
	UpdateCompanyBillingProfile(ctx context.Context, profile domain.CompanyBillingProfile) (domain.CompanyBillingProfile, error)

	// AllocateInvoiceSequence atomically reserves the next number for the given
	// financial year. Safe under concurrency; never returns a duplicate.
	AllocateInvoiceSequence(ctx context.Context, financialYear string) (int, error)
	CreateInvoice(ctx context.Context, inv domain.Invoice) (domain.Invoice, error)
	FindInvoiceByID(ctx context.Context, invoiceID string) (domain.Invoice, error)
	FindInvoiceBySubscriptionOrderID(ctx context.Context, orderID string) (domain.Invoice, error)
	FindInvoicesByUserID(ctx context.Context, userID string, pagination request.Pagination) ([]domain.Invoice, error)
	ListInvoices(ctx context.Context, filter domain.InvoiceFilter) ([]domain.Invoice, error)
	// SetInvoicePDF records the rendered PDF's object key. The only mutation
	// permitted on an issued invoice.
	SetInvoicePDF(ctx context.Context, invoiceID, objectKey string) error

	// WithTx returns a repository bound to an existing transaction, so invoice
	// issuance can join the subscription payment transaction.
	WithTx(tx *gorm.DB) InvoiceRepository
}
