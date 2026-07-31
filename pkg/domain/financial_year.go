package domain

import (
	"fmt"
	"time"

	// Embeds the IANA timezone database in the binary so Asia/Kolkata resolves
	// even in a scratch/distroless container that ships no tzdata.
	_ "time/tzdata"
)

// istLocation is Indian Standard Time. Invoice numbering keys off the Indian
// financial year, so the 1 April boundary must be evaluated in IST — using UTC
// misfiles every invoice issued in the last 5h30m of 31 March.
var istLocation = func() *time.Location {
	loc, err := time.LoadLocation("Asia/Kolkata")
	if err != nil {
		return time.FixedZone("IST", 5*60*60+30*60)
	}
	return loc
}()

// FinancialYear returns the Indian financial year containing t, formatted
// "2026-27". The year runs 1 April – 31 March, evaluated in IST.
func FinancialYear(t time.Time) string {
	ist := t.In(istLocation)
	startYear := ist.Year()
	if ist.Month() < time.April {
		startYear--
	}
	return fmt.Sprintf("%d-%02d", startYear, (startYear+1)%100)
}

// FormatInvoiceNumber builds the printed invoice number, e.g.
// "LZ/2026-27/000042". Sequences wider than six digits are not truncated.
func FormatInvoiceNumber(prefix, financialYear string, sequence int) string {
	return fmt.Sprintf("%s/%s/%06d", prefix, financialYear, sequence)
}
