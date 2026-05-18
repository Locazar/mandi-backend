package notification

import (
	"errors"
	"strings"
	"testing"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// ─────────────────────────────────────────────────────────────────────────────
// resolveEnquiryRecipients — routing logic (mirrors watcher_rules.go)
// ─────────────────────────────────────────────────────────────────────────────

func TestResolveEnquiryRecipients_SellerFacingStatuses(t *testing.T) {
	for _, status := range []string{"pending_seller_price", "pending_seller_final", "seller_final_update"} {
		t.Run(status, func(t *testing.T) {
			notifyUser, notifySeller := resolveEnquiryRecipients(map[string]interface{}{"status": status})
			if notifyUser {
				t.Errorf("status=%q: notifyUser must be false (seller-facing state)", status)
			}
			if !notifySeller {
				t.Errorf("status=%q: notifySeller must be true (seller-facing state)", status)
			}
		})
	}
}

func TestResolveEnquiryRecipients_BuyerFacingStatuses(t *testing.T) {
	for _, status := range []string{"pending_customer_price", "pending_customer_final"} {
		t.Run(status, func(t *testing.T) {
			notifyUser, notifySeller := resolveEnquiryRecipients(map[string]interface{}{"status": status})
			if !notifyUser {
				t.Errorf("status=%q: notifyUser must be true (buyer-facing state)", status)
			}
			if notifySeller {
				t.Errorf("status=%q: notifySeller must be false (buyer-facing state)", status)
			}
		})
	}
}

func TestResolveEnquiryRecipients_CompletedAccepted_ActorRouting(t *testing.T) {
	cases := []struct {
		acceptedBy string
		wantUser   bool
		wantSeller bool
	}{
		{"seller", true, false},
		{"client", false, true},
		{"customer", false, true},
		{"buyer", false, true},
		{"user", false, true},
		{"", true, true},      // unknown → both
		{"admin", true, true}, // unrecognised → both
	}
	for _, tc := range cases {
		t.Run("acceptedBy="+tc.acceptedBy, func(t *testing.T) {
			notifyUser, notifySeller := resolveEnquiryRecipients(map[string]interface{}{
				"status":     "completed_accepted",
				"acceptedBy": tc.acceptedBy,
			})
			if notifyUser != tc.wantUser || notifySeller != tc.wantSeller {
				t.Errorf("acceptedBy=%q: got notifyUser=%v notifySeller=%v, want %v/%v",
					tc.acceptedBy, notifyUser, notifySeller, tc.wantUser, tc.wantSeller)
			}
		})
	}
}

func TestResolveEnquiryRecipients_CompletedRejected_ActorRouting(t *testing.T) {
	cases := []struct {
		rejectedBy string
		acceptedBy string
		wantUser   bool
		wantSeller bool
	}{
		{"seller", "", true, false},
		{"client", "", false, true},
		{"customer", "", false, true},
		{"buyer", "", false, true},
		// No rejectedBy → fall back to acceptedBy
		{"", "seller", true, false},
		{"", "client", false, true},
		// Neither known → both
		{"", "", true, true},
	}
	for _, tc := range cases {
		label := "rejectedBy=" + tc.rejectedBy + "/acceptedBy=" + tc.acceptedBy
		t.Run(label, func(t *testing.T) {
			doc := map[string]interface{}{
				"status":     "completed_rejected",
				"rejectedBy": tc.rejectedBy,
				"acceptedBy": tc.acceptedBy,
			}
			notifyUser, notifySeller := resolveEnquiryRecipients(doc)
			if notifyUser != tc.wantUser || notifySeller != tc.wantSeller {
				t.Errorf("%s: got notifyUser=%v notifySeller=%v, want %v/%v",
					label, notifyUser, notifySeller, tc.wantUser, tc.wantSeller)
			}
		})
	}
}

func TestResolveEnquiryRecipients_AdminAndTerminalStatuses_BothParties(t *testing.T) {
	adminStatuses := []string{
		"in_progress", "on_hold", "resolved", "closed", "cancelled",
		"expired", "reopened", "counter_offer",
		"dispute", "dispute_resolved",
	}
	for _, status := range adminStatuses {
		t.Run(status, func(t *testing.T) {
			notifyUser, notifySeller := resolveEnquiryRecipients(map[string]interface{}{"status": status})
			if !notifyUser || !notifySeller {
				t.Errorf("status=%q: expected both=true, got notifyUser=%v notifySeller=%v",
					status, notifyUser, notifySeller)
			}
		})
	}
}

func TestResolveEnquiryRecipients_NewOrEmptyStatus_SellerOnly(t *testing.T) {
	cases := []string{"", "new", "pending", "unknown_xyz"}
	for _, status := range cases {
		t.Run("status="+status, func(t *testing.T) {
			notifyUser, notifySeller := resolveEnquiryRecipients(map[string]interface{}{"status": status})
			if notifyUser {
				t.Errorf("status=%q: expected notifyUser=false (new enquiry → seller only)", status)
			}
			if !notifySeller {
				t.Errorf("status=%q: expected notifySeller=true (new enquiry → seller only)", status)
			}
		})
	}
}

func TestResolveEnquiryRecipients_CaseInsensitive(t *testing.T) {
	cases := []struct {
		status       string
		expectSeller bool
		expectUser   bool
	}{
		{"PENDING_SELLER_PRICE", true, false},
		{"pending_seller_price  ", true, false}, // leading/trailing spaces
		{"Pending_Customer_Final", false, true},
	}
	for _, tc := range cases {
		t.Run(tc.status, func(t *testing.T) {
			notifyUser, notifySeller := resolveEnquiryRecipients(map[string]interface{}{"status": tc.status})
			if notifyUser != tc.expectUser {
				t.Errorf("status=%q: notifyUser got %v want %v", tc.status, notifyUser, tc.expectUser)
			}
			if notifySeller != tc.expectSeller {
				t.Errorf("status=%q: notifySeller got %v want %v", tc.status, notifySeller, tc.expectSeller)
			}
		})
	}
}

func TestResolveEnquiryRecipients_StatusFromAlternativeKey(t *testing.T) {
	// resolveEnquiryRecipients reads from "status" key; verified that no
	// alternative keys are consumed (i.e., finalStatus does NOT drive routing).
	notifyUser, notifySeller := resolveEnquiryRecipients(map[string]interface{}{
		"finalStatus": "completed_accepted",
		// "status" key absent → fallback to seller-only
	})
	if notifyUser {
		t.Error("expected notifyUser=false when status key absent")
	}
	if !notifySeller {
		t.Error("expected notifySeller=true when status key absent")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// chunkStrings
// ─────────────────────────────────────────────────────────────────────────────

func TestChunkStrings_LessThanChunkSize(t *testing.T) {
	in := []string{"a", "b", "c"}
	chunks := chunkStrings(in, 500)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
	if len(chunks[0]) != 3 {
		t.Errorf("expected chunk of 3, got %d", len(chunks[0]))
	}
}

func TestChunkStrings_ExactlyChunkSize(t *testing.T) {
	in := make([]string, 500)
	for i := range in {
		in[i] = "token"
	}
	chunks := chunkStrings(in, 500)
	if len(chunks) != 1 {
		t.Fatalf("expected 1 chunk, got %d", len(chunks))
	}
}

func TestChunkStrings_MultipleChunks(t *testing.T) {
	in := make([]string, 1100)
	for i := range in {
		in[i] = "token"
	}
	chunks := chunkStrings(in, 500)
	if len(chunks) != 3 {
		t.Fatalf("expected 3 chunks (500+500+100), got %d", len(chunks))
	}
	if len(chunks[2]) != 100 {
		t.Errorf("last chunk should be 100, got %d", len(chunks[2]))
	}
}

func TestChunkStrings_EmptyInput(t *testing.T) {
	chunks := chunkStrings([]string{}, 500)
	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks for empty input, got %d", len(chunks))
	}
}

func TestChunkStrings_NilInput(t *testing.T) {
	chunks := chunkStrings(nil, 500)
	if len(chunks) != 0 {
		t.Errorf("expected 0 chunks for nil input, got %d", len(chunks))
	}
}

func TestChunkStrings_ZeroOrNegativeSize_UsesDefault(t *testing.T) {
	in := make([]string, 10)
	// size=0 should not panic; falls back to fcmMaxTokensPerBatch
	chunks := chunkStrings(in, 0)
	if len(chunks) == 0 {
		t.Error("expected at least 1 chunk")
	}
}

func TestChunkStrings_PreservesOrder(t *testing.T) {
	in := []string{"a", "b", "c", "d", "e"}
	chunks := chunkStrings(in, 2)
	flat := []string{}
	for _, c := range chunks {
		flat = append(flat, c...)
	}
	for i, v := range in {
		if flat[i] != v {
			t.Errorf("order mismatch at idx %d: got %q want %q", i, flat[i], v)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// isUnregisteredTokenError
// ─────────────────────────────────────────────────────────────────────────────

func TestIsUnregisteredTokenError_NilError(t *testing.T) {
	if isUnregisteredTokenError(nil) {
		t.Error("nil error should not be unregistered")
	}
}

func TestIsUnregisteredTokenError_UnregisteredKeywords(t *testing.T) {
	keywords := []string{
		"registration-token-not-registered",
		"Unregistered",
		"UNREGISTERED",
		"invalid registration",
		"NotRegistered",
		"requested entity was not found",
	}
	for _, kw := range keywords {
		t.Run(kw, func(t *testing.T) {
			err := errors.New(kw)
			if !isUnregisteredTokenError(err) {
				t.Errorf("error %q should be recognised as unregistered token error", kw)
			}
		})
	}
}

func TestIsUnregisteredTokenError_OtherErrors(t *testing.T) {
	others := []string{
		"network timeout",
		"internal server error",
		"quota exceeded",
		"authentication failed",
	}
	for _, msg := range others {
		t.Run(msg, func(t *testing.T) {
			if isUnregisteredTokenError(errors.New(msg)) {
				t.Errorf("error %q should NOT be recognised as unregistered token error", msg)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// firstNonEmptyString
// ─────────────────────────────────────────────────────────────────────────────

func TestFirstNonEmptyString_ReturnsFirstNonEmpty(t *testing.T) {
	fields := map[string]interface{}{
		"a": "",
		"b": nil,
		"c": "found",
		"d": "also valid",
	}
	got := firstNonEmptyString(fields, "a", "b", "c", "d")
	if got != "found" {
		t.Errorf("got %q, want %q", got, "found")
	}
}

func TestFirstNonEmptyString_AllEmpty(t *testing.T) {
	fields := map[string]interface{}{"a": "", "b": nil}
	got := firstNonEmptyString(fields, "a", "b", "absent")
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestFirstNonEmptyString_KeyAbsent(t *testing.T) {
	got := firstNonEmptyString(map[string]interface{}{}, "x", "y")
	if got != "" {
		t.Errorf("expected empty for absent keys, got %q", got)
	}
}

func TestFirstNonEmptyString_WhitespaceOnly(t *testing.T) {
	fields := map[string]interface{}{"a": "   ", "b": "value"}
	got := firstNonEmptyString(fields, "a", "b")
	// "   " trims to "" so should skip to "b"
	if got != "value" {
		t.Errorf("whitespace-only value should be skipped; got %q", got)
	}
}

func TestFirstNonEmptyString_NilSentinel(t *testing.T) {
	fields := map[string]interface{}{"a": "<nil>", "b": "real"}
	got := firstNonEmptyString(fields, "a", "b")
	// "<nil>" string should be skipped
	if got != "real" {
		t.Errorf("'<nil>' sentinel should be skipped; got %q", got)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// buildRecipientDedupeKey
// ─────────────────────────────────────────────────────────────────────────────

func TestBuildRecipientDedupeKey_Deterministic(t *testing.T) {
	pl := &domain.NotificationPayload{
		DocumentID:    "doc1",
		Timestamp:     "2026-01-01T00:00:00Z",
		ChangedFields: []string{"status", "acceptedBy"},
	}
	r := &domain.NotificationRecipient{UserID: "user1", Type: "user"}
	k1 := buildRecipientDedupeKey(pl, r)
	k2 := buildRecipientDedupeKey(pl, r)
	if k1 != k2 {
		t.Errorf("dedupeKey is not deterministic: %q vs %q", k1, k2)
	}
	if !strings.HasPrefix(k1, "dedupe_") {
		t.Errorf("dedupeKey should start with 'dedupe_', got %q", k1)
	}
}

func TestBuildRecipientDedupeKey_ChangedFieldsOrderInvariant(t *testing.T) {
	// ChangedFields are sorted internally, so order should not matter.
	pl1 := &domain.NotificationPayload{
		DocumentID: "doc1", Timestamp: "ts",
		ChangedFields: []string{"status", "acceptedBy"},
	}
	pl2 := &domain.NotificationPayload{
		DocumentID: "doc1", Timestamp: "ts",
		ChangedFields: []string{"acceptedBy", "status"},
	}
	r := &domain.NotificationRecipient{UserID: "u1", Type: "user"}
	if buildRecipientDedupeKey(pl1, r) != buildRecipientDedupeKey(pl2, r) {
		t.Error("dedupeKey should be invariant to ChangedFields order")
	}
}

func TestBuildRecipientDedupeKey_DifferentRecipientsDifferentKeys(t *testing.T) {
	pl := &domain.NotificationPayload{DocumentID: "doc1", Timestamp: "ts"}
	r1 := &domain.NotificationRecipient{UserID: "user1", Type: "user"}
	r2 := &domain.NotificationRecipient{UserID: "seller1", Type: "seller"}
	if buildRecipientDedupeKey(pl, r1) == buildRecipientDedupeKey(pl, r2) {
		t.Error("different recipients must produce different dedupeKeys")
	}
}

func TestBuildRecipientDedupeKey_DifferentDocumentsDifferentKeys(t *testing.T) {
	r := &domain.NotificationRecipient{UserID: "u1", Type: "user"}
	pl1 := &domain.NotificationPayload{DocumentID: "doc1", Timestamp: "ts"}
	pl2 := &domain.NotificationPayload{DocumentID: "doc2", Timestamp: "ts"}
	if buildRecipientDedupeKey(pl1, r) == buildRecipientDedupeKey(pl2, r) {
		t.Error("different documents must produce different dedupeKeys")
	}
}
