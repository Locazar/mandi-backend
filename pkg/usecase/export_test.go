package usecase

import (
	"context"
	"errors"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/api/handler/response"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
)

// noopInvoiceUseCase satisfies service.InvoiceUseCase for tests that drive
// issueInvoiceForPaidOrder directly and don't care about PDF generation. Its
// GenerateAndStorePDF errors rather than panicking, so the background
// goroutine issueInvoiceForPaidOrder launches after a successful issuance logs
// a failure and returns, same as it would against a real renderer/cloud that
// isn't configured — it never crashes the test process.
type noopInvoiceUseCase struct{}

func (noopInvoiceUseCase) GetCompanyBillingProfile(ctx context.Context) (domain.CompanyBillingProfile, error) {
	return domain.CompanyBillingProfile{}, errors.New("noopInvoiceUseCase: not implemented")
}

func (noopInvoiceUseCase) UpdateCompanyBillingProfile(ctx context.Context, profile domain.CompanyBillingProfile) (domain.CompanyBillingProfile, error) {
	return domain.CompanyBillingProfile{}, errors.New("noopInvoiceUseCase: not implemented")
}

func (noopInvoiceUseCase) GenerateAndStorePDF(ctx context.Context, inv domain.Invoice) (string, error) {
	return "", errors.New("noopInvoiceUseCase: pdf generation not configured in test")
}

func (noopInvoiceUseCase) GetInvoiceDownload(ctx context.Context, invoiceID, requesterUserID string, isAdmin bool) (response.InvoiceDownloadResponse, error) {
	return response.InvoiceDownloadResponse{}, errors.New("noopInvoiceUseCase: not implemented")
}

func (noopInvoiceUseCase) ListInvoicesForUser(ctx context.Context, userID string, pagination request.Pagination) ([]response.InvoiceListItem, error) {
	return nil, errors.New("noopInvoiceUseCase: not implemented")
}

func (noopInvoiceUseCase) ListInvoicesForAdmin(ctx context.Context, filter domain.InvoiceFilter) ([]response.InvoiceListItem, error) {
	return nil, errors.New("noopInvoiceUseCase: not implemented")
}

// This file exists solely to give subscription_payment_integration_test.go
// (package usecase_test) access to subscriptionPaymentUseCase's unexported
// bits, following Go's export_test.go convention: it is compiled only by
// `go test`, never into the production binary.
//
// The integration test has to live in an EXTERNAL test package rather than
// an internal one, even though the task it implements is package-local:
// pkg/repository/alert.go imports pkg/usecase (a pre-existing dependency this
// task doesn't own), so an internal `package usecase` test file that also
// imports pkg/repository (needed for real repository constructors) would
// form an import cycle. `package usecase_test` sidesteps that — it is a
// distinct package from `usecase`, so it can import both without cycling.

// NewTestSubscriptionPaymentUseCase builds a subscriptionPaymentUseCase from
// real repositories, for tests that need to drive markPaidAndActivate against
// a live database. Only the fields markPaidAndActivate touches are wired.
func NewTestSubscriptionPaymentUseCase(
	subRepo interfaces.SubscriptionRepository,
	invoiceRepo interfaces.InvoiceRepository,
	adminRepo interfaces.AdminRepository,
) *subscriptionPaymentUseCase {
	return &subscriptionPaymentUseCase{
		subRepo:     subRepo,
		invoiceRepo: invoiceRepo,
		adminRepo:   adminRepo,
		invoiceUC:   noopInvoiceUseCase{},
	}
}

// MarkPaidAndActivate exposes the unexported markPaidAndActivate for the
// external integration test.
func (uc *subscriptionPaymentUseCase) MarkPaidAndActivate(
	ctx context.Context,
	order domain.SubscriptionOrder,
	plan domain.SubscriptionPlan,
	paymentID string,
) error {
	return uc.markPaidAndActivate(ctx, order, plan, paymentID)
}

// IssueInvoiceForPaidOrder exposes the unexported issueInvoiceForPaidOrder so
// the integration test can drive invoice issuance directly, bypassing
// markPaidAndActivate's transaction. That bypass is the point: it is the only
// way to exercise the subscription_order_id unique index, which is the actual
// duplicate guard now that issuance runs outside the payment transaction.
func (uc *subscriptionPaymentUseCase) IssueInvoiceForPaidOrder(
	ctx context.Context,
	order domain.SubscriptionOrder,
	plan domain.SubscriptionPlan,
	paymentID string,
) {
	uc.issueInvoiceForPaidOrder(ctx, order, plan, paymentID)
}
