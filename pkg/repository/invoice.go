package repository

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"gorm.io/gorm"
)

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
