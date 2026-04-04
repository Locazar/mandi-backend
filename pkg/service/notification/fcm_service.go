package notification

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/db"
	"firebase.google.com/go/v4/messaging"
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"google.golang.org/grpc/codes"
	grpcstatus "google.golang.org/grpc/status"
)

// Config holds notification service configuration
type Config struct {
	ProjectID                     string
	EnableIdempotencyCheck        bool
	FCMTokenCollection            string
	NotificationHistoryCollection string
	FirestoreTimeout              time.Duration
}

// Service handles sending FCM notifications
type Service struct {
	app        *firebase.App
	config     Config
	msgClient  *messaging.Client
	fsClient   *firestore.Client
	rtdbClient *db.Client
}

// singletonService holds the package-level Firebase clients so that Cloud
// Function warm-starts reuse the same connection instead of calling
// firebase.NewApp again (which would fail with "app already exists").
var (
	singletonOnce      sync.Once
	singletonApp       *firebase.App
	singletonMsgClient *messaging.Client
	singletonFsClient  *firestore.Client
	singletonInitErr   error
)

func initSingleton(ctx context.Context) {
	singletonOnce.Do(func() {
		app, err := firebase.NewApp(ctx, nil)
		if err != nil {
			singletonInitErr = fmt.Errorf("failed to init Firebase app: %w", err)
			return
		}
		singletonApp = app

		msgClient, err := app.Messaging(ctx)
		if err != nil {
			singletonInitErr = fmt.Errorf("failed to get messaging client: %w", err)
			return
		}
		singletonMsgClient = msgClient

		fsClient, err := app.Firestore(ctx)
		if err != nil {
			log.Printf("WARN: Firestore client not available: %v", err)
		}
		singletonFsClient = fsClient
	})
}

// NewService creates (or reuses) the notification service.
// Firebase clients are initialised only once per container process so that
// Cloud Function warm-starts do not call firebase.NewApp a second time
// (which would return "default Firebase app already exists").
func NewService(ctx context.Context, config Config) (*Service, error) {
	initSingleton(ctx)
	if singletonInitErr != nil {
		return nil, singletonInitErr
	}

	var rtdbClient *db.Client
	dbURL := os.Getenv("FIREBASE_DB_URL")
	if dbURL != "" {
		var err error
		rtdbClient, err = singletonApp.Database(ctx)
		if err != nil {
			log.Printf("WARN: Realtime DB client not available: %v", err)
		}
	}

	if config.FCMTokenCollection == "" {
		config.FCMTokenCollection = "fcmTokens"
	}
	if config.NotificationHistoryCollection == "" {
		config.NotificationHistoryCollection = "notificationHistory"
	}
	if config.FirestoreTimeout == 0 {
		config.FirestoreTimeout = 10 * time.Second
	}

	return &Service{
		app:        singletonApp,
		config:     config,
		msgClient:  singletonMsgClient,
		fsClient:   singletonFsClient,
		rtdbClient: rtdbClient,
	}, nil
}

// SendNotification sends FCM notification
func (s *Service) SendNotification(ctx context.Context, event *domain.ParsedFirestoreEvent,
	changes []domain.FieldChange, payload *domain.NotificationPayload) error {

	if event == nil || payload == nil {
		return fmt.Errorf("event and payload must not be nil")
	}

	ctx, cancel := context.WithTimeout(ctx, s.config.FirestoreTimeout)
	defer cancel()

	recipients, err := s.GetNotificationRecipients(ctx, event, payload, changes)
	if err != nil {
		return fmt.Errorf("failed to get recipients: %w", err)
	}
	if len(recipients) == 0 {
		log.Printf("INFO: No recipients found for enquiry %s — no notification sent", payload.EnquiryID)
		return nil // no tokens is not an error; graceful skip
	}

	successCount := 0

	for _, recipient := range recipients {
		dedupeKey := ""
		leaseAcquired := true
		if s.config.EnableIdempotencyCheck {
			dedupeKey = buildRecipientDedupeKey(payload, recipient)
			var leaseErr error
			leaseAcquired, leaseErr = s.acquireIdempotencyLease(ctx, dedupeKey, payload, recipient)
			if leaseErr != nil {
				log.Printf("WARN: idempotency lease error recipient=%s key=%s: %v — fail-open", recipient.UserID, dedupeKey, leaseErr)
				// Fail-open: deliver even when idempotency store is temporarily unavailable.
				leaseAcquired = true
			}
			if !leaseAcquired {
				log.Printf("INFO: duplicate notification suppressed recipient=%s key=%s document=%s", recipient.UserID, dedupeKey, payload.DocumentID)
				continue
			}
		}

		// FCM SendEachForMulticast accepts at most 500 tokens per call.
		// Chunk the token list to stay within the API limit.
		recipientSuccessCount := 0
		firstMessageID := ""

		for _, batch := range chunkStrings(recipient.Tokens, fcmMaxTokensPerBatch) {
			multiMsg := s.buildMulticastMessage(batch, payload)
			mresp, sendErr := s.msgClient.SendEachForMulticast(ctx, multiMsg)
			if sendErr != nil {
				log.Printf("WARN: FCM multicast failed for recipient %s: %v", recipient.UserID, sendErr)
				continue
			}
			for i, r := range mresp.Responses {
				if r.Success {
					recipientSuccessCount++
					if firstMessageID == "" {
						firstMessageID = r.MessageID
					}
				} else {
					log.Printf("WARN: FCM token %s failed for recipient %s: %v", batch[i], recipient.UserID, r.Error)
					// Deactivate unregistered / invalid tokens to keep the token store clean.
					if isUnregisteredTokenError(r.Error) {
						s.deactivateOwnerToken(ctx, recipient.Type+"s", recipient.UserID, batch[i])
					}
				}
			}
		}
		successCount += recipientSuccessCount

		if s.config.EnableIdempotencyCheck && dedupeKey != "" {
			if recipientSuccessCount > 0 {
				if err := s.markIdempotencyLeaseSent(ctx, dedupeKey, firstMessageID, recipientSuccessCount); err != nil {
					log.Printf("WARN: Failed to mark idempotency lease as sent key=%s: %v", dedupeKey, err)
				}
			} else if leaseAcquired {
				if err := s.releaseIdempotencyLease(ctx, dedupeKey); err != nil {
					log.Printf("WARN: Failed to release idempotency lease key=%s: %v", dedupeKey, err)
				}
			}
		}
	}

	if successCount == 0 {
		return fmt.Errorf("failed to send notification to any recipient (0 successful deliveries)")
	}
	log.Printf("INFO: Sent %d notifications for enquiry %s", successCount, payload.EnquiryID)
	return nil
}

// GetNotificationRecipients resolves which parties (buyer and/or seller) should
// receive a push for an enquiry event, based on the document's current status.
//
// Routing table (mirrors watcher_rules.go enquiryRecipientResolver):
//
//	pending_seller_price  → seller only   (seller must quote a price)
//	pending_seller_final  → seller only   (seller must send final price)
//	seller_final_update   → seller only   (seller must review buyer's final response)
//	pending_customer_price → buyer only   (buyer must respond to seller's quote)
//	pending_customer_final → buyer only   (buyer must accept/reject seller's final)
//	completed_accepted    → other party   (notify the one who did NOT finalise)
//	completed_rejected    → other party
//	new / no status       → seller only   (new enquiry arrives in seller inbox)
//	any other status      → both parties
func (s *Service) GetNotificationRecipients(ctx context.Context, event *domain.ParsedFirestoreEvent,
	payload *domain.NotificationPayload, changes []domain.FieldChange) ([]*domain.NotificationRecipient, error) {

	recipients := []*domain.NotificationRecipient{}

	newFields := map[string]interface{}{}
	if event != nil && event.NewFields != nil {
		newFields = event.NewFields
	}

	// ── Resolve IDs ───────────────────────────────────────────────────────────
	userID := firstNonEmptyString(newFields, "userId", "customerId", "user_id", "buyerId", "createdBy")
	if userID == "" {
		userID = payload.UserID
	}

	// assignedTo is a support-agent field — do NOT fall through to sellerID.
	sellerID := firstNonEmptyString(newFields, "sellerId", "seller_id", "shopId", "shop_id")
	if sellerID == "" {
		sellerID = payload.SellerID
	}

	// ── Status-based routing ──────────────────────────────────────────────────
	notifyUser, notifySeller := resolveEnquiryRecipients(newFields)

	log.Printf(
		"INFO: enquiry routing document=%s status=%q userID=%q sellerID=%q notifyUser=%v notifySeller=%v changedFields=%s",
		payload.DocumentID,
		firstNonEmptyString(newFields, "status"),
		userID,
		sellerID,
		notifyUser,
		notifySeller,
		joinChangedFieldNames(changes),
	)

	// ── Fetch tokens and build recipient list ─────────────────────────────────
	if notifyUser {
		if userID == "" {
			log.Printf("WARN: notifyUser=true but userID is empty document=%s", payload.DocumentID)
		} else {
			tokens, err := s.GetOwnerFCMTokens(ctx, "users", userID)
			if err != nil {
				log.Printf("WARN: error fetching user tokens userID=%s: %v", userID, err)
			} else if len(tokens) > 0 {
				recipients = append(recipients, &domain.NotificationRecipient{UserID: userID, Tokens: tokens, Type: "user"})
				log.Printf("INFO: found %d token(s) for user %s", len(tokens), userID)
			} else {
				log.Printf("INFO: no active FCM tokens for userID=%s", userID)
			}
		}
	}

	if notifySeller {
		if sellerID == "" {
			log.Printf("WARN: notifySeller=true but sellerID is empty document=%s", payload.DocumentID)
		} else {
			tokens, err := s.GetOwnerFCMTokens(ctx, "sellers", sellerID)
			if err != nil {
				log.Printf("WARN: error fetching seller tokens sellerID=%s: %v", sellerID, err)
			} else if len(tokens) > 0 {
				recipients = append(recipients, &domain.NotificationRecipient{UserID: sellerID, Tokens: tokens, Type: "seller"})
				log.Printf("INFO: found %d token(s) for seller %s", len(tokens), sellerID)
			} else {
				log.Printf("WARN: no active FCM tokens for sellerID=%s", sellerID)
			}
		}
	}

	return s.deduplicateRecipients(recipients), nil
}

// resolveEnquiryRecipients returns who should be notified for an enquiry event
// based on the document's current status field.
//
// This is the single source of truth for routing inside the Cloud Function
// pipeline.  The same rules are expressed declaratively in watcher_rules.go for
// the server-side Firestore watcher.
func resolveEnquiryRecipients(docFields map[string]interface{}) (notifyUser, notifySeller bool) {
	status := strings.ToLower(strings.TrimSpace(firstNonEmptyString(docFields, "status")))
	acceptedBy := strings.ToLower(strings.TrimSpace(firstNonEmptyString(docFields, "acceptedBy", "accepted_by")))
	rejectedBy := strings.ToLower(strings.TrimSpace(firstNonEmptyString(docFields, "rejectedBy", "rejected_by")))

	switch status {
	// ── Seller must act ───────────────────────────────────────────────────────
	case "pending_seller_price", "pending_seller_final", "seller_final_update":
		return false, true

	// ── Buyer must act ────────────────────────────────────────────────────────
	case "pending_customer_price", "pending_customer_final", "customer_accepted_final":
		return true, false

	// ── Deal finalised — notify the OTHER party ───────────────────────────────
	case "completed_accepted", "completed_rejected":
		actor := acceptedBy
		if status == "completed_rejected" && rejectedBy != "" {
			actor = rejectedBy
		}
		switch actor {
		case "seller":
			return true, false // seller acted → notify buyer
		case "client", "customer", "user", "buyer":
			return false, true // buyer acted → notify seller
		default:
			return true, true // actor unknown → notify both
		}

	// ── Other active / admin states — notify both ─────────────────────────────
	case "in_progress", "on_hold", "resolved", "closed", "cancelled",
		"expired", "reopened", "counter_offer",
		"dispute", "dispute_resolved":
		return true, true

	default:
		// No status or unrecognised → new enquiry arriving: notify seller.
		return false, true
	}
}

// GetUserFCMTokens fetches active tokens for a user
func (s *Service) GetUserFCMTokens(ctx context.Context, userID string) ([]string, error) {
	return s.GetOwnerFCMTokens(ctx, "users", userID)
}

// SendToTokens sends a notification directly to the provided FCM device tokens.
func (s *Service) SendToTokens(ctx context.Context, tokens []string, title, body string, data map[string]string) error {
	if len(tokens) == 0 {
		return nil
	}
	if data == nil {
		data = map[string]string{}
	}
	// Filter empty tokens.
	valid := make([]string, 0, len(tokens))
	for _, t := range tokens {
		if t != "" {
			valid = append(valid, t)
		}
	}
	if len(valid) == 0 {
		return nil
	}
	msg := s.buildMulticastMessage(valid, &domain.NotificationPayload{Title: title, Body: body})
	if len(data) > 0 {
		msg.Data = data
	}
	resp, err := s.msgClient.SendEachForMulticast(ctx, msg)
	if err != nil {
		log.Printf("WARN [SendToTokens]: multicast failed: %v", err)
		return err
	}
	for i, r := range resp.Responses {
		if !r.Success {
			log.Printf("WARN [SendToTokens]: token %s failed: %v", valid[i], r.Error)
		}
	}
	log.Printf("INFO [SendToTokens]: sent %d/%d", resp.SuccessCount, len(valid))
	return nil
}

// GetOwnerFCMTokens fetches active tokens for a Firestore owner path:
// {collection}/{ownerID}/fcmTokens where collection is typically users/sellers.
func (s *Service) GetOwnerFCMTokens(ctx context.Context, collection, ownerID string) ([]string, error) {
	if s.fsClient == nil {
		return []string{}, nil
	}
	if ownerID == "" {
		return []string{}, nil
	}
	if collection == "" {
		collection = "users"
	}

	coll := s.fsClient.Collection(collection).Doc(ownerID).Collection("fcmTokens")
	docs, err := coll.Where("isActive", "==", true).Documents(ctx).GetAll()
	if err != nil {
		return nil, fmt.Errorf("fetch fcmTokens %s/%s: %w", collection, ownerID, err)
	}
	tokens := make([]string, 0, len(docs))
	for _, doc := range docs {
		data := doc.Data()
		if t, ok := data["token"].(string); ok && t != "" {
			// Only include tokens explicitly marked isActive: true.
			// The Firestore query already filters, but double-check the field.
			if a, ok := data["isActive"].(bool); ok && a {
				tokens = append(tokens, t)
			}
		}
	}
	return tokens, nil
}

func firstNonEmptyString(fields map[string]interface{}, keys ...string) string {
	for _, k := range keys {
		v, ok := fields[k]
		if !ok || v == nil {
			continue
		}
		s := strings.TrimSpace(fmt.Sprint(v))
		if s != "" && s != "<nil>" {
			return s
		}
	}
	return ""
}

// ─── FCM delivery helpers ─────────────────────────────────────────────────────

// fcmMaxTokensPerBatch is the documented FCM SendEachForMulticast limit.
const fcmMaxTokensPerBatch = 500

// chunkStrings splits a slice into sub-slices of at most size elements.
func chunkStrings(s []string, size int) [][]string {
	if size <= 0 {
		size = fcmMaxTokensPerBatch
	}
	var chunks [][]string
	for len(s) > size {
		chunks = append(chunks, s[:size])
		s = s[size:]
	}
	if len(s) > 0 {
		chunks = append(chunks, s)
	}
	return chunks
}

// isUnregisteredTokenError returns true when the FCM error indicates the device
// token is no longer valid (unregistered, app uninstalled, wrong project, etc.).
func isUnregisteredTokenError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	return strings.Contains(msg, "unregistered") ||
		strings.Contains(msg, "requested entity was not found") ||
		strings.Contains(msg, "registration-token-not-registered") ||
		strings.Contains(msg, "invalid registration") ||
		strings.Contains(msg, "notregistered")
}

// deactivateOwnerToken marks a single FCM token as inactive in Firestore.
// ownerCollection should be "users" or "sellers".
func (s *Service) deactivateOwnerToken(ctx context.Context, ownerCollection, ownerID, token string) {
	if s.fsClient == nil || ownerCollection == "" || ownerID == "" || token == "" {
		return
	}
	docRef := s.fsClient.Collection(ownerCollection).Doc(ownerID).Collection("fcmTokens").Doc(token)
	if _, err := docRef.Update(ctx, []firestore.Update{
		{Path: "isActive", Value: false},
		{Path: "updatedAt", Value: firestore.ServerTimestamp},
	}); err != nil {
		log.Printf("WARN: deactivateOwnerToken %s/%s token=%s: %v", ownerCollection, ownerID, token, err)
	} else {
		log.Printf("INFO: deactivatedOwnerToken %s/%s token=%s", ownerCollection, ownerID, token)
	}
}

func joinChangedFieldNames(changes []domain.FieldChange) string {
	if len(changes) == 0 {
		return ""
	}
	names := make([]string, 0, len(changes))
	for _, change := range changes {
		if change.FieldName != "" {
			names = append(names, change.FieldName)
		}
	}
	return strings.Join(names, ",")
}

func buildRecipientDedupeKey(payload *domain.NotificationPayload, recipient *domain.NotificationRecipient) string {
	fields := append([]string{}, payload.ChangedFields...)
	sort.Strings(fields)

	seed := strings.Join([]string{
		payload.DocumentID,
		payload.Timestamp,
		recipient.Type,
		recipient.UserID,
		strings.Join(fields, ","),
	}, "|")

	h := sha256.Sum256([]byte(seed))
	return "dedupe_" + hex.EncodeToString(h[:])
}

func (s *Service) acquireIdempotencyLease(
	ctx context.Context,
	key string,
	payload *domain.NotificationPayload,
	recipient *domain.NotificationRecipient,
) (bool, error) {
	if s.fsClient == nil {
		return true, nil
	}

	doc := s.fsClient.Collection(s.config.NotificationHistoryCollection).Doc(key)
	entry := map[string]interface{}{
		"dedupeKey":      key,
		"documentId":     payload.DocumentID,
		"documentPath":   payload.DocumentPath,
		"recipientId":    recipient.UserID,
		"recipientType":  recipient.Type,
		"eventTimestamp": payload.Timestamp,
		"status":         "pending",
		"createdAt":      firestore.ServerTimestamp,
		"expireAt":       time.Now().Add(24 * time.Hour),
	}

	_, err := doc.Create(ctx, entry)
	if err == nil {
		return true, nil
	}
	if grpcstatus.Code(err) == codes.AlreadyExists {
		return false, nil
	}
	return false, err
}

func (s *Service) markIdempotencyLeaseSent(ctx context.Context, key, messageID string, sentCount int) error {
	if s.fsClient == nil {
		return nil
	}
	_, err := s.fsClient.Collection(s.config.NotificationHistoryCollection).Doc(key).Set(ctx, map[string]interface{}{
		"status":    "sent",
		"messageId": messageID,
		"sentCount": sentCount,
		"sentAt":    firestore.ServerTimestamp,
	}, firestore.MergeAll)
	return err
}

func (s *Service) releaseIdempotencyLease(ctx context.Context, key string) error {
	if s.fsClient == nil {
		return nil
	}
	_, err := s.fsClient.Collection(s.config.NotificationHistoryCollection).Doc(key).Delete(ctx)
	return err
}

// buildMulticastMessage constructs a MulticastMessage for sending to multiple tokens at once.
func (s *Service) buildMulticastMessage(tokens []string, payload *domain.NotificationPayload) *messaging.MulticastMessage {
	data := map[string]string{
		"enquiryId":  payload.EnquiryID,
		"documentId": payload.DocumentID,
		"timestamp":  payload.Timestamp,
		"actionUrl":  payload.ActionURL,
	}
	return &messaging.MulticastMessage{
		Tokens: tokens,
		Data:   data,
		Notification: &messaging.Notification{
			Title: payload.Title,
			Body:  payload.Body,
		},
		Android: &messaging.AndroidConfig{
			Priority: "high",
			TTL:      ptrDuration(24 * time.Hour),
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: payload.Title,
						Body:  payload.Body,
					},
					Sound: "default",
				},
			},
		},
	}
}

// buildMessage constructs Firebase message
func (s *Service) buildMessage(payload *domain.NotificationPayload) *messaging.Message {
	data := map[string]string{
		"enquiryId":  payload.EnquiryID,
		"documentId": payload.DocumentID,
		"timestamp":  payload.Timestamp,
		"actionUrl":  payload.ActionURL,
	}

	msg := &messaging.Message{
		Data: data,
		Notification: &messaging.Notification{
			Title: payload.Title,
			Body:  payload.Body,
		},
		Android: &messaging.AndroidConfig{
			Priority: "high",
			TTL:      ptrDuration(24 * time.Hour),
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: payload.Title,
						Body:  payload.Body,
					},
					Sound: "default",
				},
			},
		},
	}
	return msg
}

// recordNotification stores idempotency info
func (s *Service) recordNotification(ctx context.Context, userID, documentID, messageID string) error {
	if s.fsClient == nil {
		return fmt.Errorf("Firestore client not available")
	}

	entry := map[string]interface{}{
		"userId":     userID,
		"documentId": documentID,
		"messageId":  messageID,
		"sentAt":     firestore.ServerTimestamp,
		"expireAt":   time.Now().Add(24 * time.Hour),
	}

	_, _, err := s.fsClient.Collection(s.config.NotificationHistoryCollection).Add(ctx, entry)
	return err
}

// deduplicateRecipients removes duplicate tokens
func (s *Service) deduplicateRecipients(recipients []*domain.NotificationRecipient) []*domain.NotificationRecipient {
	seen := map[string]bool{}
	result := []*domain.NotificationRecipient{}

	for _, r := range recipients {
		unique := []string{}
		for _, t := range r.Tokens {
			if !seen[t] {
				unique = append(unique, t)
				seen[t] = true
			}
		}
		if len(unique) > 0 {
			r.Tokens = unique
			result = append(result, r)
		}
	}

	return result
}

// Close is a no-op on the singleton service — closing the shared Firestore
// client would break subsequent CF warm-start invocations. The process-level
// clients are kept alive for the lifetime of the container.
func (s *Service) Close() error {
	return nil
}

// helper for Android TTL
func ptrDuration(d time.Duration) *time.Duration { return &d }
