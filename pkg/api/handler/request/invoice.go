package request

// UpdateCompanyBillingProfileRequest is the admin-editable issuer profile
// printed on every invoice. LegalName is the only hard requirement — a tax
// invoice without a legal entity name is meaningless.
type UpdateCompanyBillingProfileRequest struct {
	LegalName           string `json:"legal_name" binding:"required"`
	GSTIN               string `json:"gstin"`
	PAN                 string `json:"pan"`
	AddressLine1        string `json:"address_line1"`
	AddressLine2        string `json:"address_line2"`
	City                string `json:"city"`
	State               string `json:"state"`
	StateCode           string `json:"state_code"`
	Pincode             string `json:"pincode"`
	Country             string `json:"country"`
	SACCode             string `json:"sac_code"`
	LogoObjectKey       string `json:"logo_object_key"`
	InvoiceNumberPrefix string `json:"invoice_number_prefix"`
	FooterNote          string `json:"footer_note"`
	Jurisdiction        string `json:"jurisdiction"`
}
