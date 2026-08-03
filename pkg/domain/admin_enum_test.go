package domain

import "testing"

func TestAdminStatusIsValid(t *testing.T) {
	valid := []AdminStatus{AdminStatusActive, AdminStatusInactive, AdminStatusSuspended}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("%q should be valid", s)
		}
	}
	if AdminStatus("deleted").IsValid() {
		t.Error("unknown status should be invalid")
	}
	if AdminStatus("").IsValid() {
		t.Error("empty status should be invalid")
	}
}

func TestVerificationStatusIsValid(t *testing.T) {
	if !VerificationStatusUnderReview.IsValid() {
		t.Error("under_review should be valid")
	}
	if VerificationStatusType("pending").IsValid() {
		t.Error("pending should be invalid")
	}
}

func TestShopStatusIsValid(t *testing.T) {
	valid := []ShopStatusType{ShopStatusActive, ShopStatusInactive, ShopStatusSuspended, ShopStatusUnderReview, ShopStatusRejected}
	for _, s := range valid {
		if !s.IsValid() {
			t.Errorf("%q should be valid", s)
		}
	}
	if ShopStatusType("pending_review").IsValid() {
		t.Error("pending_review should be invalid (use under_review)")
	}
}

func TestAdvertisementEnumsIsValid(t *testing.T) {
	if !AdvertisementStatusExpired.IsValid() || !AdvertisementPriorityHigh.IsValid() {
		t.Error("expected expired/high to be valid")
	}
	if AdvertisementStatus("paused").IsValid() {
		t.Error("paused should be invalid")
	}
	if AdvertisementPriority("urgent").IsValid() {
		t.Error("urgent should be invalid")
	}
}
