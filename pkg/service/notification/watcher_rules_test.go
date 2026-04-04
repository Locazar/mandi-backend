package notification

import (
	"strings"
	"testing"

	"github.com/rohit221990/mandi-backend/pkg/domain"
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
		{"user", false, true}, // "user" actor variant → notify seller
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

// TestEnquiryRecipientResolver_CompletedAccepted_UnknownActor verifies that
// when acceptedBy is empty or unrecognised both parties get a notification.
func TestEnquiryRecipientResolver_CompletedAccepted_UnknownActor(t *testing.T) {
	cases := []struct{ actor string }{{""}, {"admin"}, {"system"}}
	for _, tc := range cases {
		t.Run("actor="+tc.actor, func(t *testing.T) {
			doc := map[string]interface{}{
				"status":     "completed_accepted",
				"acceptedBy": tc.actor,
			}
			notifyUser, notifySeller := enquiryRecipientResolver(doc)
			if !notifyUser || !notifySeller {
				t.Errorf("actor=%q: expected both=true, got notifyUser=%v notifySeller=%v", tc.actor, notifyUser, notifySeller)
			}
		})
	}
}

// TestEnquiryRecipientResolver_CompletedRejected verifies completed_rejected routing
// using the rejectedBy field (not acceptedBy).
func TestEnquiryRecipientResolver_CompletedRejected(t *testing.T) {
	cases := []struct {
		rejectedBy string
		wantUser   bool
		wantSeller bool
	}{
		{"seller", true, false}, // seller rejected → notify buyer
		{"client", false, true}, // buyer rejected → notify seller
		{"customer", false, true},
		{"buyer", false, true},
		{"", true, true}, // unknown actor → notify both
	}
	for _, tc := range cases {
		t.Run("rejected_by="+tc.rejectedBy, func(t *testing.T) {
			doc := map[string]interface{}{
				"status":     "completed_rejected",
				"rejectedBy": tc.rejectedBy,
			}
			notifyUser, notifySeller := enquiryRecipientResolver(doc)
			if notifyUser != tc.wantUser {
				t.Errorf("rejectedBy=%q: notifyUser got %v want %v", tc.rejectedBy, notifyUser, tc.wantUser)
			}
			if notifySeller != tc.wantSeller {
				t.Errorf("rejectedBy=%q: notifySeller got %v want %v", tc.rejectedBy, notifySeller, tc.wantSeller)
			}
		})
	}
}

// TestEnquiryRecipientResolver_CompletedRejected_FallbackToAcceptedBy verifies that
// when rejectedBy is empty the resolver falls back to acceptedBy.
func TestEnquiryRecipientResolver_CompletedRejected_FallbackToAcceptedBy(t *testing.T) {
	doc := map[string]interface{}{
		"status":     "completed_rejected",
		"acceptedBy": "seller",
		// rejectedBy intentionally absent → should fall back to acceptedBy
	}
	notifyUser, notifySeller := enquiryRecipientResolver(doc)
	if !notifyUser {
		t.Error("expected notifyUser=true (seller acted via acceptedBy fallback)")
	}
	if notifySeller {
		t.Error("expected notifySeller=false (seller acted, so buyer gets the push)")
	}
}

// TestEnquiryRecipientResolver_AdminStates verifies that terminal / admin states
// always notify BOTH parties.
func TestEnquiryRecipientResolver_AdminStates(t *testing.T) {
	adminStatuses := []string{
		"in_progress",
		"on_hold",
		"resolved",
		"closed",
		"cancelled",
		"expired",
		"reopened",
		"counter_offer",
		"dispute",
		"dispute_resolved",
	}
	for _, status := range adminStatuses {
		t.Run(status, func(t *testing.T) {
			doc := map[string]interface{}{"status": status}
			notifyUser, notifySeller := enquiryRecipientResolver(doc)
			if !notifyUser {
				t.Errorf("status=%q: expected notifyUser=true", status)
			}
			if !notifySeller {
				t.Errorf("status=%q: expected notifySeller=true", status)
			}
		})
	}
}

// TestEnquiryRecipientResolver_SellerStates_VerifyUserNotNotified ensures seller-facing
// statuses do NOT trigger a user notification.
func TestEnquiryRecipientResolver_SellerStates_VerifyUserNotNotified(t *testing.T) {
	for _, status := range []string{"pending_seller_price", "pending_seller_final", "seller_final_update"} {
		t.Run(status, func(t *testing.T) {
			doc := map[string]interface{}{"status": status}
			notifyUser, _ := enquiryRecipientResolver(doc)
			if notifyUser {
				t.Errorf("status=%q: expected notifyUser=false", status)
			}
		})
	}
}

// TestEnquiryRecipientResolver_UserStates_VerifySellerNotNotified ensures buyer-facing
// statuses do NOT trigger a seller notification.
func TestEnquiryRecipientResolver_UserStates_VerifySellerNotNotified(t *testing.T) {
	for _, status := range []string{"pending_customer_price", "pending_customer_final"} {
		t.Run(status, func(t *testing.T) {
			doc := map[string]interface{}{"status": status}
			_, notifySeller := enquiryRecipientResolver(doc)
			if notifySeller {
				t.Errorf("status=%q: expected notifySeller=false", status)
			}
		})
	}
}

// TestEnquiryRecipientResolver_CaseInsensitive verifies upper/mixed-case status is
// handled correctly (the resolver lowercases before switching).
func TestEnquiryRecipientResolver_CaseInsensitive(t *testing.T) {
	cases := []struct {
		status     string
		wantUser   bool
		wantSeller bool
	}{
		{"PENDING_SELLER_PRICE", false, true},
		{"Pending_Customer_Price", true, false},
		{"COMPLETED_ACCEPTED", true, true}, // unknown actor after case fold → both
	}
	for _, tc := range cases {
		t.Run(tc.status, func(t *testing.T) {
			doc := map[string]interface{}{"status": tc.status}
			notifyUser, notifySeller := enquiryRecipientResolver(doc)
			if notifyUser != tc.wantUser {
				t.Errorf("status=%q: notifyUser got %v want %v", tc.status, notifyUser, tc.wantUser)
			}
			if notifySeller != tc.wantSeller {
				t.Errorf("status=%q: notifySeller got %v want %v", tc.status, notifySeller, tc.wantSeller)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// enquiryMessageBuilder — message copy tests
// ─────────────────────────────────────────────────────────────────────────────

func TestEnquiryMessageBuilder_StatusChanges(t *testing.T) {
	pb := NewPayloadBuilder()
	cases := []struct {
		name        string
		status      string
		fields      map[string]interface{}
		wantTitle   string
		bodyContain string
	}{
		{
			"pending_seller_price with quantity",
			"pending_seller_price",
			map[string]interface{}{"askQuantity": "50"},
			"Price Request Pending", "50",
		},
		{
			"pending_customer_price with price",
			"pending_customer_price",
			map[string]interface{}{"sellerInitialPrice": "200"},
			"Seller Price Shared", "200",
		},
		{
			"pending_seller_final with negotiated price",
			"pending_seller_final",
			map[string]interface{}{"customerNegotiatedPrice": "180"},
			"Customer Counter Offer Received", "180",
		},
		{
			"pending_customer_final with final price",
			"pending_customer_final",
			map[string]interface{}{"sellerFinalPrice": "190"},
			"Seller Final Price Shared", "190",
		},
		{
			"seller_final_update with customer response",
			"seller_final_update",
			map[string]interface{}{"customerFinalResponse": "185"},
			"Customer Final Response Received", "185",
		},
		{
			"completed_accepted with price and actor",
			"completed_accepted",
			map[string]interface{}{"acceptedPrice": "185", "acceptedBy": "seller"},
			"Deal Accepted", "185",
		},
		{
			"completed_rejected with rejectedBy",
			"completed_rejected",
			map[string]interface{}{"rejectedBy": "client"},
			"Deal Rejected", "client",
		},
		{
			"resolved",
			"resolved",
			nil, "Enquiry Resolved", "resolved",
		},
		{
			"cancelled",
			"cancelled",
			nil, "Enquiry Cancelled", "cancelled",
		},
		{
			"on_hold with reason",
			"on_hold",
			map[string]interface{}{"onHoldReason": "awaiting stock"},
			"Enquiry On Hold", "awaiting stock",
		},
		{
			"dispute with reason",
			"dispute",
			map[string]interface{}{"disputeReason": "wrong item"},
			"Dispute Raised", "wrong item",
		},
		{
			"dispute_resolved",
			"dispute_resolved",
			nil, "Dispute Resolved", "resolved",
		},
		{
			"reopened",
			"reopened",
			nil, "Enquiry Reopened", "reopened",
		},
		{
			"counter_offer",
			"counter_offer",
			nil, "Counter Offer Received", "counter offer",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fields := map[string]interface{}{"status": tc.status}
			for k, v := range tc.fields {
				fields[k] = v
			}
			change := domain.FieldChange{FieldName: "status", OldValue: "new", NewValue: tc.status}
			title, body := pb.buildStatusNotification(change, fields)
			if title != tc.wantTitle {
				t.Errorf("title: got %q want %q", title, tc.wantTitle)
			}
			if !strings.Contains(strings.ToLower(body), strings.ToLower(tc.bodyContain)) {
				t.Errorf("body %q should contain %q", body, tc.bodyContain)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// enquiryMessageBuilder — field-change message copy
// ─────────────────────────────────────────────────────────────────────────────

func TestEnquiryMessageBuilder_FieldChanges(t *testing.T) {
	enqFields := map[string]interface{}{"sellerId": "s1", "userId": "u1"}

	cases := []struct {
		field       WatchFieldChange
		bodyContain string
	}{
		{WatchFieldChange{Field: "sellerInitialPrice", NewValue: "150"}, "150"},
		{WatchFieldChange{Field: "customerNegotiatedPrice", NewValue: "130"}, "130"},
		{WatchFieldChange{Field: "sellerFinalPrice", NewValue: "145"}, "145"},
		{WatchFieldChange{Field: "customerFinalResponse", NewValue: "140"}, "140"},
		{WatchFieldChange{Field: "acceptedBy", NewValue: "seller"}, "seller"},
		{WatchFieldChange{Field: "acceptedPrice", NewValue: "145"}, "145"},
		{WatchFieldChange{Field: "availability", NewValue: "available"}, "available"},
		{WatchFieldChange{Field: "availability", NewValue: "unavailable"}, "unavailable"},
	}

	for _, tc := range cases {
		t.Run("field="+tc.field.Field, func(t *testing.T) {
			_, body := enquiryMessageBuilder(enqFields, []WatchFieldChange{tc.field})
			if !strings.Contains(strings.ToLower(body), strings.ToLower(tc.bodyContain)) {
				t.Errorf("field %q body %q should contain %q", tc.field.Field, body, tc.bodyContain)
			}
		})
	}
}

// TestEnquiryMessageBuilder_FallbackBody verifies the default body is returned
// when no monitored field change matches.
func TestEnquiryMessageBuilder_FallbackBody(t *testing.T) {
	_, body := enquiryMessageBuilder(map[string]interface{}{}, []WatchFieldChange{})
	if body == "" {
		t.Error("expected non-empty fallback body")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// DefaultEnquiryRule — structural checks
// ─────────────────────────────────────────────────────────────────────────────

func TestDefaultEnquiryRule_ContainsRejectedBy(t *testing.T) {
	rule := DefaultEnquiryRule()
	found := false
	for _, f := range rule.MonitoredFields {
		if f == "rejectedBy" {
			found = true
			break
		}
	}
	if !found {
		t.Error("DefaultEnquiryRule.MonitoredFields must include 'rejectedBy' for completed_rejected routing")
	}
}

func TestDefaultEnquiryRule_HasRecipientResolver(t *testing.T) {
	rule := DefaultEnquiryRule()
	if rule.RecipientResolver == nil {
		t.Error("DefaultEnquiryRule must have a RecipientResolver set")
	}
}

func TestDefaultEnquiryRule_NotifyOnCreate(t *testing.T) {
	rule := DefaultEnquiryRule()
	if !rule.NotifyOnCreate {
		t.Error("DefaultEnquiryRule.NotifyOnCreate should be true — new enquiries must notify the seller")
	}
}
