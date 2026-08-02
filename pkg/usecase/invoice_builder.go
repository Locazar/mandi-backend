package usecase

import (
	"strings"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// paymentMethodRazorpay is the only method that currently produces an invoice.
const paymentMethodRazorpay = "razorpay"

// BuildInvoiceInput is everything needed to snapshot an invoice. Every field is
// read at issue time and frozen onto the invoice.
type BuildInvoiceInput struct {
	Order          domain.SubscriptionOrder
	Plan           domain.SubscriptionPlan
	Shop           domain.ShopDetails
	Profile        domain.CompanyBillingProfile
	SequenceNumber int
}

// BuildInvoice assembles the immutable invoice snapshot. It is pure — no I/O,
// no clock dependency beyond the PaidAt fallback — so the whole of an invoice's
// content is testable without a database.
func BuildInvoice(in BuildInvoiceInput) domain.Invoice {
	paidAt := time.Now()
	if in.Order.PaidAt != nil {
		paidAt = *in.Order.PaidAt
	}

	// Taxable value is derived by subtraction rather than recomputed from the
	// rate. Recomputing would round independently and could disagree with the
	// stored GST by a paisa, breaking the ck_subscription_invoices_total_split
	// database CHECK.
	taxable := domain.Money{
		AmountMinor: in.Order.Price.AmountMinor - in.Order.GSTAmount.AmountMinor,
		Currency:    in.Order.Price.Currency,
	}

	financialYear := domain.FinancialYear(paidAt)

	prefix := in.Profile.InvoiceNumberPrefix
	if prefix == "" {
		prefix = "LZ"
	}

	paymentID := ""
	if in.Order.RazorpayPaymentID != nil {
		paymentID = *in.Order.RazorpayPaymentID
	}

	return domain.Invoice{
		ID:                  domain.NewID(domain.PrefixInvoice),
		SubscriptionOrderID: in.Order.ID,
		UserID:              in.Order.UserID,

		InvoiceNumber:  domain.FormatInvoiceNumber(prefix, financialYear, in.SequenceNumber),
		FinancialYear:  financialYear,
		SequenceNumber: in.SequenceNumber,
		InvoiceDate:    paidAt,

		SellerLegalName: in.Profile.LegalName,
		SellerGSTIN:     in.Profile.GSTIN,
		SellerPAN:       in.Profile.PAN,
		SellerAddress:   in.Profile.AddressBlock(),
		SellerSACCode:   in.Profile.SACCode,
		PlaceOfSupply:   in.Profile.PlaceOfSupply(),

		BuyerShopID:      in.Shop.ID,
		BuyerName:        in.Shop.ShopName,
		BuyerContactName: in.Shop.OwnerName,
		BuyerAddress:     shopAddressBlock(in.Shop),
		BuyerGSTIN:       buyerGSTIN(in.Shop),

		PlanName:        in.Plan.Name,
		PlanDescription: in.Plan.Description,
		PeriodStart:     paidAt,
		PeriodEnd:       paidAt.AddDate(0, 0, int(in.Plan.DurationDays)),

		Total:              in.Order.Price,
		TaxableValue:       taxable,
		GSTAmount:          in.Order.GSTAmount,
		GSTRateBasisPoints: in.Order.GSTRateBasisPoints,

		RazorpayPaymentID: paymentID,
		RazorpayOrderID:   in.Order.RazorpayOrderID,
		PaymentMethod:     paymentMethodRazorpay,
		PaidAt:            paidAt,
	}
}

// buyerGSTIN returns the seller's GSTIN only when they actually registered with
// one. Sellers who onboarded with PAN/Aadhaar/licence have a different number in
// Document_Value, and printing that in a GSTIN field would be wrong.
func buyerGSTIN(shop domain.ShopDetails) string {
	if shop.Document_Type == domain.ShopDocGST {
		return strings.TrimSpace(shop.Document_Value)
	}
	return ""
}

// shopAddressBlock renders the buyer's postal address, skipping blank parts.
func shopAddressBlock(shop domain.ShopDetails) string {
	cityLine := strings.TrimSpace(shop.City + " " + shop.Pincode)

	stateLine := shop.State
	if shop.Country != "" {
		if stateLine != "" {
			stateLine += ", " + shop.Country
		} else {
			stateLine = shop.Country
		}
	}

	lines := make([]string, 0, 4)
	for _, l := range []string{shop.AddressLine1, shop.AddressLine2, cityLine, stateLine} {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, l)
		}
	}
	return strings.Join(lines, "\n")
}
