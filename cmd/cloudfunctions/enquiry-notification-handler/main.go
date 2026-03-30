// Package main implements a Google Cloud Function (Gen 2) that listens to Firestore
// enquiry document update events via Eventarc.
//
// The function:
// 1. Receives Firestore document update events via Eventarc
// 2. Compares old and new field values
// 3. Identifies significant field changes (e.g., status, assignedTo)
// 4. Sends Firebase Cloud Messaging (FCM) push notifications to relevant users
// 5. Implements idempotent and production-grade processing
//
// Environment Variables:
// - MONITORED_FIELDS: Comma-separated list of fields to monitor (overrides defaults)
// - ENABLE_IDEMPOTENCY_CHECK: Set to "true" to track notification history
// - LOG_LEVEL: DEBUG, INFO, WARN, ERROR (default: INFO)
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/funcframework"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	firestoredata "github.com/googleapis/google-cloudevents-go/cloud/firestoredata"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/service/notification"
	firestoreutil "github.com/rohit221990/mandi-backend/pkg/utils/firestore"
	"google.golang.org/protobuf/proto"
)

// Logger provides structured logging with levels
type Logger struct {
	level string
}

var logger = &Logger{level: getLogLevel()}

// init registers the Cloud Function entry point
func init() {
	functions.CloudEvent("ProcessEnquiryUpdate", ProcessEnquiryUpdate)
}

// ProcessEnquiryUpdate is the main Cloud Function that processes Firestore enquiry updates
// It's called via Eventarc when a Firestore document in the enquiries collection is updated
// (google.cloud.firestore.document.v1.updated)
func ProcessEnquiryUpdate(ctx context.Context, ce cloudevents.Event) error {
	logger.Info("Starting ProcessEnquiryUpdate")

	// Extract Firestore event data from the CloudEvent payload.
	// Firestore Eventarc events arrive as Protocol Buffer binary, NOT JSON.
	rawData := ce.Data()
	if len(rawData) == 0 {
		logger.Error("CloudEvent has no data")
		return fmt.Errorf("CloudEvent missing data")
	}

	// Decode the protobuf-encoded Firestore document event.
	var docEvent firestoredata.DocumentEventData
	if err := proto.Unmarshal(rawData, &docEvent); err != nil {
		logger.Error(fmt.Sprintf("Failed to unmarshal protobuf event data: %v", err))
		return fmt.Errorf("failed to unmarshal protobuf event data: %w", err)
	}

	// Convert proto types to the existing domain model.
	event := convertProtoToFirestoreEvent(&docEvent, ce.ID())

	logger.Debug(fmt.Sprintf("Received event: %s", event.ID))

	// Process the event
	return handleEnquiryUpdate(ctx, event)
}

// convertProtoToFirestoreEvent converts a firestoredata.DocumentEventData (protobuf) to
// the domain.FirestoreEvent used by the rest of the notification pipeline.
func convertProtoToFirestoreEvent(docEvent *firestoredata.DocumentEventData, eventID string) *domain.FirestoreEvent {
	eventData := domain.FirestoreEventData{}

	if docEvent.Value != nil {
		eventData.Value = convertProtoDocument(docEvent.Value)
	}
	if docEvent.OldValue != nil {
		eventData.OldValue = convertProtoDocument(docEvent.OldValue)
	}
	if docEvent.UpdateMask != nil {
		eventData.UpdateMask = &domain.UpdateMask{
			FieldPaths: docEvent.UpdateMask.FieldPaths,
		}
	}

	return &domain.FirestoreEvent{
		Data: eventData,
		ID:   eventID,
	}
}

// convertProtoDocument converts a Firestore proto Document to the domain model.
func convertProtoDocument(doc *firestoredata.Document) *domain.FirestoreDocument {
	if doc == nil {
		return nil
	}
	fields := make(map[string]interface{}, len(doc.Fields))
	for k, v := range doc.Fields {
		fields[k] = convertProtoValue(v)
	}
	result := &domain.FirestoreDocument{
		Name:   doc.Name,
		Fields: fields,
	}
	if doc.CreateTime != nil {
		result.CreateTime = doc.CreateTime.AsTime().Format(time.RFC3339)
	}
	if doc.UpdateTime != nil {
		result.UpdateTime = doc.UpdateTime.AsTime().Format(time.RFC3339)
	}
	return result
}

// convertProtoValue converts a Firestore proto Value to a native Go value so that
// the existing ParseFields / ExtractFirestoreValue helpers can pass it through.
func convertProtoValue(v *firestoredata.Value) interface{} {
	if v == nil {
		return nil
	}
	switch val := v.ValueType.(type) {
	case *firestoredata.Value_NullValue:
		_ = val
		return nil
	case *firestoredata.Value_BooleanValue:
		return val.BooleanValue
	case *firestoredata.Value_IntegerValue:
		return fmt.Sprintf("%d", val.IntegerValue)
	case *firestoredata.Value_DoubleValue:
		return val.DoubleValue
	case *firestoredata.Value_TimestampValue:
		if val.TimestampValue != nil {
			return val.TimestampValue.AsTime().Format(time.RFC3339)
		}
		return ""
	case *firestoredata.Value_StringValue:
		return val.StringValue
	case *firestoredata.Value_BytesValue:
		return string(val.BytesValue)
	case *firestoredata.Value_ReferenceValue:
		return val.ReferenceValue
	case *firestoredata.Value_GeoPointValue:
		if val.GeoPointValue != nil {
			return map[string]interface{}{
				"latitude":  val.GeoPointValue.Latitude,
				"longitude": val.GeoPointValue.Longitude,
			}
		}
		return nil
	case *firestoredata.Value_ArrayValue:
		if val.ArrayValue == nil {
			return []interface{}{}
		}
		arr := make([]interface{}, 0, len(val.ArrayValue.Values))
		for _, item := range val.ArrayValue.Values {
			arr = append(arr, convertProtoValue(item))
		}
		return arr
	case *firestoredata.Value_MapValue:
		if val.MapValue == nil {
			return map[string]interface{}{}
		}
		m := make(map[string]interface{}, len(val.MapValue.Fields))
		for k, mv := range val.MapValue.Fields {
			m[k] = convertProtoValue(mv)
		}
		return m
	}
	return nil
}

// handleEnquiryUpdate processes a single enquiry update event
func handleEnquiryUpdate(ctx context.Context, event *domain.FirestoreEvent) error {
	defer func() {
		if r := recover(); r != nil {
			logger.Error(fmt.Sprintf("PANIC in handleEnquiryUpdate: %v", r))
		}
	}()

	// Parse and validate event
	eventHandler := firestoreutil.NewEventHandler()
	parsedEvent, err := eventHandler.ParseEvent(event)
	if err != nil {
		logger.Warn(fmt.Sprintf("Failed to parse event: %v", err))
		return nil // Not an error we should fail on - may be metadata change
	}

	logger.Info(fmt.Sprintf("Parsed enquiry update: %s", parsedEvent.DocumentID))

	// Detect field changes
	changes := eventHandler.FindChanges(parsedEvent)
	if !eventHandler.HasSignificantChanges(changes) {
		logger.Info(fmt.Sprintf("No significant changes detected for %s - skipping notification",
			parsedEvent.DocumentID))
		return nil // No changes to monitored fields
	}

	logger.Info(fmt.Sprintf("Detected %d significant changes: %v",
		len(changes), firestoreutil.GetChangedFieldNames(changes)))

	// Build notification payload
	payloadBuilder := notification.NewPayloadBuilder()
	payload := payloadBuilder.BuildPayload(parsedEvent, changes)

	// Validate payload
	if err := notification.ValidatePayload(payload); err != nil {
		logger.Error(fmt.Sprintf("Invalid notification payload: %v", err))
		return err
	}

	logger.Debug(fmt.Sprintf("Payload: title=%s, body=%s, enquiryID=%s",
		payload.Title, payload.Body, payload.EnquiryID))

	// Initialize notification service
	notifConfig := notification.Config{
		ProjectID:                     os.Getenv("GCP_PROJECT"),
		EnableIdempotencyCheck:        strings.ToLower(os.Getenv("ENABLE_IDEMPOTENCY_CHECK")) == "true",
		FCMTokenCollection:            "fcmTokens",
		NotificationHistoryCollection: "notificationHistory",
	}

	// If ProjectID not set, try to get from environment or use default
	if notifConfig.ProjectID == "" {
		notifConfig.ProjectID = os.Getenv("GOOGLE_CLOUD_PROJECT")
	}

	svc, err := notification.NewService(ctx, notifConfig)
	if err != nil {
		logger.Error(fmt.Sprintf("Failed to initialize notification service: %v", err))
		return fmt.Errorf("failed to initialize notification service: %w", err)
	}
	defer svc.Close()

	// Send notification
	if err := svc.SendNotification(ctx, parsedEvent, changes, payload); err != nil {
		logger.Error(fmt.Sprintf("Failed to send notification: %v", err))
		// Log error but don't fail - notification is secondary to main operation
		return nil
	}

	logger.Info(fmt.Sprintf("Successfully processed enquiry update: %s", parsedEvent.DocumentID))
	return nil
}

// Logger methods
func (l *Logger) Debug(msg string) {
	if l.level == "DEBUG" {
		log.Printf("[DEBUG] %s", msg)
	}
}

func (l *Logger) Info(msg string) {
	log.Printf("[INFO] %s", msg)
}

func (l *Logger) Warn(msg string) {
	log.Printf("[WARN] %s", msg)
}

func (l *Logger) Error(msg string) {
	log.Printf("[ERROR] %s", msg)
}

// getLogLevel retrieves log level from environment
func getLogLevel() string {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "INFO"
	} else {
		level = strings.ToUpper(level)
		if level != "DEBUG" && level != "INFO" && level != "WARN" && level != "ERROR" {
			level = "INFO"
		}
	}
	return level
}

// main is required for local testing
// For Cloud Functions deployment, only the ProcessEnquiryUpdate function is used
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting enquiry notification handler on port %s", port)
	if err := funcframework.Start(port); err != nil {
		log.Fatalf("funcframework.Start: %v", err)
	}
}
