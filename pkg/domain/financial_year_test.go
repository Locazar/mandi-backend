package domain

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFinancialYear(t *testing.T) {
	ist := time.FixedZone("IST", 5*60*60+30*60)

	tests := map[string]struct {
		in   time.Time
		want string
	}{
		"mid year IST":            {time.Date(2026, 7, 30, 12, 0, 0, 0, ist), "2026-27"},
		"last minute of FY IST":   {time.Date(2026, 3, 31, 23, 59, 59, 0, ist), "2025-26"},
		"first minute of FY IST":  {time.Date(2026, 4, 1, 0, 0, 0, 0, ist), "2026-27"},
		"january belongs to prev": {time.Date(2026, 1, 15, 9, 0, 0, 0, ist), "2025-26"},
		// 18:31 UTC on 31 Mar is 00:01 IST on 1 Apr — the case a naive UTC
		// implementation gets wrong.
		"UTC input crossing IST boundary": {time.Date(2026, 3, 31, 18, 31, 0, 0, time.UTC), "2026-27"},
		// 18:29 UTC on 31 Mar is 23:59 IST on 31 Mar — still the old FY.
		"UTC input just before boundary": {time.Date(2026, 3, 31, 18, 29, 0, 0, time.UTC), "2025-26"},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, tc.want, FinancialYear(tc.in))
		})
	}
}

func TestFormatInvoiceNumber(t *testing.T) {
	assert.Equal(t, "LZ/2026-27/000042", FormatInvoiceNumber("LZ", "2026-27", 42))
	assert.Equal(t, "LZ/2026-27/000001", FormatInvoiceNumber("LZ", "2026-27", 1))
	assert.Equal(t, "LZ/2026-27/123456", FormatInvoiceNumber("LZ", "2026-27", 123456))
	// Sequences beyond six digits must not be truncated.
	assert.Equal(t, "LZ/2026-27/1234567", FormatInvoiceNumber("LZ", "2026-27", 1234567))
}
