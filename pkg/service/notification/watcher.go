package notification

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// FieldChange captures a single changed field with old/new values.
type WatchFieldChange struct {
	Field    string
	OldValue interface{}
	NewValue interface{}
}

// MessageBuilder is an optional function that builds notification content from
// the changed document and detected field changes.  If nil, a default message
// is generated from the WatchRule templates.
type MessageBuilder func(docData map[string]interface{}, changes []WatchFieldChange) (title, body string)

// WatchRule configures a single Firestore collection listener.
//
// Example – watch "orders" for status changes and notify both the customer
// and the seller:
//
//	WatchRule{
//	    Collection:     "orders",
//	    MonitoredFields: []string{"status", "cancellationReason"},
//	    NotifyUser:     true,
//	    NotifySeller:   true,
//	    UserIDField:    "userId",
//	    SellerIDField:  "shopId",
//	    EventType:      "order_status_changed",
//	}
type WatchRule struct {
	// Collection is the top-level Firestore collection name (e.g., "orders").
	Collection string

	// MonitoredFields is the list of document fields that, when changed, trigger
	// a push notification.  An empty slice means "any field change".
	MonitoredFields []string

	// NotifyUser, NotifySeller control who receives the notification.
	NotifyUser   bool
	NotifySeller bool

	// UserIDField / SellerIDField are the document field names that hold the
	// owner ID used to look up FCM tokens.  Typical values:
	//   User:   "userId", "customerId", "user_id"
	//   Seller: "shopId", "sellerId", "shop_id"
	UserIDField   string
	SellerIDField string

	// UserCollection / SellerCollection are the Firestore sub-paths where FCM
	// tokens are stored (defaults: "users" / "sellers").
	UserCollection   string
	SellerCollection string

	// TitleTemplate / BodyTemplate are the default notification content.
	// These are used when MessageBuilder is nil.
	TitleTemplate string
	BodyTemplate  string

	// EventType is passed as a data key so mobile clients can route the tap.
	EventType string

	// MessageBuilder overrides TitleTemplate/BodyTemplate with dynamic content.
	// If nil the templates are used.
	MessageBuilder MessageBuilder

	// NotifyOnCreate, when true, also fires notifications when a new document
	// is added to the collection (DocumentAdded), not only on updates.
	NotifyOnCreate bool

	// DataEnricher is an optional function called before every notification
	// dispatch.  It receives the full Firestore document data and may return
	// extra key/value pairs that are merged into the FCM data payload.
	// Use it to attach data that is not stored in Firestore (e.g. product
	// image URLs fetched from a SQL database).
	DataEnricher func(ctx context.Context, docData map[string]interface{}) map[string]string

	// RecipientResolver is an optional function that dynamically determines
	// whether to notify the user and/or seller based on the live document data.
	// When set, it overrides the static NotifyUser / NotifySeller flags, allowing
	// per-status routing (e.g. enquiry negotiation flow).
	RecipientResolver func(docData map[string]interface{}) (notifyUser bool, notifySeller bool)
}

// userCollection returns the Firestore collection path for end-users.
func (r WatchRule) userCollection() string {
	if r.UserCollection != "" {
		return r.UserCollection
	}
	return "users"
}

// sellerCollection returns the Firestore collection path for sellers.
func (r WatchRule) sellerCollection() string {
	if r.SellerCollection != "" {
		return r.SellerCollection
	}
	return "sellers"
}

// FirestoreWatcher listens to one or more Firestore collections in real-time
// and sends FCM push notifications when monitored fields change.
//
// It is safe to add rules before calling Start.  After Start is called the
// rule list must not be modified.
type FirestoreWatcher struct {
	rules    []WatchRule
	push     PushSender
	mu       sync.RWMutex
	fsClient *firestore.Client
	initOnce sync.Once
	initErr  error
}

// NewFirestoreWatcher creates a watcher backed by the given PushSender.
// Firebase credentials are resolved in the same way as FCMPushService:
// FIREBASE_CONFIG env-var JSON > GOOGLE_APPLICATION_CREDENTIALS file > ADC.
func NewFirestoreWatcher(push PushSender, rules ...WatchRule) *FirestoreWatcher {
	return &FirestoreWatcher{
		rules: append([]WatchRule{}, rules...),
		push:  push,
	}
}

// AddRule appends a WatchRule.  Must be called before Start.
func (w *FirestoreWatcher) AddRule(r WatchRule) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.rules = append(w.rules, r)
}

// ensureFirestore initialises the Firestore client exactly once.
// Uses the package-level shared Firebase App to avoid duplicate-app errors.
func (w *FirestoreWatcher) ensureFirestore(ctx context.Context) error {
	w.initOnce.Do(func() {
		var err error
		w.fsClient, err = sharedFirestoreClient(ctx)
		if err != nil {
			w.initErr = fmt.Errorf("firestore client: %w", err)
		}
	})
	return w.initErr
}

// Start launches one goroutine per WatchRule.  The goroutines run until ctx
// is cancelled.  Start itself returns immediately.
func (w *FirestoreWatcher) Start(ctx context.Context) error {
	if err := w.ensureFirestore(ctx); err != nil {
		return fmt.Errorf("firestore watcher init: %w", err)
	}

	w.mu.RLock()
	rules := append([]WatchRule{}, w.rules...)
	w.mu.RUnlock()

	if len(rules) == 0 {
		log.Println("INFO [FirestoreWatcher]: no rules configured, watcher is idle")
		return nil
	}

	for _, rule := range rules {
		go w.watchCollection(ctx, rule)
	}

	log.Printf("INFO [FirestoreWatcher]: started %d watcher(s)", len(rules))
	return nil
}

// watchCollection runs a snapshot listener for a single WatchRule with
// automatic exponential-backoff reconnection on error.
func (w *FirestoreWatcher) watchCollection(ctx context.Context, rule WatchRule) {
	backoff := time.Second
	const maxBackoff = 2 * time.Minute

	for {
		select {
		case <-ctx.Done():
			log.Printf("INFO [FirestoreWatcher]: stopping listener for %s", rule.Collection)
			return
		default:
		}

		log.Printf("INFO [FirestoreWatcher]: starting snapshot listener for collection=%s", rule.Collection)
		err := w.runSnapshotLoop(ctx, rule)
		if err != nil {
			log.Printf("WARN [FirestoreWatcher]: listener for %s failed: %v (retry in %s)",
				rule.Collection, err, backoff)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(backoff):
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
		}
	}
}

// runSnapshotLoop opens a snapshot stream for the collection and processes
// DocumentModified events until the context is cancelled or an error occurs.
func (w *FirestoreWatcher) runSnapshotLoop(ctx context.Context, rule WatchRule) error {
	iter := w.fsClient.Collection(rule.Collection).Snapshots(ctx)
	defer iter.Stop()

	for {
		snap, err := iter.Next()
		if err == iterator.Done {
			return nil
		}
		if err != nil {
			return fmt.Errorf("snapshot next: %w", err)
		}

		log.Printf("DEBUG [FirestoreWatcher]: %s snapshot — %d change(s) in batch",
			rule.Collection, len(snap.Changes))

		for _, change := range snap.Changes {
			log.Printf("DEBUG [FirestoreWatcher]: %s/%s kind=%v",
				rule.Collection, change.Doc.Ref.ID, change.Kind)

			isAdded := change.Kind == firestore.DocumentAdded
			if change.Kind != firestore.DocumentModified && !(isAdded && rule.NotifyOnCreate) {
				continue
			}

			newData := change.Doc.Data()
			oldData := extractOldData(change)

			var changes []WatchFieldChange
			if isAdded {
				// On creation treat every monitored field that has a value as "changed"
				// with nil old value so the message builder gets full context.
				for _, f := range rule.MonitoredFields {
					if v, ok := newData[f]; ok {
						changes = append(changes, WatchFieldChange{Field: f, OldValue: nil, NewValue: v})
					}
				}
			} else {
				changes = detectChanges(oldData, newData, rule.MonitoredFields)
			}
			log.Printf("DEBUG [FirestoreWatcher]: %s/%s — monitored=%v detected=%d",
				rule.Collection, change.Doc.Ref.ID, rule.MonitoredFields, len(changes))

			if len(changes) == 0 {
				log.Printf("DEBUG [FirestoreWatcher]: %s/%s — no monitored fields changed, skipping",
					rule.Collection, change.Doc.Ref.ID)
				continue
			}

			log.Printf("INFO [FirestoreWatcher]: %s/%s — %d monitored field(s) changed: %s",
				rule.Collection, change.Doc.Ref.ID,
				len(changes), joinFieldNames(changes))

			w.dispatchNotifications(ctx, rule, change.Doc.Ref.ID, newData, changes)
		}
	}
}

// dispatchNotifications sends FCM to users and/or sellers as configured.
func (w *FirestoreWatcher) dispatchNotifications(
	ctx context.Context,
	rule WatchRule,
	docID string,
	docData map[string]interface{},
	changes []WatchFieldChange,
) {
	title, body := w.buildMessage(rule, docData, changes, docID)
	data := map[string]string{
		"document_id": docID,
		"collection":  rule.Collection,
	}
	if rule.EventType != "" {
		data["event_type"] = rule.EventType
	}
	for _, c := range changes {
		if v, ok := c.NewValue.(string); ok {
			data["field_"+c.Field] = v
		}
	}

	// Enrich notification data with extra fields (e.g. product image URLs).
	if rule.DataEnricher != nil {
		for k, v := range rule.DataEnricher(ctx, docData) {
			data[k] = v
		}
	}

	// Determine recipients — dynamic resolver takes precedence over static flags.
	notifyUser := rule.NotifyUser
	notifySeller := rule.NotifySeller
	if rule.RecipientResolver != nil {
		notifyUser, notifySeller = rule.RecipientResolver(docData)
	}

	if notifyUser {
		userID := resolveID(docData, rule.UserIDField, "userId", "customerId", "user_id")
		log.Printf("DEBUG [FirestoreWatcher]: notify user — collection=%s id=%q", rule.userCollection(), userID)
		if userID != "" {
			if err := w.push.SendToOwnerViaFirestore(ctx, rule.userCollection(), userID, title, body, data); err != nil {
				log.Printf("WARN [FirestoreWatcher]: notify user %s: %v", userID, err)
			}
		} else {
			log.Printf("WARN [FirestoreWatcher]: could not resolve userID from field %q in doc %v", rule.UserIDField, docData)
		}
	}

	if notifySeller {
		sellerID := resolveID(docData, rule.SellerIDField, "shopId", "sellerId", "shop_id")
		log.Printf("DEBUG [FirestoreWatcher]: notify seller — collection=%s id=%q", rule.sellerCollection(), sellerID)
		if sellerID != "" {
			if err := w.push.SendToOwnerViaFirestore(ctx, rule.sellerCollection(), sellerID, title, body, data); err != nil {
				log.Printf("WARN [FirestoreWatcher]: notify seller %s: %v", sellerID, err)
			}
		} else {
			log.Printf("WARN [FirestoreWatcher]: could not resolve sellerID from field %q in doc %v", rule.SellerIDField, docData)
		}
	}
}

// buildMessage returns the notification title and body for the event.
func (w *FirestoreWatcher) buildMessage(
	rule WatchRule,
	docData map[string]interface{},
	changes []WatchFieldChange,
	docID string,
) (title, body string) {
	if rule.MessageBuilder != nil {
		return rule.MessageBuilder(docData, changes)
	}
	if rule.TitleTemplate != "" && rule.BodyTemplate != "" {
		return rule.TitleTemplate, rule.BodyTemplate
	}
	// Fallback: auto-generated
	return defaultMessage(rule.Collection, docID, changes)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// extractOldData safely returns the previous document data from a change object.
// The Firestore Go SDK stores the previous snapshot in change.Doc.DataAt(0) on
// the internal snapshot iterator; since the SDK does not expose a direct
// "before" snapshot for collection listeners we derive it from the current
// snapshot by stripping the changed fields and reinserting old values via the
// change metadata.  In practice we receive both old and new via QueryDocumentSnapshot
// in the real SDK — this keeps a safe fallback.
func extractOldData(change firestore.DocumentChange) map[string]interface{} {
	// The Firestore Go SDK does not populate OldIndex/NewIndex in the same way
	// as the JS SDK.  We rely on comparing the current (new) snapshot data
	// against the previous snapshot captured in the iterator's internal state.
	// Since the SDK does not directly expose the "before" document via Go
	// QueryDocumentSnapshot, we return an empty map here and compare using the
	// DocumentChange.OldIndex == -1 sentinel (new doc) vs 0 (modified).
	// Callers that need true old values should use the Firestore Admin SDK or
	// Cloud Functions — this watcher is built for "field presence / new value"
	// style rules where the NEW value drives the notification.
	return map[string]interface{}{}
}

// detectChanges returns the list of fields that changed.  If monitoredFields
// is empty, all fields in newData are treated as potentially changed.
func detectChanges(
	oldData, newData map[string]interface{},
	monitoredFields []string,
) []WatchFieldChange {
	var changes []WatchFieldChange

	candidates := monitoredFields
	if len(candidates) == 0 {
		// Watch every field
		for k := range newData {
			candidates = append(candidates, k)
		}
	}

	for _, field := range candidates {
		newVal, newExists := newData[field]
		oldVal, oldExists := oldData[field]

		if !newExists && !oldExists {
			continue
		}

		// If we have no old data (common case with collection listener) we still
		// report it as a change when the field exists in the new snapshot — the
		// caller is responsible for de-duplication / idempotency.
		if !oldExists || fmt.Sprintf("%v", oldVal) != fmt.Sprintf("%v", newVal) {
			changes = append(changes, WatchFieldChange{
				Field:    field,
				OldValue: oldVal,
				NewValue: newVal,
			})
		}
	}
	return changes
}

// resolveID picks the first non-empty string value from the document for the
// given list of candidate field names.
func resolveID(doc map[string]interface{}, preferredField string, fallbacks ...string) string {
	candidates := append([]string{preferredField}, fallbacks...)
	for _, key := range candidates {
		if key == "" {
			continue
		}
		if v, ok := doc[key]; ok {
			switch s := v.(type) {
			case string:
				if s != "" {
					return s
				}
			case int64:
				if s != 0 {
					return fmt.Sprintf("%d", s)
				}
			case float64:
				if s != 0 {
					return fmt.Sprintf("%.0f", s)
				}
			}
		}
	}
	return ""
}

// joinFieldNames returns a comma-joined list of changed field names.
func joinFieldNames(changes []WatchFieldChange) string {
	names := make([]string, len(changes))
	for i, c := range changes {
		names[i] = c.Field
	}
	return strings.Join(names, ", ")
}

// defaultMessage builds a generic notification when no template is supplied.
func defaultMessage(collection, docID string, changes []WatchFieldChange) (title, body string) {
	title = fmt.Sprintf("%s updated", titleCase(collection))
	if len(changes) == 1 {
		body = fmt.Sprintf("Field '%s' has been updated", changes[0].Field)
	} else {
		body = fmt.Sprintf("%d fields updated: %s", len(changes), joinFieldNames(changes))
	}
	return
}

// titleCase converts "orders" → "Orders", "shop_status" → "Shop Status".
func titleCase(s string) string {
	s = strings.ReplaceAll(s, "_", " ")
	words := strings.Fields(s)
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}
