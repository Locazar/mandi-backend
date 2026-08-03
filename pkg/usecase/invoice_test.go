package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/mock/mockrepo"
	"github.com/stretchr/testify/assert"
)

func TestGetCompanyBillingProfile(t *testing.T) {
	tests := map[string]struct {
		buildStub func(*mockrepo.MockInvoiceRepository)
		wantName  string
		wantErr   bool
	}{
		"success": {
			buildStub: func(r *mockrepo.MockInvoiceRepository) {
				r.EXPECT().GetCompanyBillingProfile(gomock.Any()).
					Return(domain.CompanyBillingProfile{
						ID:        "cbp_default",
						LegalName: "Localzar Technologies Pvt. Ltd.",
					}, nil)
			},
			wantName: "Localzar Technologies Pvt. Ltd.",
		},
		"repository error propagates": {
			buildStub: func(r *mockrepo.MockInvoiceRepository) {
				r.EXPECT().GetCompanyBillingProfile(gomock.Any()).
					Return(domain.CompanyBillingProfile{}, errors.New("db down"))
			},
			wantErr: true,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			repo := mockrepo.NewMockInvoiceRepository(ctrl)
			tc.buildStub(repo)

			uc := NewInvoiceUseCase(repo, nil, nil, nil, nil)
			got, err := uc.GetCompanyBillingProfile(context.Background())

			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantName, got.LegalName)
		})
	}
}

func TestUpdateCompanyBillingProfile(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	in := domain.CompanyBillingProfile{LegalName: "New Name", GSTIN: "29AABCL1234M1Z7"}

	repo := mockrepo.NewMockInvoiceRepository(ctrl)
	// The usecase must force the singleton ID rather than trusting the caller.
	repo.EXPECT().UpdateCompanyBillingProfile(gomock.Any(), gomock.Any()).
		DoAndReturn(func(_ context.Context, p domain.CompanyBillingProfile) (domain.CompanyBillingProfile, error) {
			assert.Equal(t, "cbp_default", p.ID)
			return p, nil
		})

	uc := NewInvoiceUseCase(repo, nil, nil, nil, nil)
	got, err := uc.UpdateCompanyBillingProfile(context.Background(), in)

	assert.NoError(t, err)
	assert.Equal(t, "New Name", got.LegalName)
}
