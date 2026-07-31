package domain

import (
	"strings"
	"time"
)

// CompanyBillingProfile is the admin-managed singleton holding the issuer
// details printed on every invoice. It mirrors SubscriptionGSTConfig: one row,
// edited from the admin portal. Every value is snapshotted onto each invoice at
// issue time, so editing this never rewrites an issued invoice.
type CompanyBillingProfile struct {
	ID                  string    `json:"id" gorm:"primaryKey;type:varchar(32)"`
	LegalName           string    `json:"legal_name" gorm:"size:255;not null"`
	GSTIN               string    `json:"gstin" gorm:"size:20"`
	PAN                 string    `json:"pan" gorm:"size:20"`
	AddressLine1        string    `json:"address_line1" gorm:"size:255"`
	AddressLine2        string    `json:"address_line2" gorm:"size:255"`
	City                string    `json:"city" gorm:"size:100"`
	State               string    `json:"state" gorm:"size:100"`
	StateCode           string    `json:"state_code" gorm:"size:4"`
	Pincode             string    `json:"pincode" gorm:"size:12"`
	Country             string    `json:"country" gorm:"size:100"`
	SACCode             string    `json:"sac_code" gorm:"size:12"`
	LogoObjectKey       string    `json:"logo_object_key" gorm:"size:512"`
	InvoiceNumberPrefix string    `json:"invoice_number_prefix" gorm:"size:12;not null;default:'LZ'"`
	FooterNote          string    `json:"footer_note" gorm:"type:text"`
	Jurisdiction        string    `json:"jurisdiction" gorm:"size:100"`
	UpdatedAt           time.Time `json:"updated_at" gorm:"autoUpdateTime"`
}

func (CompanyBillingProfile) TableName() string { return "company_billing_profile" }

// PlaceOfSupply renders "Karnataka (29)", or just the state when no state code
// is configured, or "" when neither is set.
func (p CompanyBillingProfile) PlaceOfSupply() string {
	if p.State == "" {
		return ""
	}
	if p.StateCode == "" {
		return p.State
	}
	return p.State + " (" + p.StateCode + ")"
}

// AddressBlock renders the issuer's postal address as printable lines, skipping
// blank components so no stray separators or empty lines appear.
func (p CompanyBillingProfile) AddressBlock() string {
	cityLine := strings.TrimSpace(p.City + " " + p.Pincode)

	stateLine := p.State
	if p.Country != "" {
		if stateLine != "" {
			stateLine += ", " + p.Country
		} else {
			stateLine = p.Country
		}
	}

	lines := make([]string, 0, 4)
	for _, l := range []string{p.AddressLine1, p.AddressLine2, cityLine, stateLine} {
		if strings.TrimSpace(l) != "" {
			lines = append(lines, l)
		}
	}
	return strings.Join(lines, "\n")
}

// Invoice is an issued GST tax invoice for one paid subscription order.
//
// Every printed value is snapshotted at issue time. A later shop rename, GST
// rate change, or billing-profile edit must never alter an issued invoice —
// that is a statutory requirement, not a preference. Rows are insert-only.
type Invoice struct {
	ID                  string `json:"id" gorm:"primaryKey;type:varchar(32)"`
	SubscriptionOrderID string `json:"subscription_order_id" gorm:"type:varchar(32);uniqueIndex;not null"`
	// UserID is the subscriber's id. For sellers this is an admins.id, NOT a
	// users.id — see migration 000009_relax_subscription_user_fk. Deliberately
	// carries no foreign key for that reason.
	UserID string `json:"user_id" gorm:"type:varchar(32);not null;index"`

	InvoiceNumber  string    `json:"invoice_number" gorm:"size:64;uniqueIndex;not null"`
	FinancialYear  string    `json:"financial_year" gorm:"size:9;not null;index"`
	SequenceNumber int       `json:"sequence_number" gorm:"not null"`
	InvoiceDate    time.Time `json:"invoice_date" gorm:"not null"`

	// Issuer snapshot (from CompanyBillingProfile at issue time).
	SellerLegalName string `json:"seller_legal_name" gorm:"size:255;not null"`
	SellerGSTIN     string `json:"seller_gstin" gorm:"size:20"`
	SellerPAN       string `json:"seller_pan" gorm:"size:20"`
	SellerAddress   string `json:"seller_address" gorm:"type:text"`
	SellerSACCode   string `json:"seller_sac_code" gorm:"size:12"`
	PlaceOfSupply   string `json:"place_of_supply" gorm:"size:128"`

	// Buyer snapshot (from ShopDetails at issue time). BuyerGSTIN is empty for
	// sellers who onboarded with PAN/Aadhaar/licence instead of GSTIN.
	BuyerShopID      string `json:"buyer_shop_id" gorm:"type:varchar(32)"`
	BuyerName        string `json:"buyer_name" gorm:"size:255"`
	BuyerContactName string `json:"buyer_contact_name" gorm:"size:255"`
	BuyerAddress     string `json:"buyer_address" gorm:"type:text"`
	BuyerGSTIN       string `json:"buyer_gstin" gorm:"size:20"`

	// Line snapshot.
	PlanName        string    `json:"plan_name" gorm:"size:255"`
	PlanDescription string    `json:"plan_description" gorm:"type:text"`
	PeriodStart     time.Time `json:"period_start"`
	PeriodEnd       time.Time `json:"period_end"`

	// Money. Prices are GST-inclusive: Total = TaxableValue + GSTAmount.
	Total              Money `json:"total" gorm:"embedded;embeddedPrefix:total_"`
	TaxableValue       Money `json:"taxable_value" gorm:"embedded;embeddedPrefix:taxable_value_"`
	GSTAmount          Money `json:"gst_amount" gorm:"embedded;embeddedPrefix:gst_amount_"`
	GSTRateBasisPoints int   `json:"gst_rate_basis_points" gorm:"not null;default:0"`

	// Payment snapshot.
	RazorpayPaymentID string    `json:"razorpay_payment_id" gorm:"size:255"`
	RazorpayOrderID   string    `json:"razorpay_order_id" gorm:"size:255"`
	PaymentMethod     string    `json:"payment_method" gorm:"size:32"`
	PaidAt            time.Time `json:"paid_at"`

	// PDF cache. Empty PDFObjectKey means "not rendered yet" — the download
	// path re-renders from this snapshot and backfills the key.
	PDFObjectKey   string     `json:"pdf_object_key" gorm:"size:512"`
	PDFGeneratedAt *time.Time `json:"pdf_generated_at"`
	CreatedAt      time.Time  `json:"created_at" gorm:"autoCreateTime"`
}

func (Invoice) TableName() string { return "subscription_invoices" }

// CompanyBillingProfileID is the fixed primary key of the singleton issuer
// profile row seeded by migration 000034. Declared here so the repository and
// usecase layers share one definition instead of each keeping a copy that can
// drift.
const CompanyBillingProfileID = "cbp_default"

// FileName is the download filename. Slashes in the invoice number are path
// separators and must not survive into a filename.
func (i Invoice) FileName() string {
	return strings.ReplaceAll(i.InvoiceNumber, "/", "-") + ".pdf"
}

// InvoiceNumberSequence is the per-financial-year counter backing gapless
// invoice numbering. Allocated with a single upsert-returning statement.
type InvoiceNumberSequence struct {
	FinancialYear string `gorm:"primaryKey;size:9"`
	LastSequence  int    `gorm:"not null;default:0"`
}

func (InvoiceNumberSequence) TableName() string { return "invoice_number_sequences" }
