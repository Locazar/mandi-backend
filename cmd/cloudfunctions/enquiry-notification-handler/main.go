// Package enquirynotification implements a Google Cloud Function (Gen 2) that listens to Firestore
// enquiry document creation and update events via Eventarc.
//
// Registered Cloud Function entry-points:
//   - ProcessEnquiryCreate  — google.cloud.firestore.document.v1.created
//   - ProcessEnquiryUpdate  — google.cloud.firestore.document.v1.updated
//
// Business pipeline (both functions):
//  1. Hard-deadline context enforced (FUNCTION_TIMEOUT_SECONDS, default 540 s).
//  2. CloudEvent payload deserialised from Firestore protobuf binary.
//  3. Old/new document fields compared; only monitored fields trigger notifications.
//  4. Notification payloads built with context-accurate copy per status transition.
//  5. Active FCM tokens fetched for buyer and/or seller from Firestore sub-collections.
//  6. Multi-platform push sent (Android high-priority / APNs / Web).
//  7. Invalid tokens automatically deactivated in Firestore after each send.
//  8. Delivery recorded in Firestore for optional idempotency (ENABLE_IDEMPOTENCY_CHECK).
//
// Error-handling contract:
//   - Non-retriable errors (malformed payload, missing IDs) return nil so Eventarc
//     does not generate infinite retries.
//   - Retriable errors (FCM/Firestore service unavailable) are returned so that
//     Eventarc's exponential back-off can reprocess the event.
//   - Panics are recovered, logged with full stack trace, and swallowed.
//
// Environment Variables:
//   - GCP_PROJECT / GOOGLE_CLOUD_PROJECT  : GCP project ID (auto-detected on Cloud Run)
//   - MONITORED_FIELDS                    : comma-separated override for watched fields
//   - ENABLE_IDEMPOTENCY_CHECK            : "true" enables Firestore deduplication
//   - LOG_LEVEL                           : DEBUG | INFO | WARN | ERROR  (default INFO)
//   - FUNCTION_TIMEOUT_SECONDS            : per-invocation deadline in seconds (default 540)
package enquirynotification

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"
	"time"

	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	firestoredata "github.com/googleapis/google-cloudevents-go/cloud/firestoredata"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/service/notification"
	firestoreutil "github.com/rohit221990/mandi-backend/pkg/utils/firestore"
	"google.golang.org/protobuf/proto"
)

// ─── Structured GCP logger ────────────────────────────────────────────────────

// invocationLogger writes JSON log lines to stdout.  Cloud Logging ingests these
// lines and attaches them to the correct Cloud Function invocation via the trace ID.
type invocationLogger struct {
	level   string
	traceID string // set to CloudEvent ID so all lines for one invocation share a trace
	project string
}

func newInvocationLogger(eventID string) *invocationLogger {
	lvl := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	switch lvl {
	case "DEBUG", "INFO", "WARN", "ERROR":
	default:
		lvl = "INFO"
	}
	return &invocationLogger{
		level:   lvl,
		traceID: eventID,
		project: resolveProjectID(),
	}
}

func (l *invocationLogger) emit(severity, msg string) {
	entry := map[string]interface{}{
		"severity": severity,
		"message":  msg,
		"time":     time.Now().UTC().Format(time.RFC3339Nano),
		"eventId":  l.traceID,
	}
	if l.project != "" && l.traceID != "" {
		entry["logging.googleapis.com/trace"] = fmt.Sprintf("projects/%s/traces/%s", l.project, l.traceID)
	}
	b, _ := json.Marshal(entry)
	fmt.Fprintln(os.Stdout, string(b))
}

func (l *invocationLogger) Debug(msg string) {
	if l.level == "DEBUG" {
		l.emit("DEBUG", msg)
	}
}
func (l *invocationLogger) Info(msg string)     { l.emit("INFO", msg) }
func (l *invocationLogger) Warn(msg string)     { l.emit("WARNING", msg) }
func (l *invocationLogger) Error(msg string)    { l.emit("ERROR", msg) }
func (l *invocationLogger) Critical(msg string) { l.emit("CRITICAL", msg) }

// ─── Configuration helpers ────────────────────────────────────────────────────

// resolveProjectID returns the GCP project from well-known environment variables.
func resolveProjectID() string {
	if v := os.Getenv("GCP_PROJECT"); v != "" {
		return v
	}
	return os.Getenv("GOOGLE_CLOUD_PROJECT")
}

// invocationTimeout returns the configured hard deadline per invocation.
func invocationTimeout() time.Duration {
	if s := os.Getenv("FUNCTION_TIMEOUT_SECONDS"); s != "" {
		if n, err := strconv.Atoi(s); err == nil && n > 0 {
			return time.Duration(n) * time.Second
		}
	}
	return 540 * time.Second // Cloud Functions Gen2 maximum
}

// buildNotifConfig constructs a notification.Config from the runtime environment.
func buildNotifConfig() notification.Config {
	return notification.Config{
		ProjectID:                     resolveProjectID(),
		EnableIdempotencyCheck:        strings.ToLower(os.Getenv("ENABLE_IDEMPOTENCY_CHECK")) == "true",
		FCMTokenCollection:            "fcmTokens",
		NotificationHistoryCollection: "notificationHistory",
	}
}

// firstFieldValue returns the first non-empty string found under any of the given keys
// in fields.  Safe against nil values and the "<nil>" stringer artefact.
func firstFieldValue(fields map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if v, ok := fields[key]; ok {
			if s := strings.TrimSpace(fmt.Sprintf("%v", v)); s != "" && s != "<nil>" {
				return s
			}
		}
	}
	return ""
}

// ─── Cloud Function registration ─────────────────────────────────────────────

// init registers the Cloud Function entry points.
func init() {
	functions.CloudEvent("ProcessEnquiryUpdate", ProcessEnquiryUpdate)
	functions.CloudEvent("ProcessEnquiryCreate", ProcessEnquiryCreate)
}

// ProcessEnquiryCreate handles google.cloud.firestore.document.v1.created events.
// A new enquiry always notifies the seller regardless of initial status.
func ProcessEnquiryCreate(ctx context.Context, ce cloudevents.Event) error {
	log := newInvocationLogger(ce.ID())
	log.Info(fmt.Sprintf("ProcessEnquiryCreate start eventId=%s source=%s", ce.ID(), ce.Source()))

	// Enforce hard per-invocation deadline.
	ctx, cancel := context.WithTimeout(ctx, invocationTimeout())
	defer cancel()

	// ── Deserialise protobuf payload ──────────────────────────────────────────
	rawData := ce.Data()
	if len(rawData) == 0 {
		log.Error("CloudEvent data is empty — non-retriable, skipping")
		return nil // malformed event; retrying will not help
	}

	var docEvent firestoredata.DocumentEventData
	if err := proto.Unmarshal(rawData, &docEvent); err != nil {
		log.Error(fmt.Sprintf("proto.Unmarshal failed: %v — non-retriable, skipping", err))
		return nil // corrupted protobuf will never succeed on retry
	}

	if docEvent.Value == nil {
		log.Warn("document.v1.created event carries no Value field — skipping")
		return nil
	}

	newDoc := convertProtoDocument(docEvent.Value)
	if newDoc == nil {
		log.Warn("convertProtoDocument returned nil — skipping")
		return nil
	}

	fields := newDoc.Fields

	// ── Resolve IDs from document fields ─────────────────────────────────────
	sellerID := firstFieldValue(fields, "sellerId", "seller_id", "shopId", "shop_id", "assignedTo")
	if sellerID == "" {
		// A document without a sellerId is legitimate (e.g., admin enquiries).
		log.Warn(fmt.Sprintf("No sellerId in document %s — skipping seller notification", newDoc.Name))
		return nil
	}

	enquiryID := firstFieldValue(fields, "queryId", "enquiryId", "id")
	buyerID := firstFieldValue(fields, "userId", "customerId", "user_id", "buyerId")

	log.Info(fmt.Sprintf("New enquiry document=%s enquiryId=%s sellerId=%s buyerId=%s",
		newDoc.Name, enquiryID, sellerID, buyerID))

	// ── Init notification service (singleton — safe on warm starts) ───────────
	svc, err := notification.NewService(ctx, buildNotifConfig())
	if err != nil {
		log.Error(fmt.Sprintf("Failed to init notification service: %v", err))
		return fmt.Errorf("init notification service: %w", err) // retriable
	}
	defer svc.Close()

	// ── Fetch seller FCM tokens ───────────────────────────────────────────────
	tokens, err := svc.GetOwnerFCMTokens(ctx, "sellers", sellerID)
	if err != nil {
		log.Warn(fmt.Sprintf("Error fetching FCM tokens sellerId=%s: %v — gracefully skipping", sellerID, err))
		return nil // seller token fetch errors are non-fatal
	}
	if len(tokens) == 0 {
		log.Info(fmt.Sprintf("No active FCM tokens for sellerId=%s — no notification sent", sellerID))
		return nil
	}

	// ── Build seller notification copy ───────────────────────────────────────
	title := "New Enquiry Received"
	body := "A buyer has sent a new enquiry. Tap to review and respond."
	for _, key := range []string{"askQuantity", "ask_quantity", "quantity"} {
		if qty := firstFieldValue(fields, key); qty != "" {
			body = fmt.Sprintf("A buyer is enquiring for quantity %s. Tap to respond.", qty)
			break
		}
	}

	data := map[string]string{
		"event_type":     "enquiry_created",
		"seller_id":      sellerID,
		"recipient_type": "seller",
	}
	if enquiryID != "" {
		data["enquiry_id"] = enquiryID
		data["action_url"] = fmt.Sprintf("/enquiry/%s", enquiryID)
	}
	if buyerID != "" {
		data["buyer_id"] = buyerID
	}

	// ── Send notification ─────────────────────────────────────────────────────
	if err := svc.SendToTokens(ctx, tokens, title, body, data); err != nil {
		// Notification failure must not cause Eventarc to re-deliver the event,
		// which would spam the seller. Log and move on.
		log.Error(fmt.Sprintf("SendToTokens failed sellerId=%s: %v — non-fatal", sellerID, err))
		return nil
	}

	log.Info(fmt.Sprintf("ProcessEnquiryCreate done sellerId=%s enquiryId=%s tokens=%d",
		sellerID, enquiryID, len(tokens)))
	return nil
}

// ProcessEnquiryUpdate handles google.cloud.firestore.document.v1.updated events.
func ProcessEnquiryUpdate(ctx context.Context, ce cloudevents.Event) error {
	log := newInvocationLogger(ce.ID())
	log.Info(fmt.Sprintf("ProcessEnquiryUpdate start eventId=%s source=%s", ce.ID(), ce.Source()))

	// Enforce hard per-invocation deadline.
	ctx, cancel := context.WithTimeout(ctx, invocationTimeout())
	defer cancel()

	rawData := ce.Data()
	if len(rawData) == 0 {
		log.Error("CloudEvent data is empty — non-retriable, skipping")
		return nil
	}

	var docEvent firestoredata.DocumentEventData
	if err := proto.Unmarshal(rawData, &docEvent); err != nil {
		log.Error(fmt.Sprintf("proto.Unmarshal failed: %v — non-retriable, skipping", err))
		return nil
	}

	event := convertProtoToFirestoreEvent(&docEvent, ce.ID())
	log.Debug(fmt.Sprintf("Event converted eventId=%s", event.ID))

	return handleEnquiryUpdate(ctx, log, event)
}

// convertProtoToFirestoreEvent maps firestoredata.DocumentEventData → domain.FirestoreEvent.
// UpdateMask is guarded against nil before dereferencing.
func convertProtoToFirestoreEvent(docEvent *firestoredata.DocumentEventData, eventID string) *domain.FirestoreEvent {
	ed := domain.FirestoreEventData{}
	if docEvent.Value != nil {
		ed.Value = convertProtoDocument(docEvent.Value)
	}
	if docEvent.OldValue != nil {
		ed.OldValue = convertProtoDocument(docEvent.OldValue)
	}
	// UpdateMask is optional; guard against nil before accessing FieldPaths.
	if docEvent.UpdateMask != nil && len(docEvent.UpdateMask.FieldPaths) > 0 {
		ed.UpdateMask = &domain.UpdateMask{
			FieldPaths: docEvent.UpdateMask.FieldPaths,
		}
	}
	return &domain.FirestoreEvent{Data: ed, ID: eventID}
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

// handleEnquiryUpdate is the core processing pipeline for enquiry update events.
// It is separated from ProcessEnquiryUpdate to make the logic unit-testable and to
// allow the panic-recovery defer to cover the full business pipeline.
func handleEnquiryUpdate(ctx context.Context, log *invocationLogger, event *domain.FirestoreEvent) (retErr error) {
	// Recover panics with full stack trace.  A bug in any downstream library must
	// not produce continuous Eventarc retries — swallow and log at CRITICAL.
	defer func() {
		if r := recover(); r != nil {
			stack := string(debug.Stack())
			log.Critical(fmt.Sprintf("PANIC recovered in handleEnquiryUpdate: %v\nStackTrace:\n%s", r, stack))
			retErr = nil // not retriable
		}
	}()

	// ── Parse event ───────────────────────────────────────────────────────────
	eventHandler := firestoreutil.NewEventHandler()
	parsedEvent, err := eventHandler.ParseEvent(event)
	if err != nil {
		// Metadata-only or administrative changes often produce events with no
		// document path.  Treat as non-retriable and skip gracefully.
		log.Warn(fmt.Sprintf("ParseEvent failed: %v — likely metadata-only change, skipping", err))
		return nil
	}

	log.Info(fmt.Sprintf("Parsed update documentId=%s updateTime=%s updatedPaths=%d",
		parsedEvent.DocumentID, parsedEvent.UpdateTime, len(parsedEvent.UpdatedPaths)))

	// ── Detect significant field changes ──────────────────────────────────────
	changes := eventHandler.FindChanges(parsedEvent)
	if !eventHandler.HasSignificantChanges(changes) {
		log.Info(fmt.Sprintf("No significant changes documentId=%s — skipping notification",
			parsedEvent.DocumentID))
		return nil
	}

	changedFields := firestoreutil.GetChangedFieldNames(changes)
	log.Info(fmt.Sprintf("Significant changes documentId=%s fields=%v",
		parsedEvent.DocumentID, changedFields))

	// ── Build notification payload ────────────────────────────────────────────
	payloadBuilder := notification.NewPayloadBuilder()
	payload := payloadBuilder.BuildPayload(parsedEvent, changes)

	if err := notification.ValidatePayload(payload); err != nil {
		log.Error(fmt.Sprintf("Invalid notification payload documentId=%s: %v — skipping",
			parsedEvent.DocumentID, err))
		return nil // invalid payload is not retriable
	}

	log.Debug(fmt.Sprintf("Payload ready title=%q enquiryId=%s userId=%s sellerId=%s",
		payload.Title, payload.EnquiryID, payload.UserID, payload.SellerID))

	// ── Init notification service (singleton; safe on warm starts) ────────────
	svc, err := notification.NewService(ctx, buildNotifConfig())
	if err != nil {
		log.Error(fmt.Sprintf("Failed to init notification service: %v", err))
		return fmt.Errorf("init notification service: %w", err) // retriable
	}
	defer svc.Close()

	// ── Deliver notifications ─────────────────────────────────────────────────
	log.Info(fmt.Sprintf("Sending notification documentId=%s enquiryId=%s sellerId=%s",
		parsedEvent.DocumentID, payload.EnquiryID, payload.SellerID))
	if err := svc.SendNotification(ctx, parsedEvent, changes, payload); err != nil {
		// A persistent FCM error must not trigger replay — it would spam users.
		log.Error(fmt.Sprintf("SendNotification failed documentId=%s: %v — non-fatal",
			parsedEvent.DocumentID, err))
		return nil
	}

	log.Info(fmt.Sprintf("ProcessEnquiryUpdate done documentId=%s", parsedEvent.DocumentID))
	return nil
}

// Logger methods — kept as no-ops to satisfy any remaining references during
// the transition period; the real logging is done via invocationLogger above.
