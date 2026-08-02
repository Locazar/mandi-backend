package usecase

import (
	"context"
	"errors"
	"mime/multipart"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/mock/mockrepo"
	repoIface "github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
	invoicesvc "github.com/rohit221990/mandi-backend/pkg/service/invoice"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

func storedInvoice() domain.Invoice {
	return domain.Invoice{
		ID:            "inv_1",
		UserID:        "adm_owner",
		InvoiceNumber: "LZ/2026-27/000042",
		PDFObjectKey:  "invoices/LZ-2026-27-000042.pdf",
		InvoiceDate:   time.Now(),
		Total:         domain.INR(149900),
		TaxableValue:  domain.INR(127034),
		GSTAmount:     domain.INR(22866),
	}
}

func TestGetInvoiceDownloadRejectsOtherSellers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockrepo.NewMockInvoiceRepository(ctrl)
	repo.EXPECT().FindInvoiceByID(gomock.Any(), "inv_1").Return(storedInvoice(), nil)

	uc := newInvoiceUseCaseForTest(repo, nil, nil)
	_, err := uc.GetInvoiceDownload(context.Background(), "inv_1", "adm_someone_else", false)

	// A seller must never be able to fetch another seller's invoice.
	assert.ErrorIs(t, err, ErrInvoiceNotOwned)
}

func TestGetInvoiceDownloadAllowsAdminRegardlessOfOwner(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockrepo.NewMockInvoiceRepository(ctrl)
	repo.EXPECT().FindInvoiceByID(gomock.Any(), "inv_1").Return(storedInvoice(), nil)

	uc := newInvoiceUseCaseForTest(repo, nil, stubCloud{url: "https://s3.example/signed"})
	got, err := uc.GetInvoiceDownload(context.Background(), "inv_1", "", true)

	assert.NoError(t, err)
	assert.Equal(t, "https://s3.example/signed", got.DownloadURL)
	assert.Equal(t, "LZ-2026-27-000042.pdf", got.FileName)
}

func TestGetInvoiceDownloadMapsMissingRowToSentinel(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockrepo.NewMockInvoiceRepository(ctrl)
	repo.EXPECT().FindInvoiceByID(gomock.Any(), "inv_missing").
		Return(domain.Invoice{}, gorm.ErrRecordNotFound)

	uc := newInvoiceUseCaseForTest(repo, nil, nil)
	_, err := uc.GetInvoiceDownload(context.Background(), "inv_missing", "adm_owner", false)

	assert.ErrorIs(t, err, ErrInvoiceNotFound)
}

// A lost or never-written S3 object must self-heal: re-render from the snapshot
// and backfill the key rather than failing the download.
func TestGetInvoiceDownloadRerendersWhenObjectKeyMissing(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	inv := storedInvoice()
	inv.PDFObjectKey = ""

	repo := mockrepo.NewMockInvoiceRepository(ctrl)
	repo.EXPECT().FindInvoiceByID(gomock.Any(), "inv_1").Return(inv, nil)
	repo.EXPECT().GetCompanyBillingProfile(gomock.Any()).
		Return(domain.CompanyBillingProfile{}, nil)
	repo.EXPECT().SetInvoicePDF(gomock.Any(), "inv_1", "invoices/regenerated.pdf").Return(nil)

	uc := newInvoiceUseCaseForTest(
		repo,
		stubRenderer{out: []byte("%PDF-1.4\n%%EOF")},
		stubCloud{key: "invoices/regenerated.pdf", url: "https://s3.example/fresh"},
	)

	got, err := uc.GetInvoiceDownload(context.Background(), "inv_1", "adm_owner", false)

	assert.NoError(t, err)
	assert.Equal(t, "https://s3.example/fresh", got.DownloadURL)
}

func TestGenerateAndStorePDFFailureIsReported(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	repo := mockrepo.NewMockInvoiceRepository(ctrl)
	repo.EXPECT().GetCompanyBillingProfile(gomock.Any()).
		Return(domain.CompanyBillingProfile{}, nil)

	uc := newInvoiceUseCaseForTest(
		repo,
		stubRenderer{err: errors.New("font missing")},
		stubCloud{},
	)

	_, err := uc.GenerateAndStorePDF(context.Background(), storedInvoice())
	assert.Error(t, err)
}

type stubRenderer struct {
	out []byte
	err error
}

func (s stubRenderer) Render(_ context.Context, _ domain.Invoice, _ []byte) ([]byte, error) {
	return s.out, s.err
}

type stubCloud struct {
	key string
	url string
	err error
}

func (s stubCloud) SaveFile(_ context.Context, _ *multipart.FileHeader, _ cloud.SaveOptions) (string, error) {
	return s.key, s.err
}
func (s stubCloud) SaveBytes(_ context.Context, _ []byte, _ cloud.SaveOptions) (string, error) {
	return s.key, s.err
}
func (s stubCloud) PublicURL(_ string) string { return s.url }
func (s stubCloud) PresignedURL(_ context.Context, _ string, _ time.Duration) (string, error) {
	return s.url, s.err
}
func (s stubCloud) DeleteObject(_ context.Context, _ string) error { return nil }
func (s stubCloud) ListObjects(_ context.Context, _ string) ([]string, error) {
	return nil, nil
}

// GetBytes is on the real CloudService interface (pkg/service/cloud/cloud.go)
// and is how the renderer reads the logo. Omitting it means stubCloud does not
// satisfy the interface and the package will not compile.
func (s stubCloud) GetBytes(_ context.Context, _ string) ([]byte, error) {
	return nil, errors.New("no logo in tests")
}

// newInvoiceUseCaseForTest builds the concrete usecase so tests can reach
// methods not on the public interface.
func newInvoiceUseCaseForTest(
	repo repoIface.InvoiceRepository,
	renderer invoicesvc.Renderer,
	cs cloud.CloudService,
) *invoiceUseCase {
	return &invoiceUseCase{invoiceRepo: repo, renderer: renderer, cloud: cs}
}
