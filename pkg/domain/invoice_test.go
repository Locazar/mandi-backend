package domain

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInvoiceFileName(t *testing.T) {
	inv := Invoice{InvoiceNumber: "LZ/2026-27/000042"}
	// Slashes are path separators — they must not survive into a filename.
	assert.Equal(t, "LZ-2026-27-000042.pdf", inv.FileName())
}

func TestCompanyBillingProfilePlaceOfSupply(t *testing.T) {
	p := CompanyBillingProfile{State: "Karnataka", StateCode: "29"}
	assert.Equal(t, "Karnataka (29)", p.PlaceOfSupply())

	noCode := CompanyBillingProfile{State: "Karnataka"}
	assert.Equal(t, "Karnataka", noCode.PlaceOfSupply())

	empty := CompanyBillingProfile{}
	assert.Equal(t, "", empty.PlaceOfSupply())
}

func TestCompanyBillingProfileAddressBlock(t *testing.T) {
	p := CompanyBillingProfile{
		AddressLine1: "No. 42, 3rd Floor, 12th Main",
		AddressLine2: "Indiranagar",
		City:         "Bengaluru",
		State:        "Karnataka",
		Pincode:      "560038",
		Country:      "India",
	}
	assert.Equal(t,
		"No. 42, 3rd Floor, 12th Main\nIndiranagar\nBengaluru 560038\nKarnataka, India",
		p.AddressBlock(),
	)

	// Blank components must not leave stray separators or empty lines.
	sparse := CompanyBillingProfile{AddressLine1: "No. 42", City: "Bengaluru"}
	assert.Equal(t, "No. 42\nBengaluru", sparse.AddressBlock())
}
