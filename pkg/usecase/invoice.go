package usecase

import (
	"context"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoIface "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
	invoicesvc "github.com/rohit221990/mandi-backend/pkg/service/invoice"
	service "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
	"gorm.io/gorm"
)

type invoiceUseCase struct {
	invoiceRepo repoIface.InvoiceRepository
	subRepo     repoIface.SubscriptionRepository
	adminRepo   repoIface.AdminRepository
	renderer    invoicesvc.Renderer
	cloud       cloud.CloudService
}

func NewInvoiceUseCase(
	invoiceRepo repoIface.InvoiceRepository,
	subRepo repoIface.SubscriptionRepository,
	adminRepo repoIface.AdminRepository,
	renderer invoicesvc.Renderer,
	cs cloud.CloudService,
) service.InvoiceUseCase {
	return &invoiceUseCase{
		invoiceRepo: invoiceRepo,
		subRepo:     subRepo,
		adminRepo:   adminRepo,
		renderer:    renderer,
		cloud:       cs,
	}
}

func (uc *invoiceUseCase) GetCompanyBillingProfile(ctx context.Context) (domain.CompanyBillingProfile, error) {
	return uc.invoiceRepo.GetCompanyBillingProfile(ctx)
}

// UpdateCompanyBillingProfile overwrites the singleton. The ID is forced rather
// than trusted from the caller so a stray body value can't create a second row.
func (uc *invoiceUseCase) UpdateCompanyBillingProfile(ctx context.Context, profile domain.CompanyBillingProfile) (domain.CompanyBillingProfile, error) {
	profile.ID = domain.CompanyBillingProfileID
	if profile.InvoiceNumberPrefix == "" {
		profile.InvoiceNumberPrefix = "LZ"
	}
	return uc.invoiceRepo.UpdateCompanyBillingProfile(ctx, profile)
}

// invoiceURLTTL is how long a presigned download link stays valid. Short,
// because the client fetches it immediately.
const invoiceURLTTL = 15 * time.Minute

// GenerateAndStorePDF renders the invoice and uploads it to private storage,
// returning the object key. Callers on the payment path must treat failure as
// non-fatal: the invoice row is already durable and the download path will
// re-render on demand.
func (uc *invoiceUseCase) GenerateAndStorePDF(ctx context.Context, inv domain.Invoice) (string, error) {
	profile, err := uc.invoiceRepo.GetCompanyBillingProfile(ctx)
	if err != nil {
		log.Printf("[INVOICE_PDF] profile lookup failed, rendering without logo: %v", err)
	}

	logo := uc.fetchLogo(ctx, profile.LogoObjectKey)

	pdfBytes, err := uc.renderer.Render(ctx, inv, logo)
	if err != nil {
		return "", fmt.Errorf("render invoice %s: %w", inv.InvoiceNumber, err)
	}

	key, err := uc.cloud.SaveBytes(ctx, pdfBytes, cloud.SaveOptions{
		Namespace:   "invoices",
		Visibility:  cloud.VisibilityPrivate,
		ContentType: "application/pdf",
		Filename:    inv.FileName(),
	})
	if err != nil {
		return "", fmt.Errorf("upload invoice %s: %w", inv.InvoiceNumber, err)
	}
	return key, nil
}

// fetchLogo best-effort reads the configured logo straight from object
// storage. A missing or unreadable logo renders a logo-less invoice rather
// than failing — the invoice matters, the logo does not.
func (uc *invoiceUseCase) fetchLogo(ctx context.Context, objectKey string) []byte {
	if objectKey == "" || uc.cloud == nil {
		return nil
	}
	b, err := uc.cloud.GetBytes(ctx, objectKey)
	if err != nil {
		log.Printf("[INVOICE_PDF] logo %q unreadable, rendering without it: %v", objectKey, err)
		return nil
	}
	return b
}

// GetInvoiceDownload returns a short-lived presigned URL for the invoice PDF,
// re-rendering and backfilling the object when the cache is empty or lost.
func (uc *invoiceUseCase) GetInvoiceDownload(ctx context.Context, invoiceID, requesterUserID string, isAdmin bool) (response.InvoiceDownloadResponse, error) {
	inv, err := uc.invoiceRepo.FindInvoiceByID(ctx, invoiceID)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return response.InvoiceDownloadResponse{}, ErrInvoiceNotFound
		}
		return response.InvoiceDownloadResponse{}, fmt.Errorf("find invoice: %w", err)
	}

	if !isAdmin && inv.UserID != requesterUserID {
		return response.InvoiceDownloadResponse{}, ErrInvoiceNotOwned
	}

	objectKey := inv.PDFObjectKey
	if objectKey == "" {
		key, err := uc.GenerateAndStorePDF(ctx, inv)
		if err != nil {
			return response.InvoiceDownloadResponse{}, err
		}
		if err := uc.invoiceRepo.SetInvoicePDF(ctx, inv.ID, key); err != nil {
			// The PDF exists; failing to cache the key only costs a re-render.
			log.Printf("[INVOICE_PDF] backfill key failed for %s: %v", inv.ID, err)
		}
		objectKey = key
	}

	url, err := uc.cloud.PresignedURL(ctx, objectKey, invoiceURLTTL)
	if err != nil {
		return response.InvoiceDownloadResponse{}, fmt.Errorf("presign invoice url: %w", err)
	}

	return response.InvoiceDownloadResponse{
		DownloadURL: url,
		FileName:    inv.FileName(),
		ExpiresAt:   time.Now().Add(invoiceURLTTL).UTC().Format(time.RFC3339),
	}, nil
}

func (uc *invoiceUseCase) ListInvoicesForUser(ctx context.Context, userID string, pagination request.Pagination) ([]response.InvoiceListItem, error) {
	invoices, err := uc.invoiceRepo.FindInvoicesByUserID(ctx, userID, pagination)
	if err != nil {
		return nil, err
	}
	return toInvoiceListItems(invoices), nil
}

func (uc *invoiceUseCase) ListInvoicesForAdmin(ctx context.Context, filter domain.InvoiceFilter) ([]response.InvoiceListItem, error) {
	invoices, err := uc.invoiceRepo.ListInvoices(ctx, filter)
	if err != nil {
		return nil, err
	}
	return toInvoiceListItems(invoices), nil
}

func toInvoiceListItems(invoices []domain.Invoice) []response.InvoiceListItem {
	items := make([]response.InvoiceListItem, 0, len(invoices))
	for _, inv := range invoices {
		items = append(items, response.InvoiceListItem{
			ID:                 inv.ID,
			InvoiceNumber:      inv.InvoiceNumber,
			InvoiceDate:        inv.InvoiceDate.UTC().Format(time.RFC3339),
			PlanName:           inv.PlanName,
			BuyerName:          inv.BuyerName,
			TotalMinor:         inv.Total.AmountMinor,
			TaxableValueMinor:  inv.TaxableValue.AmountMinor,
			GSTAmountMinor:     inv.GSTAmount.AmountMinor,
			GSTRateBasisPoints: inv.GSTRateBasisPoints,
			Currency:           inv.Total.Currency,
			RazorpayPaymentID:  inv.RazorpayPaymentID,
			PaidAt:             inv.PaidAt.UTC().Format(time.RFC3339),
		})
	}
	return items
}
