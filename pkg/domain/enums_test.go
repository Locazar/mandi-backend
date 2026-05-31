package domain

import "testing"

func TestEnumIsValid(t *testing.T) {
	cases := []struct {
		name  string
		valid bool
		ok    bool
	}{
		{"user-type-user", UserTypeUser.IsValid(), true},
		{"user-type-bad", UserType("root").IsValid(), false},
		{"shop-type-bad", ShopType("xyz").IsValid(), false},
		{"offer-type-percentage", OfferTypePercentage.IsValid(), true},
		{"field-type-dropdown", FieldTypeDropdown.IsValid(), true},
		{"address-type-bad", AddressType("nowhere").IsValid(), false},
		{"consent-type-terms", ConsentTypeTerms.IsValid(), true},
		{"notif-status-bad", NotificationStatus("weird").IsValid(), false},
		{"subscription-status-paid", SubscriptionStatusPaid.IsValid(), true},
	}
	for _, c := range cases {
		if c.valid != c.ok {
			t.Errorf("%s: got %v want %v", c.name, c.valid, c.ok)
		}
	}
}
