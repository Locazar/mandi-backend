package firestore

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// FieldComparator handles comparison of Firestore document fields
type FieldComparator struct {
	// Monitored fields: only notify if these fields change
	MonitoredFields map[string]bool
	// Ignored fields: never notify if only these change
	IgnoredFields map[string]bool
}

// NewFieldComparator creates a new field comparator with default monitored fields
func NewFieldComparator() *FieldComparator {
	// Load monitored fields from environment or use defaults
	monitoredFieldsEnv := os.Getenv("MONITORED_FIELDS")
	monitoredFields := getDefaultMonitoredFields()

	if monitoredFieldsEnv != "" {
		// Override with environment variable (comma-separated)
		customFields := strings.Split(monitoredFieldsEnv, ",")
		monitoredFields = make(map[string]bool)
		for _, field := range customFields {
			field = strings.TrimSpace(field)
			if field != "" {
				monitoredFields[field] = true
			}
		}
	}

	return &FieldComparator{
		MonitoredFields: monitoredFields,
		IgnoredFields: map[string]bool{
			"updatedAt":    true,
			"lastModified": true,
			"viewCount":    true,
		},
	}
}

// DetectChanges compares old and new values for monitored fields
// Returns slice of FieldChange for fields that changed
// Returns empty slice if no significant changes detected (idempotent)
func (fc *FieldComparator) DetectChanges(oldFields, newFields map[string]interface{}) []domain.FieldChange {
	changes := []domain.FieldChange{}

	// Check all monitored fields for changes
	for fieldName := range fc.MonitoredFields {
		oldVal, oldExists := oldFields[fieldName]
		newVal, newExists := newFields[fieldName]

		// Skip if field doesn't exist in either version
		if !oldExists && !newExists {
			continue
		}

		// Skip if field exists in both and values are identical
		if oldExists && newExists && ValuesEqual(oldVal, newVal) {
			continue
		}

		// Skip if in ignored fields (e.g., system timestamps)
		if fc.IgnoredFields[fieldName] {
			log.Printf("DEBUG: Ignoring change to ignored field: %s", fieldName)
			continue
		}

		// Change detected
		changes = append(changes, domain.FieldChange{
			FieldName: fieldName,
			OldValue:  oldVal,
			NewValue:  newVal,
		})

		log.Printf("DEBUG: Field %s changed from %v to %v", fieldName, oldVal, newVal)
	}

	return changes
}

// DetectChangesByUpdateMask compares old and new values only for fields in updateMask
// More efficient when updateMask is provided
func (fc *FieldComparator) DetectChangesByUpdateMask(
	oldFields, newFields map[string]interface{},
	updateMaskPaths []string,
) []domain.FieldChange {
	changes := []domain.FieldChange{}

	// If updateMask is provided, only check those fields
	if len(updateMaskPaths) > 0 {
		for _, fieldPath := range updateMaskPaths {
			// Handle nested paths like "metadata.status" -> only check "metadata"
			fieldName := strings.Split(fieldPath, ".")[0]

			// Skip if not in monitored fields
			if !fc.MonitoredFields[fieldName] {
				continue
			}

			oldVal, oldExists := oldFields[fieldName]
			newVal, newExists := newFields[fieldName]

			if oldExists || newExists {
				if !ValuesEqual(oldVal, newVal) {
					changes = append(changes, domain.FieldChange{
						FieldName: fieldName,
						OldValue:  oldVal,
						NewValue:  newVal,
					})
					log.Printf("DEBUG: Field %s changed (via updateMask) from %v to %v", fieldName, oldVal, newVal)
				}
			}
		}
	} else {
		// Fall back to full comparison
		changes = fc.DetectChanges(oldFields, newFields)
	}

	return changes
}

// IsSignificantChange determines if a change warrants a notification
func (fc *FieldComparator) IsSignificantChange(change domain.FieldChange) bool {
	// Don't notify on changes to ignored fields
	if fc.IgnoredFields[change.FieldName] {
		return false
	}

	// Don't notify if field is not monitored
	if !fc.MonitoredFields[change.FieldName] {
		return false
	}

	return true
}

// GetChangesSummary returns a human-readable summary of changes
func GetChangesSummary(changes []domain.FieldChange) string {
	if len(changes) == 0 {
		return "No significant changes"
	}

	summaryParts := []string{}
	for _, change := range changes {
		oldStr := fmt.Sprintf("%v", change.OldValue)
		newStr := fmt.Sprintf("%v", change.NewValue)

		if change.OldValue == nil {
			oldStr = "(none)"
		}
		if change.NewValue == nil {
			newStr = "(deleted)"
		}

		summaryParts = append(summaryParts,
			fmt.Sprintf("%s: %s → %s", change.FieldName, oldStr, newStr))
	}

	return strings.Join(summaryParts, "; ")
}

// getDefaultMonitoredFields returns the default set of fields to monitor
func getDefaultMonitoredFields() map[string]bool {
	return map[string]bool{
		// Enquiry status and assignment
		"status":         true,
		"assignedTo":     true,
		"assignedToName": true,
		"assignedToRole": true,

		// Enquiry metadata
		"priority":       true,
		"resolutionDate": true,
		"closedAt":       true,
		"tags":           true,

		// Enquiry content
		"subject":       false, // usually not needed
		"description":   false,
		"category":      true,
		"type":          true,

		// Communication
		"notes":         false, // might be too verbose
		"message":       false,
		"response":      true,
		"respondedAt":   true,

		// Custom fields (add more based on your schema)
		"customStatus":  true,
		"department":    true,
	}
}
