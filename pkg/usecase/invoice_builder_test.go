package usecase

import (
	"testing"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func baseBuildInput() BuildInvoiceInput {
	ist := time.FixedZone("IST", 5*60*60+30*60)
	paidAt := time.Date(2026, 7, 30, 16, 18, 0, 0, ist)

	return BuildInvoiceInput{
		Order: domain.SubscriptionOrder{
			ID:                 "subo_abc",
			UserID:             "adm_seller1",
			PlanID:             "subp_3m",
			Price:              domain.INR(149900),
			GSTRateBasisPoints: 1800,
			GSTAmount:          domain.INR(22866),
			RazorpayOrderID:    "order_QxK8pJ4YdN2sc9",
			RazorpayPaymentID:  strPtr("pay_QxK9mR2VbL7ta1"),
			PaidAt:             &paidAt,
		},
		Plan: domain.SubscriptionPlan{
			Name:         "3 Months",
			DurationDays: 90,
			Description:  "Marketplace listing & support services",
		},
		Shop: domain.ShopDetails{
			ID:             "shp_8f2ac91b",
			ShopName:       "Sharma Kirana Store",
			OwnerName:      "Rakesh Sharma",
			AddressLine1:   "Shop 7, Bannerghatta Road",
			City:           "Bengaluru",
			State:          "Karnataka",
			Pincode:        "560076",
			Document_Type:  domain.ShopDocGST,
			Document_Value: "29ABCDE1234F2Z5",
		},
		Profile: domain.CompanyBillingProfile{
			LegalName:           "Locazar Technologies Pvt. Ltd.",
			GSTIN:               "29AABCL1234M1Z7",
			PAN:                 "AABCL1234M",
			AddressLine1:        "No. 42, 3rd Floor, 12th Main",
			AddressLine2:        "Indiranagar",
			City:                "Bengaluru",
			State:               "Karnataka",
			StateCode:           "29",
			Pincode:             "560038",
			Country:             "India",
			SACCode:             "998599",
			InvoiceNumberPrefix: "LZ",
			FooterNote:          "Custom footer set by an admin.",
			Jurisdiction:        "Bengaluru",
		},
		SequenceNumber: 42,
	}
}

func strPtr(s string) *string { return &s }

func TestBuildInvoiceAmounts(t *testing.T) {
	inv := BuildInvoice(baseBuildInput())

	assert.Equal(t, int64(149900), inv.Total.AmountMinor)
	assert.Equal(t, int64(22866), inv.GSTAmount.AmountMinor)
	// Taxable value is derived by subtraction, never recomputed from the rate —
	// so total always reconciles exactly and satisfies the DB CHECK.
	assert.Equal(t, int64(127034), inv.TaxableValue.AmountMinor)
	assert.Equal(t,
		inv.Total.AmountMinor,
		inv.TaxableValue.AmountMinor+inv.GSTAmount.AmountMinor,
	)
	assert.Equal(t, "INR", inv.TaxableValue.Currency)
	assert.Equal(t, 1800, inv.GSTRateBasisPoints)
}

func TestBuildInvoiceNumbering(t *testing.T) {
	inv := BuildInvoice(baseBuildInput())
	assert.Equal(t, "LZ/2026-27/000042", inv.InvoiceNumber)
	assert.Equal(t, "2026-27", inv.FinancialYear)
	assert.Equal(t, 42, inv.SequenceNumber)
}

func TestBuildInvoiceSnapshots(t *testing.T) {
	inv := BuildInvoice(baseBuildInput())

	assert.Equal(t, "Locazar Technologies Pvt. Ltd.", inv.SellerLegalName)
	assert.Equal(t, "29AABCL1234M1Z7", inv.SellerGSTIN)
	assert.Equal(t, "998599", inv.SellerSACCode)
	assert.Equal(t, "Karnataka (29)", inv.PlaceOfSupply)
	// FooterNote/Jurisdiction must be snapshotted from the profile at issue
	// time, same as every other issuer field — otherwise an admin editing
	// them in the portal has no effect on any invoice, issued or future.
	assert.Equal(t, "Custom footer set by an admin.", inv.SellerFooterNote)
	assert.Equal(t, "Bengaluru", inv.SellerJurisdiction)

	assert.Equal(t, "shp_8f2ac91b", inv.BuyerShopID)
	assert.Equal(t, "Sharma Kirana Store", inv.BuyerName)
	assert.Equal(t, "Rakesh Sharma", inv.BuyerContactName)
	assert.Equal(t, "29ABCDE1234F2Z5", inv.BuyerGSTIN)

	assert.Equal(t, "3 Months", inv.PlanName)
	assert.Equal(t, "razorpay", inv.PaymentMethod)
	assert.Equal(t, "pay_QxK9mR2VbL7ta1", inv.RazorpayPaymentID)
	assert.Equal(t, "subo_abc", inv.SubscriptionOrderID)
	assert.Equal(t, "adm_seller1", inv.UserID)

	// Period runs from payment for the plan's duration.
	assert.Equal(t, 90, int(inv.PeriodEnd.Sub(inv.PeriodStart).Hours()/24))
}

func TestBuildInvoiceOmitsBuyerGSTINForNonGSTDocument(t *testing.T) {
	in := baseBuildInput()
	in.Shop.Document_Type = domain.ShopDocPAN
	in.Shop.Document_Value = "ABCDE1234F"

	inv := BuildInvoice(in)

	// A PAN number must never be printed in the GSTIN field.
	assert.Empty(t, inv.BuyerGSTIN)
}

func TestBuildInvoiceHandlesMissingPaidAt(t *testing.T) {
	in := baseBuildInput()
	in.Order.PaidAt = nil

	inv := BuildInvoice(in)

	// Falls back to Now() rather than producing a zero-time invoice date.
	assert.False(t, inv.InvoiceDate.IsZero())
	assert.False(t, inv.PaidAt.IsZero())
}

func TestBuildInvoiceZeroGSTRate(t *testing.T) {
	in := baseBuildInput()
	in.Order.GSTRateBasisPoints = 0
	in.Order.GSTAmount = domain.INR(0)

	inv := BuildInvoice(in)

	assert.Equal(t, int64(149900), inv.TaxableValue.AmountMinor)
	assert.Equal(t, int64(0), inv.GSTAmount.AmountMinor)
}

// TestBuildInvoiceTaxableValueIsSubtractedNotRecomputed pins the rule the
// builder exists to enforce, with numbers that actually distinguish the two
// approaches.
//
// The reconciliation assertion in TestBuildInvoiceAmounts cannot catch a
// regression on its own: TaxableValue is *defined* as Total-GST, so
// Total == Taxable+GST holds by construction no matter what. This test asserts
// exact expected values instead, chosen so that the realistic regression — a Go
// recompute like `total*10000/(10000+rate)`, which truncates because integer
// division truncates — produces a visibly different number.
func TestBuildInvoiceTaxableValueIsSubtractedNotRecomputed(t *testing.T) {
	tests := map[string]struct {
		totalMinor      int64
		gstMinor        int64
		rateBasisPoints int
		wantTaxable     int64
		wantNaiveTrunc  int64 // what a truncating recompute would wrongly produce
	}{
		"3-month plan at 18%": {149900, 22866, 1800, 127034, 127033},
		"round total at 18%":  {100000, 15254, 1800, 84746, 84745},
		"small amount at 18%": {9900, 1510, 1800, 8390, 8389},
		"lower rate at 5%":    {149900, 7138, 500, 142762, 142761},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			in := baseBuildInput()
			in.Order.Price = domain.INR(tc.totalMinor)
			in.Order.GSTAmount = domain.INR(tc.gstMinor)
			in.Order.GSTRateBasisPoints = tc.rateBasisPoints

			inv := BuildInvoice(in)

			require.Equal(t, tc.wantTaxable, inv.TaxableValue.AmountMinor)
			// Guard the guard: if these two ever coincide the case has stopped
			// distinguishing the approaches and needs new numbers.
			require.NotEqual(t, tc.wantNaiveTrunc, tc.wantTaxable,
				"test case no longer distinguishes subtraction from a truncating recompute")
			require.NotEqual(t, tc.wantNaiveTrunc, inv.TaxableValue.AmountMinor,
				"taxable value looks recomputed-and-truncated rather than subtracted")

			// And the split must still reconcile exactly, or the DB CHECK rejects it.
			require.Equal(t, tc.totalMinor,
				inv.TaxableValue.AmountMinor+inv.GSTAmount.AmountMinor)
		})
	}
}

// TestBuildInvoiceDefaultsPrefixWhenProfileBlank covers the fallback branch that
// every other test bypasses by always setting InvoiceNumberPrefix.
func TestBuildInvoiceDefaultsPrefixWhenProfileBlank(t *testing.T) {
	in := baseBuildInput()
	in.Profile.InvoiceNumberPrefix = ""

	inv := BuildInvoice(in)

	require.Equal(t, "LZ/2026-27/000042", inv.InvoiceNumber)
}

// TestBuildInvoiceHandlesNilPaymentID covers an order that reached the builder
// without a Razorpay payment id (the field is a nullable pointer).
func TestBuildInvoiceHandlesNilPaymentID(t *testing.T) {
	in := baseBuildInput()
	in.Order.RazorpayPaymentID = nil

	inv := BuildInvoice(in)

	require.Empty(t, inv.RazorpayPaymentID)
	require.NotEmpty(t, inv.InvoiceNumber)
}
