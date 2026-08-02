package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

type InvoiceRepository interface {
	// GetCompanyBillingProfile returns the singleton issuer profile.
	GetCompanyBillingProfile(ctx context.Context) (domain.CompanyBillingProfile, error)
	// UpdateCompanyBillingProfile overwrites the singleton issuer profile.
	UpdateCompanyBillingProfile(ctx context.Context, profile domain.CompanyBillingProfile) (domain.CompanyBillingProfile, error)
}
