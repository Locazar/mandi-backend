package notification

import (
	"testing"
)

// TestEnquiryRecipientResolver_NewEnquiry verifies that a newly created enquiry
// (status "pending" / empty) always routes to the seller.
func TestEnquiryRecipientResolver_NewEnquiry(t *testing.T) {
	cases := []struct {
		name   string
		status string
	}{
		{"empty status", ""},
		{"pending", "pending"},
		{"new", "new"},
		{"open", "open"},
		{"unknown", "some_unrecognised_status"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			doc := map[string]interface{}{
				"status":   tc.status,
				"sellerId": "seller_123",
				"userId":   "user_456",
			}
			notifyUser, notifySeller := enquiryRecipientResolver(doc)
			if notifyUser {
				t.Errorf("status=%q: expected notifyUser=false, got true", tc.status)
			}
			if !notifySeller {
				t.Errorf("status=%q: expected notifySeller=true, got false", tc.status)
			}
		})
	}
}

// TestEnquiryRecipientResolver_SellerStates verifies seller-facing statuses.
func TestEnquiryRecipientResolver_SellerStates(t *testing.T) {
	sellerStatuses := []string{
		"pending_seller_price",
		"pending_seller_final",
		"seller_final_update",
	}
	for _, status := range sellerStatuses {
		t.Run(status, func(t *testing.T) {
			doc := map[string]interface{}{"status": status}
			_, notifySeller := enquiryRecipientResolver(doc)
			if !notifySeller {
				t.Errorf("status=%q: expected notifySeller=true", status)
			}
		})
	}
}

// TestEnquiryRecipientResolver_UserStates verifies user-facing statuses.
func TestEnquiryRecipientResolver_UserStates(t *testing.T) {
	userStatuses := []string{
		"pending_customer_price",
		"pending_customer_final",
	}
	for _, status := range userStatuses {
		t.Run(status, func(t *testing.T) {
			doc := map[string]interface{}{"status": status}
			notifyUser, _ := enquiryRecipientResolver(doc)
			if !notifyUser {
				t.Errorf("status=%q: expected notifyUser=true", status)
			}
		})
	}
}

// TestEnquiryRecipientResolver_CompletedAccepted verifies completed_accepted routing.
func TestEnquiryRecipientResolver_CompletedAccepted(t *testing.T) {
	cases := []struct {
		acceptedBy string
		wantUser   bool
		wantSeller bool
	}{
		{"seller", true, false},
		{"client", false, true},
		{"customer", false, true},
		{"buyer", false, true},
	}
	for _, tc := range cases {
		t.Run("accepted_by="+tc.acceptedBy, func(t *testing.T) {
			doc := map[string]interface{}{
				"status":     "completed_accepted",
				"acceptedBy": tc.acceptedBy,
			}
			notifyUser, notifySeller := enquiryRecipientResolver(doc)
			if notifyUser != tc.wantUser {
				t.Errorf("acceptedBy=%q: notifyUser got %v want %v", tc.acceptedBy, notifyUser, tc.wantUser)
			}
			if notifySeller != tc.wantSeller {
				t.Errorf("acceptedBy=%q: notifySeller got %v want %v", tc.acceptedBy, notifySeller, tc.wantSeller)
			}
		})
	}
}
