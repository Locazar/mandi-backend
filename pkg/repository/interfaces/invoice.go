package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type InvoiceRepository interface {
	// GetCompanyBillingProfile returns the singleton issuer profile.
	GetCompanyBillingProfile(ctx context.Context) (domain.CompanyBillingProfile, error)
	// UpdateCompanyBillingProfile overwrites the singleton issuer profile.
	UpdateCompanyBillingProfile(ctx context.Context, profile domain.CompanyBillingProfile) (domain.CompanyBillingProfile, error)

	// AllocateInvoiceSequence atomically reserves the next number for the given
	// financial year. Safe under concurrency; never returns a duplicate.
	//
	// No production caller uses this directly anymore — both the live payment
	// path and the backfill command use CreateInvoiceWithSequence, which wraps
	// this same allocation and the insert in one transaction so a failed
	// insert doesn't burn the number. This method is kept exported only
	// because TestAllocateInvoiceSequenceIsConcurrencySafe exercises it
	// directly to prove the underlying upsert is race-safe. Do not call it on
	// its own to issue an invoice — that's the exact bug
	// CreateInvoiceWithSequence exists to prevent.
	AllocateInvoiceSequence(ctx context.Context, financialYear string) (int, error)
	CreateInvoice(ctx context.Context, inv domain.Invoice) (domain.Invoice, error)
	// CreateInvoiceWithSequence allocates the next sequence number for
	// financialYear and creates the invoice built from it in a single
	// transaction. If build's invoice fails to insert for any reason —
	// including a benign subscription_order_id duplicate — the sequence
	// allocation rolls back too, so a failed attempt never burns a number in a
	// series required to be gapless.
	CreateInvoiceWithSequence(ctx context.Context, financialYear string, build func(sequence int) domain.Invoice) (domain.Invoice, error)
	FindInvoiceByID(ctx context.Context, invoiceID string) (domain.Invoice, error)
	FindInvoiceBySubscriptionOrderID(ctx context.Context, orderID string) (domain.Invoice, error)
	FindInvoicesByUserID(ctx context.Context, userID string, pagination request.Pagination) ([]domain.Invoice, error)
	ListInvoices(ctx context.Context, filter domain.InvoiceFilter) ([]domain.Invoice, error)
	// SetInvoicePDF records the rendered PDF's object key. The only mutation
	// permitted on an issued invoice.
	SetInvoicePDF(ctx context.Context, invoiceID, objectKey string) error
	// FindInvoiceNumbersByOrderIDs returns invoices keyed by subscription order
	// id, so billing history can be decorated in one query rather than N.
	FindInvoiceNumbersByOrderIDs(ctx context.Context, orderIDs []string) (map[string]domain.Invoice, error)
}
