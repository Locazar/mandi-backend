// Package main implements a generic Google Cloud Function (Gen 2) that listens to
// Firestore document update events via Eventarc and sends Firebase Cloud Messaging
// (FCM) push notifications to sellers and/or customers.
//
// Supported collection groups (configure via COLLECTION_TYPE env var):
//   - "order"   → orders/{orderId}   : notifies the customer and/or seller
//   - "product" → products/{id}       : notifies the seller
//   - "shop"    → shops/{id}           : notifies the seller
//
// Configuration environment variables:
//
//	COLLECTION_TYPE     - "order" | "product" | "shop"  (required)
//	MONITORED_FIELDS    - comma-separated list of fields to watch (overrides defaults)
//	NOTIFY_USER         - "true" | "false"  (default: true for orders)
//	NOTIFY_SELLER       - "true" | "false"  (default: true for products/shops)
//	FIREBASE_PROJECT_ID - GCP project ID (optional, falls back to ADC project)
//	LOG_LEVEL           - DEBUG | INFO | WARN | ERROR  (default: INFO)
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/messaging"
	"github.com/GoogleCloudPlatform/functions-framework-go/functions"
)

func init() {
	functions.CloudEvent("ProcessEcommerceUpdate", ProcessEcommerceUpdate)
}

// ---------------------------------------------------------------------------
// Domain types (self-contained — no import of parent module)
// ---------------------------------------------------------------------------

// FirestoreDocument is a Firestore document as received in a CloudEvent payload.
type FirestoreDocument struct {
	Name       string                 `json:"name"`
	Fields     map[string]interface{} `json:"fields"`
	CreateTime string                 `json:"createTime"`
	UpdateTime string                 `json:"updateTime"`
}

// UpdateMask lists which field paths were modified.
type UpdateMask struct {
	FieldPaths []string `json:"fieldPaths"`
}

// FirestoreEventData is the "data" portion of the CloudEvent.
type FirestoreEventData struct {
	Value      *FirestoreDocument `json:"value"`
	OldValue   *FirestoreDocument `json:"oldValue"`
	UpdateMask *UpdateMask        `json:"updateMask"`
}

// FieldChange captures a single field's old and new value.
type FieldChange struct {
	Field    string
	OldValue interface{}
	NewValue interface{}
}

// NotificationTemplate defines what to send for a given field change.
type NotificationTemplate struct {
	Title     string
	Body      string
	EventType string
}

// ---------------------------------------------------------------------------
// Globals (single initialisation per cold start)
// ---------------------------------------------------------------------------

var (
	msgClient *messaging.Client
	fsClient  *firestore.Client
)

func initFirebase(ctx context.Context) error {
	if msgClient != nil {
		return nil // already initialised
	}

	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return fmt.Errorf("firebase.NewApp: %w", err)
	}

	msgClient, err = app.Messaging(ctx)
	if err != nil {
		return fmt.Errorf("app.Messaging: %w", err)
	}

	fsClient, err = app.Firestore(ctx)
	if err != nil {
		return fmt.Errorf("app.Firestore: %w", err)
	}

	return nil
}

// ---------------------------------------------------------------------------
// Cloud Function entry point
// ---------------------------------------------------------------------------

// ProcessEcommerceUpdate is the Eventarc-triggered Cloud Function.
func ProcessEcommerceUpdate(ctx context.Context, cloudEvent interface{}) error {
	logInfo("ProcessEcommerceUpdate triggered")

	if err := initFirebase(ctx); err != nil {
		return fmt.Errorf("firebase init: %w", err)
	}

	// Marshal / unmarshal to get a typed struct from the raw interface{}
	ceMap, ok := cloudEvent.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected CloudEvent type %T", cloudEvent)
	}

	dataBytes, err := json.Marshal(ceMap["data"])
	if err != nil {
		return fmt.Errorf("marshal event data: %w", err)
	}

	var eventData FirestoreEventData
	if err := json.Unmarshal(dataBytes, &eventData); err != nil {
		return fmt.Errorf("unmarshal event data: %w", err)
	}

	if eventData.Value == nil {
		logWarn("event has no current value – skipping")
		return nil
	}

	monitoredFields := getMonitoredFields()
	changes := findChanges(eventData, monitoredFields)
	if len(changes) == 0 {
		logInfo("no monitored fields changed – skipping")
		return nil
	}

	docPath := documentPath(eventData.Value.Name)
	docID := documentID(docPath)
	newFields := parseFields(eventData.Value.Fields)

	logInfo(fmt.Sprintf("processing %s (id=%s, changes=%d)", docPath, docID, len(changes)))

	collectionType := strings.ToLower(os.Getenv("COLLECTION_TYPE"))
	return dispatchNotifications(ctx, collectionType, docID, newFields, changes)
}

// ---------------------------------------------------------------------------
// Dispatch logic per collection type
// ---------------------------------------------------------------------------

func dispatchNotifications(ctx context.Context, collType, docID string, fields map[string]interface{}, changes []FieldChange) error {
	switch collType {
	case "order":
		return handleOrderUpdate(ctx, docID, fields, changes)
	case "product":
		return handleProductUpdate(ctx, docID, fields, changes)
	case "shop":
		return handleShopUpdate(ctx, docID, fields, changes)
	default:
		logWarn(fmt.Sprintf("unknown COLLECTION_TYPE=%q – using generic handler", collType))
		return handleGenericUpdate(ctx, docID, fields, changes)
	}
}

// handleOrderUpdate sends notifications when an order's status changes.
// Customer and/or seller are notified based on NOTIFY_USER / NOTIFY_SELLER env vars.
func handleOrderUpdate(ctx context.Context, orderID string, fields map[string]interface{}, changes []FieldChange) error {
	status := getStringField(fields, "status", "orderStatus")
	userID := getStringField(fields, "userId", "customerId", "user_id")
	sellerID := getStringField(fields, "shopId", "sellerId", "shop_id")

	tmpl := orderStatusTemplate(status, orderID)
	var lastErr error

	if envBool("NOTIFY_USER", true) && userID != "" {
		if err := sendToOwner(ctx, "users", userID, tmpl.Title, tmpl.Body, map[string]string{
			"event_type": tmpl.EventType,
			"order_id":   orderID,
			"status":     status,
		}); err != nil {
			logWarn(fmt.Sprintf("failed to notify user %s: %v", userID, err))
			lastErr = err
		}
	}

	if envBool("NOTIFY_SELLER", true) && sellerID != "" {
		sellerTmpl := sellerOrderTemplate(status, orderID)
		if err := sendToOwner(ctx, "sellers", sellerID, sellerTmpl.Title, sellerTmpl.Body, map[string]string{
			"event_type": sellerTmpl.EventType,
			"order_id":   orderID,
			"status":     status,
		}); err != nil {
			logWarn(fmt.Sprintf("failed to notify seller %s: %v", sellerID, err))
			lastErr = err
		}
	}

	return lastErr
}

// handleProductUpdate notifies the seller when product fields change (e.g. stockStatus, verificationStatus).
func handleProductUpdate(ctx context.Context, productID string, fields map[string]interface{}, changes []FieldChange) error {
	sellerID := getStringField(fields, "sellerId", "shopId", "shop_id", "seller_id")
	if sellerID == "" {
		logInfo(fmt.Sprintf("product %s: no seller ID found – skipping", productID))
		return nil
	}

	for _, c := range changes {
		tmpl := productChangeTemplate(c.Field, fmt.Sprint(c.NewValue), productID)
		if err := sendToOwner(ctx, "sellers", sellerID, tmpl.Title, tmpl.Body, map[string]string{
			"event_type": tmpl.EventType,
			"product_id": productID,
			"field":      c.Field,
		}); err != nil {
			logWarn(fmt.Sprintf("failed to notify seller %s: %v", sellerID, err))
		}
	}
	return nil
}

// handleShopUpdate notifies the shop owner on key field changes (e.g. verificationStatus, blockStatus).
func handleShopUpdate(ctx context.Context, shopID string, fields map[string]interface{}, changes []FieldChange) error {
	ownerID := getStringField(fields, "ownerId", "adminId", "owner_id", "admin_id")
	if ownerID == "" {
		logInfo(fmt.Sprintf("shop %s: no owner ID found – skipping", shopID))
		return nil
	}

	for _, c := range changes {
		tmpl := shopChangeTemplate(c.Field, fmt.Sprint(c.NewValue), shopID)
		if err := sendToOwner(ctx, "sellers", ownerID, tmpl.Title, tmpl.Body, map[string]string{
			"event_type": tmpl.EventType,
			"shop_id":    shopID,
			"field":      c.Field,
		}); err != nil {
			logWarn(fmt.Sprintf("failed to notify shop owner %s: %v", ownerID, err))
		}
	}
	return nil
}

// handleGenericUpdate sends the "fields changed" summary to whoever is referenced in the doc.
func handleGenericUpdate(ctx context.Context, docID string, fields map[string]interface{}, changes []FieldChange) error {
	ownerID := getStringField(fields, "userId", "ownerId", "adminId")
	ownerType := "users"
	if ot := getStringField(fields, "ownerType", "owner_type"); ot == "seller" {
		ownerType = "sellers"
	}

	if ownerID == "" {
		return nil
	}

	fieldNames := make([]string, 0, len(changes))
	for _, c := range changes {
		fieldNames = append(fieldNames, c.Field)
	}

	return sendToOwner(ctx, ownerType, ownerID,
		"Record Updated",
		fmt.Sprintf("Fields updated: %s", strings.Join(fieldNames, ", ")),
		map[string]string{
			"event_type": "document_updated",
			"doc_id":     docID,
		},
	)
}

// ---------------------------------------------------------------------------
// FCM delivery
// ---------------------------------------------------------------------------

// sendToOwner fetches active FCM tokens from Firestore and sends the message.
// ownerCollection is "users" or "sellers".
func sendToOwner(ctx context.Context, ownerCollection, ownerID, title, body string, data map[string]string) error {
	tokens, err := getActiveTokens(ctx, ownerCollection, ownerID)
	if err != nil {
		return fmt.Errorf("getActiveTokens: %w", err)
	}
	if len(tokens) == 0 {
		logInfo(fmt.Sprintf("no active tokens for %s/%s", ownerCollection, ownerID))
		return nil
	}

	if data == nil {
		data = map[string]string{}
	}
	data["timestamp"] = time.Now().UTC().Format(time.RFC3339)

	msg := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title: title,
			Body:  body,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{Title: title, Body: body},
					Sound: "default",
				},
			},
		},
	}

	resp, err := msgClient.SendEachForMulticast(ctx, msg)
	if err != nil {
		return fmt.Errorf("SendEachForMulticast: %w", err)
	}

	logInfo(fmt.Sprintf("FCM: sent %d/%d to %s/%s", resp.SuccessCount, len(tokens), ownerCollection, ownerID))

	// Clean up invalid tokens
	for i, r := range resp.Responses {
		if !r.Success {
			code := messaging.IsRegistrationTokenNotRegistered(r.Error)
			if code {
				_ = deactivateToken(ctx, ownerCollection, ownerID, tokens[i])
			}
		}
	}
	return nil
}

// getActiveTokens retrieves all active FCM tokens from Firestore.
// Path: {collection}/{ownerID}/fcmTokens (where isActive == true)
func getActiveTokens(ctx context.Context, collection, ownerID string) ([]string, error) {
	docs, err := fsClient.
		Collection(collection).
		Doc(ownerID).
		Collection("fcmTokens").
		Where("isActive", "==", true).
		Documents(ctx).
		GetAll()
	if err != nil {
		return nil, err
	}

	tokens := make([]string, 0, len(docs))
	for _, d := range docs {
		if t, ok := d.Data()["token"].(string); ok && t != "" {
			tokens = append(tokens, t)
		}
	}
	return tokens, nil
}

// deactivateToken marks an invalid token as inactive so it isn't tried again.
func deactivateToken(ctx context.Context, collection, ownerID, token string) error {
	_, err := fsClient.
		Collection(collection).
		Doc(ownerID).
		Collection("fcmTokens").
		Doc(token).
		Update(ctx, []firestore.Update{
			{Path: "isActive", Value: false},
			{Path: "updatedAt", Value: firestore.ServerTimestamp},
		})
	return err
}

// ---------------------------------------------------------------------------
// Notification templates
// ---------------------------------------------------------------------------

func orderStatusTemplate(status, orderID string) NotificationTemplate {
	m := map[string]NotificationTemplate{
		"order placed":     {Title: "Order Confirmed!", Body: fmt.Sprintf("Your order #%s is being processed.", orderID), EventType: "order_placed"},
		"payment pending":  {Title: "Payment Required", Body: fmt.Sprintf("Complete payment for order #%s.", orderID), EventType: "payment_pending"},
		"order cancelled":  {Title: "Order Cancelled", Body: fmt.Sprintf("Your order #%s has been cancelled.", orderID), EventType: "order_cancelled"},
		"order delivered":  {Title: "Order Delivered!", Body: fmt.Sprintf("Your order #%s has arrived. Enjoy!", orderID), EventType: "order_delivered"},
		"return requested": {Title: "Return Requested", Body: fmt.Sprintf("Return for order #%s is being reviewed.", orderID), EventType: "return_requested"},
		"return approved":  {Title: "Return Approved", Body: fmt.Sprintf("Your return for order #%s is approved.", orderID), EventType: "return_approved"},
		"return cancelled": {Title: "Return Declined", Body: fmt.Sprintf("Return for order #%s was declined.", orderID), EventType: "return_cancelled"},
		"order returned":   {Title: "Order Returned", Body: fmt.Sprintf("Your order #%s has been returned. Refund initiated.", orderID), EventType: "order_returned"},
	}
	if t, ok := m[strings.ToLower(status)]; ok {
		return t
	}
	return NotificationTemplate{
		Title:     "Order Update",
		Body:      fmt.Sprintf("Your order #%s status changed to: %s", orderID, status),
		EventType: "order_updated",
	}
}

func sellerOrderTemplate(status, orderID string) NotificationTemplate {
	m := map[string]NotificationTemplate{
		"order placed":    {Title: "New Order Received!", Body: fmt.Sprintf("Order #%s placed. Prepare for dispatch.", orderID), EventType: "seller_new_order"},
		"order cancelled": {Title: "Order Cancelled", Body: fmt.Sprintf("Order #%s has been cancelled by the customer.", orderID), EventType: "seller_order_cancelled"},
		"return requested": {Title: "Return Requested", Body: fmt.Sprintf("Customer requested a return for order #%s.", orderID), EventType: "seller_return_requested"},
	}
	if t, ok := m[strings.ToLower(status)]; ok {
		return t
	}
	return NotificationTemplate{Title: "Order Status Changed", Body: fmt.Sprintf("Order #%s: %s", orderID, status), EventType: "seller_order_updated"}
}

func productChangeTemplate(field, newValue, productID string) NotificationTemplate {
	switch strings.ToLower(field) {
	case "verificationstatus", "verification_status":
		return NotificationTemplate{
			Title:     "Product Verification Update",
			Body:      fmt.Sprintf("Your product #%s verification status: %s", productID, newValue),
			EventType: "product_verification_updated",
		}
	case "blockstatus", "block_status":
		if strings.ToLower(newValue) == "true" {
			return NotificationTemplate{Title: "Product Blocked", Body: fmt.Sprintf("Product #%s has been blocked.", productID), EventType: "product_blocked"}
		}
		return NotificationTemplate{Title: "Product Unblocked", Body: fmt.Sprintf("Product #%s is now active.", productID), EventType: "product_unblocked"}
	case "stockstatus", "stock_status", "stockquantity", "stock_quantity":
		return NotificationTemplate{Title: "Stock Alert", Body: fmt.Sprintf("Stock for product #%s updated.", productID), EventType: "product_stock_updated"}
	}
	return NotificationTemplate{Title: "Product Updated", Body: fmt.Sprintf("Product #%s has been updated.", productID), EventType: "product_updated"}
}

func shopChangeTemplate(field, newValue, shopID string) NotificationTemplate {
	switch strings.ToLower(field) {
	case "verificationstatus", "verification_status":
		return NotificationTemplate{
			Title:     "Shop Verification Update",
			Body:      fmt.Sprintf("Your shop verification status changed to: %s", newValue),
			EventType: "shop_verification_updated",
		}
	case "blockstatus", "block_status":
		if strings.ToLower(newValue) == "true" {
			return NotificationTemplate{Title: "Shop Blocked", Body: "Your shop has been blocked. Contact support.", EventType: "shop_blocked"}
		}
		return NotificationTemplate{Title: "Shop Unblocked", Body: "Your shop is now live!", EventType: "shop_unblocked"}
	}
	return NotificationTemplate{Title: "Shop Updated", Body: fmt.Sprintf("Shop #%s details have changed.", shopID), EventType: "shop_updated"}
}

// ---------------------------------------------------------------------------
// Firestore field parsing utilities
// ---------------------------------------------------------------------------

// parseFields converts a raw Firestore "fields" map to a flat Go map.
func parseFields(raw map[string]interface{}) map[string]interface{} {
	out := make(map[string]interface{}, len(raw))
	for k, v := range raw {
		out[k] = extractValue(v)
	}
	return out
}

// extractValue extracts a typed value from Firestore's "typed value" format.
func extractValue(v interface{}) interface{} {
	m, ok := v.(map[string]interface{})
	if !ok {
		return v
	}
	for key, val := range m {
		switch key {
		case "stringValue":
			return fmt.Sprint(val)
		case "integerValue":
			return val
		case "doubleValue":
			return val
		case "booleanValue":
			return val
		case "nullValue":
			return nil
		case "timestampValue":
			return fmt.Sprint(val)
		case "mapValue":
			if nested, ok := val.(map[string]interface{}); ok {
				if fields, ok := nested["fields"].(map[string]interface{}); ok {
					return parseFields(fields)
				}
			}
		}
	}
	return v
}

// findChanges returns only the monitored fields that changed value.
func findChanges(event FirestoreEventData, monitored []string) []FieldChange {
	if event.Value == nil || event.OldValue == nil {
		return nil
	}

	newF := parseFields(event.Value.Fields)
	oldF := parseFields(event.OldValue.Fields)

	// If UpdateMask is provided, restrict to those paths
	watchSet := make(map[string]bool, len(monitored))
	for _, f := range monitored {
		watchSet[strings.ToLower(f)] = true
	}

	changes := []FieldChange{}
	for field, newVal := range newF {
		if len(watchSet) > 0 && !watchSet[strings.ToLower(field)] {
			continue
		}
		oldVal := oldF[field]
		if fmt.Sprint(newVal) != fmt.Sprint(oldVal) {
			changes = append(changes, FieldChange{Field: field, OldValue: oldVal, NewValue: newVal})
		}
	}
	return changes
}

// getMonitoredFields reads from MONITORED_FIELDS env var or returns collection defaults.
func getMonitoredFields() []string {
	if v := os.Getenv("MONITORED_FIELDS"); v != "" {
		parts := strings.Split(v, ",")
		fields := make([]string, 0, len(parts))
		for _, p := range parts {
			if trimmed := strings.TrimSpace(p); trimmed != "" {
				fields = append(fields, trimmed)
			}
		}
		return fields
	}

	// Defaults per collection type
	switch strings.ToLower(os.Getenv("COLLECTION_TYPE")) {
	case "order":
		return []string{"status", "orderStatus", "paymentStatus"}
	case "product":
		return []string{"verificationStatus", "blockStatus", "stockStatus", "stockQuantity"}
	case "shop":
		return []string{"verificationStatus", "blockStatus", "isActive"}
	}
	return nil // monitor all fields
}

// getStringField returns the value of the first matching key found in fields.
func getStringField(fields map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		if v, ok := fields[k]; ok {
			if s := fmt.Sprint(v); s != "" && s != "<nil>" {
				return s
			}
		}
	}
	return ""
}

// documentPath extracts "collection/docId" from a full Firestore resource name.
// e.g. "projects/p/databases/(default)/documents/orders/abc123" → "orders/abc123"
func documentPath(resourceName string) string {
	const docSep = "/documents/"
	if idx := strings.Index(resourceName, docSep); idx != -1 {
		return resourceName[idx+len(docSep):]
	}
	return resourceName
}

// documentID extracts the final segment from the document path.
func documentID(docPath string) string {
	parts := strings.Split(docPath, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return docPath
}

// envBool reads a boolean env var with a fallback default.
func envBool(key string, defaultVal bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "true" || v == "1" {
		return true
	}
	if v == "false" || v == "0" {
		return false
	}
	return defaultVal
}

// ---------------------------------------------------------------------------
// Logging helpers
// ---------------------------------------------------------------------------

func logInfo(msg string) {
	if level() <= 1 {
		log.Printf("[INFO] %s", msg)
	}
}

func logWarn(msg string) {
	if level() <= 2 {
		log.Printf("[WARN] %s", msg)
	}
}

func level() int {
	switch strings.ToUpper(os.Getenv("LOG_LEVEL")) {
	case "DEBUG":
		return 0
	case "INFO":
		return 1
	case "WARN":
		return 2
	case "ERROR":
		return 3
	}
	return 1
}

// main is only used for local testing; Cloud Functions deployment ignores this.
func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("[INFO] Starting local function server on :%s", port)

	// The functions framework handles HTTP serving when deployed.
	// Locally: FUNCTION_TARGET=ProcessEcommerceUpdate go run main.go
}
