package invoice

import (
	"context"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// Renderer turns an issued invoice into a PDF document.
//
// It takes only the invoice — every printed value, including the issuer's
// logo object key, is already snapshotted on it — plus the logo's raw bytes
// (fetched by the caller using that snapshotted key). It deliberately does
// NOT take the live CompanyBillingProfile, so editing the profile (including
// swapping the logo) can never change how an already-issued invoice renders.
type Renderer interface {
	Render(ctx context.Context, inv domain.Invoice, logoPNG []byte) ([]byte, error)
}
