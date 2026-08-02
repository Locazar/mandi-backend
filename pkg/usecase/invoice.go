package usecase

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	repoIface "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	service "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type invoiceUseCase struct {
	invoiceRepo repoIface.InvoiceRepository
	subRepo     repoIface.SubscriptionRepository
	adminRepo   repoIface.AdminRepository
}

func NewInvoiceUseCase(
	invoiceRepo repoIface.InvoiceRepository,
	subRepo repoIface.SubscriptionRepository,
	adminRepo repoIface.AdminRepository,
) service.InvoiceUseCase {
	return &invoiceUseCase{
		invoiceRepo: invoiceRepo,
		subRepo:     subRepo,
		adminRepo:   adminRepo,
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
