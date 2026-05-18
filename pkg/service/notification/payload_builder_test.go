package notification

import (
	"strings"
	"testing"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// ─────────────────────────────────────────────────────────────────────────────
// helpers
// ─────────────────────────────────────────────────────────────────────────────

func newPB() *PayloadBuilder { return NewPayloadBuilder() }

// parsedEvent builds a minimal ParsedFirestoreEvent for testing BuildPayload.
func parsedEvent(docID string, fields map[string]interface{}) *domain.ParsedFirestoreEvent {
	return &domain.ParsedFirestoreEvent{
		DocumentID:   docID,
		DocumentPath: "enquiry/" + docID,
		UpdateTime:   "2026-01-01T00:00:00Z",
		OldFields:    map[string]interface{}{},
		NewFields:    fields,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// BuildPayload — field extraction
// ─────────────────────────────────────────────────────────────────────────────

func TestBuildPayload_BasicFieldExtraction(t *testing.T) {
	event := parsedEvent("enquiry-1", map[string]interface{}{
		"userId":   "user-123",
		"sellerId": "seller-456",
		"queryId":  "q-789",
		"status":   "pending_seller_price",
	})
	change := domain.FieldChange{FieldName: "status", OldValue: "new", NewValue: "pending_seller_price"}
	pb := newPB()
	payload := pb.BuildPayload(event, []domain.FieldChange{change})

	if payload.DocumentID != "enquiry-1" {
		t.Errorf("DocumentID: got %q want %q", payload.DocumentID, "enquiry-1")
	}
	if payload.UserID != "user-123" {
		t.Errorf("UserID: got %q want %q", payload.UserID, "user-123")
	}
	if payload.SellerID != "seller-456" {
		t.Errorf("SellerID: got %q want %q", payload.SellerID, "seller-456")
	}
	if payload.EnquiryID != "q-789" {
		t.Errorf("EnquiryID: got %q want %q", payload.EnquiryID, "q-789")
	}
	if !strings.Contains(payload.ActionURL, "q-789") {
		t.Errorf("ActionURL should contain enquiryID; got %q", payload.ActionURL)
	}
}

func TestBuildPayload_SellerID_AlternativeFields(t *testing.T) {
	cases := []struct {
		name   string
		fields map[string]interface{}
	}{
		{"shopId", map[string]interface{}{"shopId": "shop-999"}},
		{"shop_id", map[string]interface{}{"shop_id": "shop-999"}},
		{"seller_id", map[string]interface{}{"seller_id": "shop-999"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			event := parsedEvent("d1", tc.fields)
			payload := newPB().BuildPayload(event, nil)
			if payload.SellerID != "shop-999" {
				t.Errorf("%s: SellerID got %q want %q", tc.name, payload.SellerID, "shop-999")
			}
		})
	}
}

func TestBuildPayload_SellerID_FallbackToAssignedTo(t *testing.T) {
	// When sellerId/shopId absent, SellerID falls back to assignedTo.
	event := parsedEvent("d1", map[string]interface{}{"assignedTo": "support-agent-1"})
	payload := newPB().BuildPayload(event, nil)
	if payload.SellerID != "support-agent-1" {
		t.Errorf("SellerID fallback: got %q want %q", payload.SellerID, "support-agent-1")
	}
}

func TestBuildPayload_ActionURL_FallsBackToDocumentID(t *testing.T) {
	event := parsedEvent("doc-xyz", map[string]interface{}{})
	payload := newPB().BuildPayload(event, nil)
	if !strings.Contains(payload.ActionURL, "doc-xyz") {
		t.Errorf("ActionURL fallback should contain documentID; got %q", payload.ActionURL)
	}
}

func TestBuildPayload_Timestamp(t *testing.T) {
	event := parsedEvent("d1", map[string]interface{}{})
	payload := newPB().BuildPayload(event, nil)
	if payload.Timestamp == "" {
		t.Error("Timestamp must not be empty")
	}
}

func TestBuildPayload_ChangedFields(t *testing.T) {
	event := parsedEvent("d1", map[string]interface{}{"status": "resolved"})
	changes := []domain.FieldChange{
		{FieldName: "status", OldValue: "in_progress", NewValue: "resolved"},
		{FieldName: "closedAt", OldValue: nil, NewValue: "2026-01-01"},
	}
	payload := newPB().BuildPayload(event, changes)
	if payload.ChangeCount != 2 {
		t.Errorf("ChangeCount: got %d want 2", payload.ChangeCount)
	}
	if len(payload.ChangedFields) != 2 {
		t.Errorf("ChangedFields len: got %d want 2", len(payload.ChangedFields))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// buildStatusNotification — all 19 status cases
// ─────────────────────────────────────────────────────────────────────────────

func TestBuildStatusNotification_AllStatuses(t *testing.T) {
	pb := newPB()

	type want struct {
		title       string
		bodyContain string
	}

	cases := []struct {
		status string
		fields map[string]interface{}
		want   want
	}{
		{
			"pending_seller_price",
			map[string]interface{}{"askQuantity": "100"},
			want{"Price Request Pending", "100"},
		},
		{
			"pending_seller_price",
			map[string]interface{}{"availability": "limited"},
			want{"Price Request Pending", "limited"},
		},
		{
			"pending_customer_price",
			map[string]interface{}{"sellerInitialPrice": "250"},
			want{"Seller Price Shared", "250"},
		},
		{
			"pending_customer_price",
			nil,
			want{"Seller Price Shared", "initial price"},
		},
		{
			"pending_seller_final",
			map[string]interface{}{"customerNegotiatedPrice": "220"},
			want{"Customer Counter Offer Received", "220"},
		},
		{
			"pending_customer_final",
			map[string]interface{}{"sellerFinalPrice": "235"},
			want{"Seller Final Price Shared", "235"},
		},
		{
			"seller_final_update",
			map[string]interface{}{"customerFinalResponse": "230"},
			want{"Customer Final Response Received", "230"},
		},
		{
			"completed_accepted",
			map[string]interface{}{"acceptedPrice": "230", "acceptedBy": "seller"},
			want{"Deal Accepted", "230"},
		},
		{
			"completed_rejected",
			map[string]interface{}{"rejectedBy": "client"},
			want{"Deal Rejected", "client"},
		},
		{
			"completed_rejected",
			map[string]interface{}{"acceptedBy": "seller"},
			want{"Deal Rejected", "seller"}, // falls back to acceptedBy
		},
		{
			"new",
			nil,
			want{"Enquiry Created", "created"},
		},
		{
			"in_progress",
			nil,
			want{"Enquiry In Progress", ""},
		},
		{
			"inprogress",
			nil,
			want{"Enquiry In Progress", ""},
		},
		{
			"resolved",
			nil,
			want{"Enquiry Resolved", "resolved"},
		},
		{
			"closed",
			nil,
			want{"Enquiry Closed", "closed"},
		},
		{
			"rejected",
			nil,
			want{"Enquiry Rejected", "rejected"},
		},
		{
			"cancelled",
			nil,
			want{"Enquiry Cancelled", "cancelled"},
		},
		{
			"expired",
			nil,
			want{"Enquiry Expired", "expired"},
		},
		{
			"on_hold",
			map[string]interface{}{"onHoldReason": "stock pending"},
			want{"Enquiry On Hold", "stock pending"},
		},
		{
			"on_hold",
			nil,
			want{"Enquiry On Hold", "on hold"},
		},
		{
			"reopened",
			nil,
			want{"Enquiry Reopened", "reopened"},
		},
		{
			"counter_offer",
			nil,
			want{"Counter Offer Received", "counter offer"},
		},
		{
			"dispute",
			map[string]interface{}{"disputeReason": "product mismatch"},
			want{"Dispute Raised", "product mismatch"},
		},
		{
			"dispute",
			nil,
			want{"Dispute Raised", "support"},
		},
		{
			"dispute_resolved",
			nil,
			want{"Dispute Resolved", "resolved"},
		},
	}

	for _, tc := range cases {
		name := tc.status
		if tc.fields != nil {
			for k := range tc.fields {
				name += "+" + k
			}
		}
		t.Run(name, func(t *testing.T) {
			fields := map[string]interface{}{"status": tc.status}
			for k, v := range tc.fields {
				fields[k] = v
			}
			change := domain.FieldChange{FieldName: "status", OldValue: "new", NewValue: tc.status}
			title, body := pb.buildStatusNotification(change, fields)
			if title != tc.want.title {
				t.Errorf("title: got %q want %q", title, tc.want.title)
			}
			if tc.want.bodyContain != "" && !strings.Contains(strings.ToLower(body), strings.ToLower(tc.want.bodyContain)) {
				t.Errorf("body %q should contain %q", body, tc.want.bodyContain)
			}
		})
	}
}

func TestBuildStatusNotification_Default_ContainsOldAndNewStatus(t *testing.T) {
	pb := newPB()
	change := domain.FieldChange{FieldName: "status", OldValue: "open", NewValue: "custom_unknown_status"}
	title, body := pb.buildStatusNotification(change, map[string]interface{}{})
	if title != "Enquiry Status Updated" {
		t.Errorf("default title: got %q", title)
	}
	if !strings.Contains(body, "open") || !strings.Contains(body, "custom_unknown_status") {
		t.Errorf("default body should mention old and new status; got %q", body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// buildResponseNotification
// ─────────────────────────────────────────────────────────────────────────────

func TestBuildResponseNotification_WithPreview(t *testing.T) {
	pb := newPB()
	change := domain.FieldChange{FieldName: "response", NewValue: "We can offer you a better deal soon."}
	title, body := pb.buildResponseNotification(change)
	if title != "New Response" {
		t.Errorf("title: got %q", title)
	}
	if !strings.Contains(body, "We can offer") {
		t.Errorf("body should contain preview; got %q", body)
	}
}

func TestBuildResponseNotification_PreviewTruncatedAt100(t *testing.T) {
	pb := newPB()
	long := strings.Repeat("x", 200)
	change := domain.FieldChange{FieldName: "response", NewValue: long}
	_, body := pb.buildResponseNotification(change)
	// Preview should be at most 100 chars of content + "New response: " prefix + "..."
	if len(body) > len("New response: ")+103 {
		t.Errorf("body too long (preview not truncated): len=%d", len(body))
	}
	if !strings.HasSuffix(body, "...") {
		t.Errorf("truncated preview should end with '...'; got %q", body)
	}
}

func TestBuildResponseNotification_NilNewValue(t *testing.T) {
	pb := newPB()
	change := domain.FieldChange{FieldName: "response", NewValue: nil}
	title, body := pb.buildResponseNotification(change)
	if title == "" || body == "" {
		t.Error("title and body must be non-empty even with nil NewValue")
	}
}

func TestBuildResponseNotification_EmptyNewValue(t *testing.T) {
	pb := newPB()
	change := domain.FieldChange{FieldName: "response", NewValue: ""}
	_, body := pb.buildResponseNotification(change)
	if body == "" {
		t.Error("body must not be empty")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// buildAssignmentNotification
// ─────────────────────────────────────────────────────────────────────────────

func TestBuildAssignmentNotification_WithAgent(t *testing.T) {
	pb := newPB()
	event := parsedEvent("d1", map[string]interface{}{"assignedToName": "Alice"})
	change := domain.FieldChange{FieldName: "assignedTo", NewValue: "agent-1"}
	title, body := pb.buildAssignmentNotification(change, event)
	if title != "Enquiry Assigned" {
		t.Errorf("title: got %q", title)
	}
	if !strings.Contains(body, "Alice") {
		t.Errorf("body should mention agent name; got %q", body)
	}
}

func TestBuildAssignmentNotification_UnassignedEmptyValue(t *testing.T) {
	pb := newPB()
	event := parsedEvent("d1", map[string]interface{}{})
	change := domain.FieldChange{FieldName: "assignedTo", NewValue: ""}
	_, body := pb.buildAssignmentNotification(change, event)
	if !strings.Contains(strings.ToLower(body), "no longer") {
		t.Errorf("unassigned body should mention removal; got %q", body)
	}
}

func TestBuildAssignmentNotification_NoAgentName(t *testing.T) {
	pb := newPB()
	event := parsedEvent("d1", map[string]interface{}{})
	change := domain.FieldChange{FieldName: "assignedTo", NewValue: "agent-42"}
	_, body := pb.buildAssignmentNotification(change, event)
	if !strings.Contains(strings.ToLower(body), "agent") {
		t.Errorf("should fall back to generic agent text; got %q", body)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// generateNotificationContent — priority ordering
// ─────────────────────────────────────────────────────────────────────────────

func TestGenerateNotificationContent_StatusTakesPriority(t *testing.T) {
	pb := newPB()
	event := parsedEvent("d1", map[string]interface{}{"status": "resolved"})
	changes := []domain.FieldChange{
		{FieldName: "status", OldValue: "in_progress", NewValue: "resolved"},
		{FieldName: "assignedTo", OldValue: nil, NewValue: "agent-1"},
	}
	title, _ := pb.generateNotificationContent(event, changes)
	if title != "Enquiry Resolved" {
		t.Errorf("status change should take priority; got title=%q", title)
	}
}

func TestGenerateNotificationContent_AssignmentFallback(t *testing.T) {
	pb := newPB()
	event := parsedEvent("d1", map[string]interface{}{"assignedToName": "Bob"})
	changes := []domain.FieldChange{
		{FieldName: "assignedTo", OldValue: nil, NewValue: "agent-5"},
	}
	title, body := pb.generateNotificationContent(event, changes)
	if title != "Enquiry Assigned" {
		t.Errorf("expected 'Enquiry Assigned'; got %q", title)
	}
	if !strings.Contains(body, "Bob") {
		t.Errorf("body should contain agent name; got %q", body)
	}
}

func TestGenerateNotificationContent_ResponseFallback(t *testing.T) {
	pb := newPB()
	event := parsedEvent("d1", map[string]interface{}{})
	changes := []domain.FieldChange{
		{FieldName: "response", NewValue: "Updated."},
	}
	title, _ := pb.generateNotificationContent(event, changes)
	if title != "New Response" {
		t.Errorf("expected 'New Response'; got %q", title)
	}
}

func TestGenerateNotificationContent_GenericMultiChange(t *testing.T) {
	pb := newPB()
	event := parsedEvent("d1", map[string]interface{}{})
	changes := []domain.FieldChange{
		{FieldName: "notes", OldValue: nil, NewValue: "note1"},
		{FieldName: "priority", OldValue: "low", NewValue: "high"},
	}
	title, body := pb.generateNotificationContent(event, changes)
	if title != "Enquiry Updated" {
		t.Errorf("expected generic title; got %q", title)
	}
	if body == "" {
		t.Error("body must not be empty")
	}
}

func TestGenerateNotificationContent_NoChanges(t *testing.T) {
	pb := newPB()
	event := parsedEvent("d1", map[string]interface{}{})
	title, body := pb.generateNotificationContent(event, nil)
	if title == "" || body == "" {
		t.Error("zero-change notification must still produce non-empty title/body")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ValidatePayload
// ─────────────────────────────────────────────────────────────────────────────

func TestValidatePayload_Valid(t *testing.T) {
	p := &domain.NotificationPayload{
		DocumentID: "d1",
		Title:      "Hello",
		Body:       "World",
		Timestamp:  "2026-01-01T00:00:00Z",
	}
	if err := ValidatePayload(p); err != nil {
		t.Errorf("valid payload should pass validation; err: %v", err)
	}
}

func TestValidatePayload_Nil(t *testing.T) {
	if err := ValidatePayload(nil); err == nil {
		t.Error("nil payload should fail validation")
	}
}

func TestValidatePayload_MissingDocumentID(t *testing.T) {
	p := &domain.NotificationPayload{Title: "T", Body: "B", Timestamp: "ts"}
	if err := ValidatePayload(p); err == nil {
		t.Error("missing DocumentID should fail validation")
	}
}

func TestValidatePayload_MissingTitle(t *testing.T) {
	p := &domain.NotificationPayload{DocumentID: "d1", Body: "B", Timestamp: "ts"}
	if err := ValidatePayload(p); err == nil {
		t.Error("missing Title should fail validation")
	}
}

func TestValidatePayload_MissingBody(t *testing.T) {
	p := &domain.NotificationPayload{DocumentID: "d1", Title: "T", Timestamp: "ts"}
	if err := ValidatePayload(p); err == nil {
		t.Error("missing Body should fail validation")
	}
}

func TestValidatePayload_MissingTimestamp_SetsDefault(t *testing.T) {
	p := &domain.NotificationPayload{DocumentID: "d1", Title: "T", Body: "B"}
	if err := ValidatePayload(p); err != nil {
		t.Errorf("missing Timestamp should be auto-filled; got err: %v", err)
	}
	if p.Timestamp == "" {
		t.Error("Timestamp should be set automatically when absent")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// formatFieldName
// ─────────────────────────────────────────────────────────────────────────────

func TestFormatFieldName(t *testing.T) {
	cases := []struct{ in, want string }{
		{"status", "Status"},
		{"acceptedBy", "Accepted By"},
		{"sellerFinalPrice", "Seller Final Price"},
		{"customerNegotiatedPrice", "Customer Negotiated Price"},
		{"id", "Id"},
	}
	for _, tc := range cases {
		got := formatFieldName(tc.in)
		if got != tc.want {
			t.Errorf("formatFieldName(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// firstNonEmptyField
// ─────────────────────────────────────────────────────────────────────────────

func TestFirstNonEmptyField(t *testing.T) {
	fields := map[string]interface{}{
		"a": "",
		"b": "found",
		"c": "other",
	}
	got := firstNonEmptyField(fields, "a", "b", "c")
	if got != "found" {
		t.Errorf("got %q want %q", got, "found")
	}
}

func TestFirstNonEmptyField_AllAbsent(t *testing.T) {
	got := firstNonEmptyField(map[string]interface{}{}, "x", "y")
	if got != "" {
		t.Errorf("expected empty; got %q", got)
	}
}

func TestFirstNonEmptyField_NonStringValue_Skipped(t *testing.T) {
	fields := map[string]interface{}{
		"num":  100, // not a string
		"name": "Alice",
	}
	got := firstNonEmptyField(fields, "num", "name")
	if got != "Alice" {
		t.Errorf("non-string value should be skipped; got %q", got)
	}
}
