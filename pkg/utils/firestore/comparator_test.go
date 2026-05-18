package firestore

import (
	"os"
	"strings"
	"testing"
)

// ─────────────────────────────────────────────────────────────────────────────
// getDefaultMonitoredFields — structural integrity
// ─────────────────────────────────────────────────────────────────────────────

func TestDefaultMonitoredFields_ContainsNegotiationFields(t *testing.T) {
	required := []string{
		"status", "finalStatus",
		"acceptedBy", "rejectedBy", "acceptedPrice",
		"sellerInitialPrice", "customerNegotiatedPrice",
		"sellerFinalPrice", "customerFinalResponse",
		"availability",
	}
	fields := getDefaultMonitoredFields()
	for _, f := range required {
		if !fields[f] {
			t.Errorf("expected monitored field %q to be present", f)
		}
	}
}

func TestDefaultMonitoredFields_NoPaymentFields(t *testing.T) {
	banned := []string{"paymentStatus", "paymentReceivedAt", "awaitingPaymentAt"}
	fields := getDefaultMonitoredFields()
	for _, f := range banned {
		if fields[f] {
			t.Errorf("payment field %q must NOT be in monitored fields", f)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// NewFieldComparator — environment override
// ─────────────────────────────────────────────────────────────────────────────

func TestNewFieldComparator_DefaultsWhenNoEnv(t *testing.T) {
	os.Unsetenv("MONITORED_FIELDS")
	fc := NewFieldComparator()
	if !fc.MonitoredFields["status"] {
		t.Error("default comparator must monitor 'status'")
	}
}

func TestNewFieldComparator_EnvOverride(t *testing.T) {
	t.Setenv("MONITORED_FIELDS", "myField,otherField")
	fc := NewFieldComparator()
	if !fc.MonitoredFields["myField"] || !fc.MonitoredFields["otherField"] {
		t.Error("env override should set exactly the specified fields")
	}
	// Default fields should NOT be present when overridden.
	if fc.MonitoredFields["status"] {
		t.Error("env override should replace (not merge) default fields")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// DetectChanges
// ─────────────────────────────────────────────────────────────────────────────

func TestDetectChanges_StatusChange(t *testing.T) {
	os.Unsetenv("MONITORED_FIELDS")
	fc := NewFieldComparator()
	changes := fc.DetectChanges(
		map[string]interface{}{"status": "new"},
		map[string]interface{}{"status": "pending_seller_price"},
	)
	if len(changes) != 1 {
		t.Fatalf("expected 1 change, got %d", len(changes))
	}
	if changes[0].FieldName != "status" {
		t.Errorf("expected FieldName=status; got %q", changes[0].FieldName)
	}
	if changes[0].OldValue != "new" {
		t.Errorf("OldValue: got %v", changes[0].OldValue)
	}
	if changes[0].NewValue != "pending_seller_price" {
		t.Errorf("NewValue: got %v", changes[0].NewValue)
	}
}

func TestDetectChanges_NoChangeForSameValue(t *testing.T) {
	os.Unsetenv("MONITORED_FIELDS")
	fc := NewFieldComparator()
	changes := fc.DetectChanges(
		map[string]interface{}{"status": "in_progress"},
		map[string]interface{}{"status": "in_progress"},
	)
	if len(changes) != 0 {
		t.Errorf("expected 0 changes for identical values; got %d", len(changes))
	}
}

func TestDetectChanges_IgnoresNonMonitoredFields(t *testing.T) {
	os.Unsetenv("MONITORED_FIELDS")
	fc := NewFieldComparator()
	changes := fc.DetectChanges(
		map[string]interface{}{"someUnmonitoredField": "old"},
		map[string]interface{}{"someUnmonitoredField": "new"},
	)
	if len(changes) != 0 {
		t.Errorf("non-monitored field changes should be ignored; got %d", len(changes))
	}
}

func TestDetectChanges_IgnoresIgnoredFields(t *testing.T) {
	os.Unsetenv("MONITORED_FIELDS")
	fc := NewFieldComparator()
	// updatedAt is in the ignored set; inject it into monitored too to test ignored wins.
	fc.MonitoredFields["updatedAt"] = true
	changes := fc.DetectChanges(
		map[string]interface{}{"updatedAt": "t1"},
		map[string]interface{}{"updatedAt": "t2"},
	)
	if len(changes) != 0 {
		t.Errorf("ignored field 'updatedAt' should produce 0 changes; got %d", len(changes))
	}
}

func TestDetectChanges_NewFieldAppearance(t *testing.T) {
	os.Unsetenv("MONITORED_FIELDS")
	fc := NewFieldComparator()
	changes := fc.DetectChanges(
		map[string]interface{}{},
		map[string]interface{}{"acceptedPrice": "200"},
	)
	found := false
	for _, c := range changes {
		if c.FieldName == "acceptedPrice" {
			found = true
			if c.OldValue != nil {
				t.Errorf("OldValue should be nil for new field")
			}
		}
	}
	if !found {
		t.Error("new field 'acceptedPrice' should be detected as a change")
	}
}

func TestDetectChanges_FieldRemovedFromDocument(t *testing.T) {
	os.Unsetenv("MONITORED_FIELDS")
	fc := NewFieldComparator()
	changes := fc.DetectChanges(
		map[string]interface{}{"availability": "available"},
		map[string]interface{}{},
	)
	found := false
	for _, c := range changes {
		if c.FieldName == "availability" {
			found = true
		}
	}
	if !found {
		t.Error("removed field 'availability' should be detected as a change")
	}
}

func TestDetectChanges_MultipleFieldsChanged(t *testing.T) {
	os.Unsetenv("MONITORED_FIELDS")
	fc := NewFieldComparator()
	changes := fc.DetectChanges(
		map[string]interface{}{"status": "new", "acceptedPrice": "", "sellerInitialPrice": "100"},
		map[string]interface{}{"status": "pending_seller_price", "acceptedPrice": "200", "sellerInitialPrice": "100"},
	)
	if len(changes) != 2 {
		t.Errorf("expected 2 changes (status + acceptedPrice), got %d", len(changes))
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// DetectChangesByUpdateMask
// ─────────────────────────────────────────────────────────────────────────────

func TestDetectChangesByUpdateMask_OnlyMaskedField(t *testing.T) {
	os.Unsetenv("MONITORED_FIELDS")
	fc := NewFieldComparator()
	changes := fc.DetectChangesByUpdateMask(
		map[string]interface{}{"status": "new", "availability": "yes"},
		map[string]interface{}{"status": "resolved", "availability": "yes"},
		[]string{"status"},
	)
	if len(changes) != 1 || changes[0].FieldName != "status" {
		t.Errorf("expected only 'status' change; got %+v", changes)
	}
}

func TestDetectChangesByUpdateMask_MaskedFieldUnmonitored(t *testing.T) {
	os.Unsetenv("MONITORED_FIELDS")
	fc := NewFieldComparator()
	// "viewCount" is not in monitored set → should be skipped even if in mask.
	changes := fc.DetectChangesByUpdateMask(
		map[string]interface{}{"viewCount": 1},
		map[string]interface{}{"viewCount": 2},
		[]string{"viewCount"},
	)
	if len(changes) != 0 {
		t.Errorf("unmonitored masked field should be ignored; got %d changes", len(changes))
	}
}

func TestDetectChangesByUpdateMask_EmptyMask_FallsBackToFullComparison(t *testing.T) {
	os.Unsetenv("MONITORED_FIELDS")
	fc := NewFieldComparator()
	changes := fc.DetectChangesByUpdateMask(
		map[string]interface{}{"status": "new"},
		map[string]interface{}{"status": "resolved"},
		[]string{},
	)
	if len(changes) == 0 {
		t.Error("empty mask should fall back to full comparison and detect the status change")
	}
}

func TestDetectChangesByUpdateMask_DottedPath_UsesRootField(t *testing.T) {
	// When the mask contains "metadata.status", only the root "metadata" part
	// should be looked up. If "metadata" is monitored and changed, it's a change.
	os.Unsetenv("MONITORED_FIELDS")
	fc := NewFieldComparator()
	fc.MonitoredFields["metadata"] = true
	changes := fc.DetectChangesByUpdateMask(
		map[string]interface{}{"metadata": map[string]interface{}{"step": 1}},
		map[string]interface{}{"metadata": map[string]interface{}{"step": 2}},
		[]string{"metadata.step"},
	)
	if len(changes) == 0 {
		t.Error("dotted-path mask should resolve to root field 'metadata' and detect change")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// IsSignificantChange
// ─────────────────────────────────────────────────────────────────────────────

func TestIsSignificantChange_MonitoredField(t *testing.T) {
	os.Unsetenv("MONITORED_FIELDS")
	fc := NewFieldComparator()
	change := FieldChange{FieldName: "status", OldValue: "new", NewValue: "resolved"}
	if !fc.IsSignificantChange(change) {
		t.Error("change to monitored field 'status' should be significant")
	}
}

func TestIsSignificantChange_IgnoredField(t *testing.T) {
	os.Unsetenv("MONITORED_FIELDS")
	fc := NewFieldComparator()
	change := FieldChange{FieldName: "updatedAt", OldValue: "t1", NewValue: "t2"}
	if fc.IsSignificantChange(change) {
		t.Error("change to ignored field should NOT be significant")
	}
}

func TestIsSignificantChange_UnmonitoredField(t *testing.T) {
	os.Unsetenv("MONITORED_FIELDS")
	fc := NewFieldComparator()
	change := FieldChange{FieldName: "unknownField", OldValue: "a", NewValue: "b"}
	if fc.IsSignificantChange(change) {
		t.Error("change to unmonitored field should NOT be significant")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// ValuesEqual
// ─────────────────────────────────────────────────────────────────────────────

func TestValuesEqual(t *testing.T) {
	cases := []struct {
		name string
		a, b interface{}
		want bool
	}{
		{"both nil", nil, nil, true},
		{"nil vs value", nil, "x", false},
		{"same string", "hello", "hello", true},
		{"different string", "hello", "world", false},
		{"same int", int64(42), int64(42), true},
		{"same bool", true, true, true},
		{"different bool", true, false, false},
		{"int64 vs float64 same val", int64(10), float64(10), false}, // JSON: "10" vs "10" → true actually
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := ValuesEqual(tc.a, tc.b)
			if got != tc.want {
				t.Errorf("ValuesEqual(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
			}
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// GetChangesSummary
// ─────────────────────────────────────────────────────────────────────────────

func TestGetChangesSummary_Empty(t *testing.T) {
	s := GetChangesSummary(nil)
	if s != "No significant changes" {
		t.Errorf("empty changes summary: got %q", s)
	}
}

func TestGetChangesSummary_SingleChange(t *testing.T) {
	changes := []FieldChange{{FieldName: "status", OldValue: "new", NewValue: "resolved"}}
	s := GetChangesSummary(changes)
	if !strings.Contains(s, "status") || !strings.Contains(s, "new") || !strings.Contains(s, "resolved") {
		t.Errorf("summary should mention field and values; got %q", s)
	}
}

func TestGetChangesSummary_NilOldValue(t *testing.T) {
	changes := []FieldChange{{FieldName: "acceptedPrice", OldValue: nil, NewValue: "200"}}
	s := GetChangesSummary(changes)
	if !strings.Contains(s, "(none)") {
		t.Errorf("nil oldValue should show '(none)'; got %q", s)
	}
}

func TestGetChangesSummary_NilNewValue(t *testing.T) {
	changes := []FieldChange{{FieldName: "availability", OldValue: "available", NewValue: nil}}
	s := GetChangesSummary(changes)
	if !strings.Contains(s, "(deleted)") {
		t.Errorf("nil newValue should show '(deleted)'; got %q", s)
	}
}
