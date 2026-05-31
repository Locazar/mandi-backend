package domain

import (
	"testing"
	"time"
)

func TestUserHygieneFields(t *testing.T) {
	u := User{
		DateOfBirth:   time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		EmailVerified: true,
		PhoneVerified: false,
	}
	if u.AgeYears(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)) != 36 {
		t.Fatalf("got %d", u.AgeYears(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)))
	}
}

func TestAddressPincodeString(t *testing.T) {
	a := Address{Pincode: "007012"}
	if a.Pincode != "007012" {
		t.Fatal("pincode must preserve leading zeros as string")
	}
}
