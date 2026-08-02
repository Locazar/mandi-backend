package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

//go:generate mockgen -source=invoice.go -destination=../../mock/mockusecase/invoice_mock.go -package=mockusecase

type InvoiceUseCase interface {
	GetCompanyBillingProfile(ctx context.Context) (domain.CompanyBillingProfile, error)
	UpdateCompanyBillingProfile(ctx context.Context, profile domain.CompanyBillingProfile) (domain.CompanyBillingProfile, error)

	GenerateAndStorePDF(ctx context.Context, inv domain.Invoice) (objectKey string, err error)
	GetInvoiceDownload(ctx context.Context, invoiceID string, requesterUserID string, isAdmin bool) (response.InvoiceDownloadResponse, error)
	ListInvoicesForUser(ctx context.Context, userID string, pagination request.Pagination) ([]response.InvoiceListItem, error)
	ListInvoicesForAdmin(ctx context.Context, filter domain.InvoiceFilter) ([]response.InvoiceListItem, error)
}
