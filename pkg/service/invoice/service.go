package invoice

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// Renderer turns an issued invoice into a PDF document.
//
// It takes only the invoice — every printed value is already snapshotted on it
// — plus optional logo bytes. It deliberately does NOT take the live
// CompanyBillingProfile, so editing the profile can never change how an
// already-issued invoice renders.
type Renderer interface {
	Render(ctx context.Context, inv domain.Invoice, logoPNG []byte) ([]byte, error)
}
