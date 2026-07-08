package domain

import (
	"fmt"
)

// CurrencyINR is the default currency for the marketplace.
const CurrencyINR = "INR"

// Money is the value type for all monetary amounts. Amounts are stored as
// integer minor units (paise for INR) to avoid floating-point rounding errors,
// alongside an ISO 4217 currency code.
//
// Persisted via GORM as an embedded struct, producing two columns per field,
// e.g. `embedded;embeddedPrefix:order_total_` →
// `order_total_amount_minor BIGINT` + `order_total_currency CHAR(3)`.
type Money struct {
	AmountMinor int64  `json:"amount_minor" gorm:"column:amount_minor;not null;default:0"`
	Currency    string `json:"currency" gorm:"column:currency;type:char(3);not null;default:'INR'"`
}

// NewMoney builds a Money in the given currency (ISO 4217).
func NewMoney(amountMinor int64, currency string) Money {
	return Money{AmountMinor: amountMinor, Currency: currency}
}

// INR builds a Money in Indian Rupees from an amount in paise.
func INR(amountMinor int64) Money {
	return Money{AmountMinor: amountMinor, Currency: CurrencyINR}
}

// Add returns the sum of two amounts, erroring on currency mismatch.
func (m Money) Add(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("money: cannot add %s to %s", other.Currency, m.Currency)
	}
	return Money{AmountMinor: m.AmountMinor + other.AmountMinor, Currency: m.Currency}, nil
}

// Sub returns the difference of two amounts, erroring on currency mismatch.
func (m Money) Sub(other Money) (Money, error) {
	if m.Currency != other.Currency {
		return Money{}, fmt.Errorf("money: cannot subtract %s from %s", other.Currency, m.Currency)
	}
	return Money{AmountMinor: m.AmountMinor - other.AmountMinor, Currency: m.Currency}, nil
}

// InclusiveTaxPortion returns the tax already contained in a tax-inclusive
// amount, given a rate in basis points (1800 = 18.00%). Because the amount is
// tax-inclusive, the tax component is amount * rate / (100% + rate); with the
// rate in basis points that is amount * rate / (10000 + rate). The result is
// rounded to the nearest minor unit and carries the same currency.
//
// A non-positive rate yields a zero amount (no tax).
func (m Money) InclusiveTaxPortion(rateBasisPoints int) Money {
	if rateBasisPoints <= 0 {
		return Money{AmountMinor: 0, Currency: m.Currency}
	}
	denom := int64(10000) + int64(rateBasisPoints)
	tax := (m.AmountMinor*int64(rateBasisPoints) + denom/2) / denom
	return Money{AmountMinor: tax, Currency: m.Currency}
}

// IsZero reports whether the amount is exactly zero.
func (m Money) IsZero() bool { return m.AmountMinor == 0 }

// IsNegative reports whether the amount is below zero.
func (m Money) IsNegative() bool { return m.AmountMinor < 0 }

// Validate checks the currency is a 3-letter uppercase ISO 4217 code.
func (m Money) Validate() error {
	if len(m.Currency) != 3 {
		return fmt.Errorf("money: currency %q must be a 3-letter ISO 4217 code", m.Currency)
	}
	for _, r := range m.Currency {
		if r < 'A' || r > 'Z' {
			return fmt.Errorf("money: currency %q must be uppercase letters", m.Currency)
		}
	}
	return nil
}

// String renders the amount in major units with two decimal places, prefixed
// by the currency code, e.g. "INR 1234.56".
func (m Money) String() string {
	sign := ""
	v := m.AmountMinor
	if v < 0 {
		sign = "-"
		v = -v
	}
	return fmt.Sprintf("%s %s%d.%02d", m.Currency, sign, v/100, v%100)
}
