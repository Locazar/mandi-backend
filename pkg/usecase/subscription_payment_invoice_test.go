package usecase

import (
	"testing"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/stretchr/testify/assert"
)

// The invoice built during payment must reconcile exactly and carry the
// payment identifiers, because those are what a seller's accountant matches
// against the bank statement.
func TestInvoiceIssuedDuringPaymentReconciles(t *testing.T) {
	paidAt := time.Date(2026, 7, 30, 16, 18, 0, 0, time.UTC)
	order := domain.SubscriptionOrder{
		ID:                 "subo_abc",
		UserID:             "adm_seller1",
		Price:              domain.INR(149900),
		GSTAmount:          domain.INR(22866),
		GSTRateBasisPoints: 1800,
		RazorpayOrderID:    "order_x",
		RazorpayPaymentID:  strPtr("pay_x"),
		PaidAt:             &paidAt,
	}

	inv := BuildInvoice(BuildInvoiceInput{
		Order:          order,
		Plan:           domain.SubscriptionPlan{Name: "3 Months", DurationDays: 90},
		Shop:           domain.ShopDetails{ShopName: "Sharma Kirana Store"},
		Profile:        domain.CompanyBillingProfile{LegalName: "Locazar", InvoiceNumberPrefix: "LZ"},
		SequenceNumber: 1,
	})

	assert.Equal(t, inv.Total.AmountMinor, inv.TaxableValue.AmountMinor+inv.GSTAmount.AmountMinor)
	assert.Equal(t, "pay_x", inv.RazorpayPaymentID)
	assert.Equal(t, "order_x", inv.RazorpayOrderID)
	assert.Equal(t, "subo_abc", inv.SubscriptionOrderID)
	assert.Equal(t, "LZ/2026-27/000001", inv.InvoiceNumber)
}

// A shop lookup failure must not block the payment — the invoice still issues,
// just without buyer details. Losing a seller's subscription because their shop
// row is missing would be far worse than an invoice with a blank buyer block.
func TestBuildInvoiceToleratesEmptyShop(t *testing.T) {
	paidAt := time.Now()
	inv := BuildInvoice(BuildInvoiceInput{
		Order: domain.SubscriptionOrder{
			ID: "subo_abc", UserID: "adm_x",
			Price: domain.INR(100000), GSTAmount: domain.INR(15254),
			PaidAt: &paidAt,
		},
		Plan:           domain.SubscriptionPlan{Name: "1 Month", DurationDays: 30},
		Shop:           domain.ShopDetails{}, // lookup failed
		Profile:        domain.CompanyBillingProfile{LegalName: "Locazar", InvoiceNumberPrefix: "LZ"},
		SequenceNumber: 7,
	})

	assert.Empty(t, inv.BuyerName)
	assert.Empty(t, inv.BuyerGSTIN)
	assert.Equal(t, int64(100000), inv.Total.AmountMinor)
	assert.NotEmpty(t, inv.InvoiceNumber)
}
