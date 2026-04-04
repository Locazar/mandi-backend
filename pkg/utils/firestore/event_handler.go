package firestore

import (
	"fmt"
	"log"

	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// EventHandler parses and processes Firestore events
type EventHandler struct {
	comparator *FieldComparator
}

// NewEventHandler creates a new event handler
func NewEventHandler() *EventHandler {
	return &EventHandler{
		comparator: NewFieldComparator(),
	}
}

// ParseEvent parses a raw Firestore Eventarc event into a structured format
func (eh *EventHandler) ParseEvent(event *domain.FirestoreEvent) (*domain.ParsedFirestoreEvent, error) {
	if event == nil {
		return nil, fmt.Errorf("event is nil")
	}

	// Extract document path and ID from the value's name
	documentPath := ""
	documentID := ""

	if event.Data.Value != nil && event.Data.Value.Name != "" {
		documentPath, documentID = ExtractDocumentPath(event.Data.Value.Name)
	}

	if documentID == "" && event.Data.OldValue != nil && event.Data.OldValue.Name != "" {
		documentPath, documentID = ExtractDocumentPath(event.Data.OldValue.Name)
	}

	if documentID == "" {
		return nil, fmt.Errorf("could not extract document ID from event")
	}

	// Guard against nil UpdateMask — events without a mask do a full field comparison.
	var updatedPaths []string
	if event.Data.UpdateMask != nil {
		updatedPaths = event.Data.UpdateMask.FieldPaths
	}

	parsed := &domain.ParsedFirestoreEvent{
		DocumentPath: documentPath,
		DocumentID:   documentID,
		UpdateTime:   "",
		OldFields:    make(map[string]interface{}),
		NewFields:    make(map[string]interface{}),
		UpdatedPaths: updatedPaths,
	}

	// Parse old value (before update)
	if event.Data.OldValue != nil {
		if len(event.Data.OldValue.Fields) > 0 {
			parsed.OldFields = ParseFields(event.Data.OldValue.Fields)
		}
		if event.Data.OldValue.UpdateTime != "" {
			parsed.UpdateTime = event.Data.OldValue.UpdateTime
		}
	}

	// Parse new value (after update)
	if event.Data.Value != nil {
		if len(event.Data.Value.Fields) > 0 {
			parsed.NewFields = ParseFields(event.Data.Value.Fields)
		}
		if event.Data.Value.UpdateTime != "" {
			parsed.UpdateTime = event.Data.Value.UpdateTime
		}
	}

	log.Printf("INFO: Parsed event - DocumentID: %s, UpdateTime: %s, Fields: old=%d, new=%d",
		parsed.DocumentID, parsed.UpdateTime, len(parsed.OldFields), len(parsed.NewFields))

	return parsed, nil
}

// FindChanges detects which fields have changed and should trigger notifications
func (eh *EventHandler) FindChanges(event *domain.ParsedFirestoreEvent) []domain.FieldChange {
	if event == nil {
		log.Printf("WARN: Event is nil")
		return []domain.FieldChange{}
	}

	// Use update mask if available for efficiency
	if len(event.UpdatedPaths) > 0 {
		log.Printf("DEBUG: Using updateMask with %d paths", len(event.UpdatedPaths))
		return eh.comparator.DetectChangesByUpdateMask(
			event.OldFields,
			event.NewFields,
			event.UpdatedPaths,
		)
	}

	// Fall back to full field comparison
	return eh.comparator.DetectChanges(event.OldFields, event.NewFields)
}

// HasSignificantChanges checks if there are any changes that warrant a notification
func (eh *EventHandler) HasSignificantChanges(changes []domain.FieldChange) bool {
	log.Printf("DEBUG: Checking %d changes for significance", len(changes))
	for _, change := range changes {
		if eh.comparator.IsSignificantChange(change) {
			return true
		}
	}
	return false
}

// GetChangedFieldNames returns list of field names that changed
func GetChangedFieldNames(changes []domain.FieldChange) []string {
	names := make([]string, len(changes))
	for i, change := range changes {
		names[i] = change.FieldName
	}
	return names
}
