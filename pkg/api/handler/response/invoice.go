package response

// InvoiceListItem is one invoice in a listing. Amounts are minor units (paise);
// GST is already included in TotalMinor because prices are tax-inclusive.
type InvoiceListItem struct {
	ID                 string `json:"id"`
	InvoiceNumber      string `json:"invoice_number"`
	InvoiceDate        string `json:"invoice_date"`
	PlanName           string `json:"plan_name"`
	BuyerName          string `json:"buyer_name,omitempty"`
	TotalMinor         int64  `json:"total_minor"`
	TaxableValueMinor  int64  `json:"taxable_value_minor"`
	GSTAmountMinor     int64  `json:"gst_amount_minor"`
	GSTRateBasisPoints int    `json:"gst_rate_basis_points"`
	Currency           string `json:"currency"`
	RazorpayPaymentID  string `json:"razorpay_payment_id"`
	PaidAt             string `json:"paid_at"`
}

// InvoiceDownloadResponse carries a short-lived presigned URL. The PDF itself
// is never streamed through the API — the client fetches the URL directly.
type InvoiceDownloadResponse struct {
	DownloadURL string `json:"download_url"`
	FileName    string `json:"file_name"`
	ExpiresAt   string `json:"expires_at"`
}
