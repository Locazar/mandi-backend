package interfaces

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

//go:generate mockgen -source=invoice.go -destination=../../mock/mockusecase/invoice_mock.go -package=mockusecase

type InvoiceUseCase interface {
	GetCompanyBillingProfile(ctx context.Context) (domain.CompanyBillingProfile, error)
	UpdateCompanyBillingProfile(ctx context.Context, profile domain.CompanyBillingProfile) (domain.CompanyBillingProfile, error)
}
