package notification

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/utils/firestore"
)

// PayloadBuilder constructs notification payloads from Firestore events
type PayloadBuilder struct {
}

// NewPayloadBuilder creates a new payload builder
func NewPayloadBuilder() *PayloadBuilder {
	return &PayloadBuilder{}
}

// BuildPayload creates a notification payload from event and changes
func (pb *PayloadBuilder) BuildPayload(
	event *domain.ParsedFirestoreEvent,
	changes []domain.FieldChange,
) *domain.NotificationPayload {
	payload := &domain.NotificationPayload{
		DocumentID:    event.DocumentID,
		DocumentPath:  event.DocumentPath,
		Timestamp:     event.UpdateTime,
		ChangeCount:   len(changes),
		ChangedFields: firestore.GetChangedFieldNames(changes),
	}

	// Extract key fields from new values
	payload.EnquiryID = firstNonEmptyField(event.NewFields, "queryId", "enquiryId", "id")
	payload.UserID = firstNonEmptyField(event.NewFields, "userId", "customerId", "createdBy")
	payload.AssignedTo = firstNonEmptyField(event.NewFields, "assignedTo", "assignedToId")

	// Set action URL based on enquiry ID
	if payload.EnquiryID != "" {
		payload.ActionURL = fmt.Sprintf("/enquiries/%s", payload.EnquiryID)
	} else {
		payload.ActionURL = fmt.Sprintf("/documents/%s", event.DocumentID)
	}

	// Generate notification title and body based on changes
	payload.Title, payload.Body = pb.generateNotificationContent(event, changes)

	log.Printf("DEBUG: Payload generated - EnquiryID: %s, User: %s, Assigned: %s",
		payload.EnquiryID, payload.UserID, payload.AssignedTo)

	return payload
}

// generateNotificationContent creates user-friendly notification title and body
func (pb *PayloadBuilder) generateNotificationContent(
	event *domain.ParsedFirestoreEvent,
	changes []domain.FieldChange,
) (title, body string) {
	if len(changes) == 0 {
		return "Enquiry Updated", "Your enquiry has been updated"
	}

	// Build notification content based on field changes
	changeMap := make(map[string]domain.FieldChange)
	for _, change := range changes {
		changeMap[change.FieldName] = change
	}

	// Check for specific field changes and prioritize
	if statusChange, hasStatus := changeMap["status"]; hasStatus {
		return pb.buildStatusNotification(statusChange, event.NewFields)
	}

	if assignedChange, hasAssigned := changeMap["assignedTo"]; hasAssigned {
		return pb.buildAssignmentNotification(assignedChange, event)
	}

	if responseChange, hasResponse := changeMap["response"]; hasResponse {
		return pb.buildResponseNotification(responseChange)
	}

	if closedChange, hasClosed := changeMap["closedAt"]; hasClosed {
		if closedChange.NewValue != nil && closedChange.NewValue != "" {
			return "Enquiry Closed", "Your enquiry has been resolved and closed"
		}
	}

	// Generic message for multiple changes
	changeList := []string{}
	for _, fieldName := range firestore.GetChangedFieldNames(changes) {
		changeList = append(changeList, formatFieldName(fieldName))
	}

	title = "Enquiry Updated"
	body = fmt.Sprintf("%d field(s) updated: %s",
		len(changes),
		strings.Join(changeList, ", "))

	if len(body) > 240 {
		body = fmt.Sprintf("%d field(s) have been updated", len(changes))
	}

	return title, body
}

// buildStatusNotification creates production-ready notification copy for status changes.
func (pb *PayloadBuilder) buildStatusNotification(change domain.FieldChange, fields map[string]interface{}) (title, body string) {
	oldStatus := fmt.Sprintf("%v", change.OldValue)
	newStatus := strings.ToLower(strings.TrimSpace(fmt.Sprintf("%v", change.NewValue)))
	enquiryRef := "this enquiry"
	if enquiryID := firstNonEmptyField(fields, "queryId", "enquiryId", "id"); enquiryID != "" {
		enquiryRef = fmt.Sprintf("enquiry #%s", enquiryID)
	}

	formatPrice := func(v interface{}) string {
		s := strings.TrimSpace(fmt.Sprintf("%v", v))
		if s == "" || s == "<nil>" {
			return ""
		}
		return "Rs. " + s
	}

	availabilityText := strings.TrimSpace(firstNonEmptyField(fields, "availability"))
	askQuantity := strings.TrimSpace(firstNonEmptyField(fields, "askQuantity", "ask_quantity"))
	sellerInitialPrice := formatPrice(fields["sellerInitialPrice"])
	customerNegotiatedPrice := formatPrice(fields["customerNegotiatedPrice"])
	sellerFinalPrice := formatPrice(fields["sellerFinalPrice"])
	customerFinalResponse := formatPrice(fields["customerFinalResponse"])
	acceptedPrice := formatPrice(fields["acceptedPrice"])
	acceptedBy := firstNonEmptyField(fields, "acceptedBy", "accepted_by")
	rejectedBy := firstNonEmptyField(fields, "rejectedBy", "rejected_by")

	title = "Enquiry Updated"

	switch newStatus {
	case "pending_seller_price":
		title = "Price Request Pending"
		segments := []string{fmt.Sprintf("A buyer update requires your price response for %s.", enquiryRef)}
		if askQuantity != "" {
			segments = append(segments, fmt.Sprintf("Requested quantity: %s.", askQuantity))
		}
		if availabilityText != "" {
			segments = append(segments, fmt.Sprintf("Availability: %s.", availabilityText))
		}
		return title, strings.Join(segments, " ")
	case "pending_customer_price":
		title = "Seller Price Shared"
		if sellerInitialPrice != "" {
			return title, fmt.Sprintf("The seller has shared an initial price of %s for %s.", sellerInitialPrice, enquiryRef)
		}
		return title, fmt.Sprintf("The seller has shared an initial price update for %s.", enquiryRef)
	case "pending_seller_final":
		title = "Customer Counter Offer Received"
		if customerNegotiatedPrice != "" {
			return title, fmt.Sprintf("The customer proposed %s for %s. Review and send your final response.", customerNegotiatedPrice, enquiryRef)
		}
		return title, fmt.Sprintf("The customer updated their negotiated price for %s.", enquiryRef)
	case "pending_customer_final":
		title = "Seller Final Price Shared"
		if sellerFinalPrice != "" {
			return title, fmt.Sprintf("The seller has shared a final price of %s for %s.", sellerFinalPrice, enquiryRef)
		}
		return title, fmt.Sprintf("The seller has shared the final price for %s.", enquiryRef)
	case "seller_final_update":
		title = "Customer Final Response Received"
		if customerFinalResponse != "" {
			return title, fmt.Sprintf("The customer submitted a final response of %s for %s.", customerFinalResponse, enquiryRef)
		}
		return title, fmt.Sprintf("The customer submitted a final response for %s.", enquiryRef)
	case "completed_accepted":
		title = "Deal Accepted"
		body = fmt.Sprintf("%s has been accepted", enquiryRef)
		if acceptedPrice != "" {
			body += fmt.Sprintf(" at %s", acceptedPrice)
		}
		if acceptedBy != "" {
			body += fmt.Sprintf(" by %s", acceptedBy)
		}
		return title, body + "."
	case "completed_rejected":
		title = "Deal Rejected"
		body = fmt.Sprintf("%s has been rejected", enquiryRef)
		if acceptedPrice != "" {
			body += fmt.Sprintf(" at %s", acceptedPrice)
		}
		if rejectedBy != "" {
			body += fmt.Sprintf(" by %s", rejectedBy)
		} else if acceptedBy != "" {
			body += fmt.Sprintf(" by %s", acceptedBy)
		}
		return title, body + "."
	case "new":
		return "Enquiry Created", "A new enquiry has been created."
	case "in_progress", "inprogress":
		return "Enquiry In Progress", fmt.Sprintf("%s is now being handled.", enquiryRef)
	case "resolved":
		return "Enquiry Resolved", fmt.Sprintf("%s has been resolved.", enquiryRef)
	case "closed":
		return "Enquiry Closed", fmt.Sprintf("%s has been closed.", enquiryRef)
	case "rejected":
		return "Enquiry Rejected", fmt.Sprintf("%s status changed to rejected.", enquiryRef)
	default:
		return "Enquiry Status Updated", fmt.Sprintf("Status changed from %s to %s.", oldStatus, change.NewValue)
	}
}

// buildAssignmentNotification creates notification for assignment changes
func (pb *PayloadBuilder) buildAssignmentNotification(
	change domain.FieldChange,
	event *domain.ParsedFirestoreEvent,
) (title, body string) {
	assignedToName := firestore.GetFieldAsString(event.NewFields, "assignedToName")
	assignedTo := fmt.Sprintf("%v", change.NewValue)

	title = "Enquiry Assigned"

	if assignedTo == "" || assignedTo == "<nil>" {
		body = "Your enquiry is no longer assigned"
	} else if assignedToName != "" {
		body = fmt.Sprintf("Your enquiry has been assigned to %s", assignedToName)
	} else {
		body = fmt.Sprintf("Your enquiry has been assigned to an agent")
	}

	return title, body
}

// buildResponseNotification creates notification for response changes
func (pb *PayloadBuilder) buildResponseNotification(change domain.FieldChange) (title, body string) {
	title = "New Response"
	body = "There's a new response to your enquiry"
	return title, body
}

// formatFieldName converts camelCase field names to readable format
func formatFieldName(fieldName string) string {
	// Convert camelCase to readable format
	result := ""
	for i, char := range fieldName {
		if i > 0 && char >= 'A' && char <= 'Z' {
			result += " " + string(char)
		} else {
			result += string(char)
		}
	}
	// Capitalize first letter
	if len(result) > 0 {
		result = strings.ToUpper(string(result[0])) + result[1:]
	}
	return result
}

// ValidatePayload ensures payload has all required fields
func ValidatePayload(payload *domain.NotificationPayload) error {
	if payload == nil {
		return fmt.Errorf("payload is nil")
	}

	if payload.DocumentID == "" {
		return fmt.Errorf("payload missing documentID")
	}

	if payload.Title == "" {
		return fmt.Errorf("payload missing title")
	}

	if payload.Body == "" {
		return fmt.Errorf("payload missing body")
	}

	if payload.Timestamp == "" {
		// Use current time if not set
		payload.Timestamp = time.Now().Format(time.RFC3339)
	}

	return nil
}

// firstNonEmptyField returns the first non-empty string value found under any
// of the provided field names.  It replaces the incorrect multi-arg calls to
// firestore.GetFieldAsString that only accepts a single key.
func firstNonEmptyField(fields map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := fields[k]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}
