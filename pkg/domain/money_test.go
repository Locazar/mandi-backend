package domain

import "testing"

func TestINRConstructor(t *testing.T) {
	m := INR(12345)
	if m.AmountMinor != 12345 || m.Currency != CurrencyINR {
		t.Fatalf("INR(12345) = %+v", m)
	}
}

func TestMoneyAdd(t *testing.T) {
	got, err := INR(100).Add(INR(250))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AmountMinor != 350 {
		t.Fatalf("100 + 250 = %d, want 350", got.AmountMinor)
	}
}

func TestMoneySub(t *testing.T) {
	got, err := INR(500).Sub(INR(150))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AmountMinor != 350 {
		t.Fatalf("500 - 150 = %d, want 350", got.AmountMinor)
	}
}

func TestMoneyCurrencyMismatch(t *testing.T) {
	a := NewMoney(100, "INR")
	b := NewMoney(100, "USD")
	if _, err := a.Add(b); err == nil {
		t.Fatal("expected error adding mismatched currencies")
	}
}

func TestMoneyValidate(t *testing.T) {
	if err := INR(0).Validate(); err != nil {
		t.Fatalf("zero INR should be valid: %v", err)
	}
	if err := (Money{AmountMinor: 100, Currency: "rupee"}).Validate(); err == nil {
		t.Fatal("non-ISO currency should be invalid")
	}
	if err := (Money{AmountMinor: 100, Currency: ""}).Validate(); err == nil {
		t.Fatal("empty currency should be invalid")
	}
}

func TestMoneyString(t *testing.T) {
	if got := INR(123456).String(); got != "INR 1234.56" {
		t.Fatalf("String() = %q, want %q", got, "INR 1234.56")
	}
	if got := INR(-500).String(); got != "INR -5.00" {
		t.Fatalf("String() = %q, want %q", got, "INR -5.00")
	}
}

func TestMoneyHelpers(t *testing.T) {
	if !INR(0).IsZero() {
		t.Fatal("INR(0).IsZero() should be true")
	}
	if !INR(-1).IsNegative() {
		t.Fatal("INR(-1).IsNegative() should be true")
	}
	if INR(1).IsNegative() {
		t.Fatal("INR(1).IsNegative() should be false")
	}
}
