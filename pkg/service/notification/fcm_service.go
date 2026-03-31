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

// NewService creates a new notification service
func NewService(ctx context.Context, config Config) (*Service, error) {
	app, err := firebase.NewApp(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to init Firebase app: %w", err)
	}

	msgClient, err := app.Messaging(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get messaging client: %w", err)
	}

	fsClient, err := app.Firestore(ctx)
	if err != nil {
		log.Printf("WARN: Firestore client not available: %v", err)
	}

	var rtdbClient *db.Client
	dbURL := os.Getenv("FIREBASE_DB_URL")
	if dbURL != "" {
		rtdbClient, err = app.Database(ctx)
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
		app:        app,
		config:     config,
		msgClient:  msgClient,
		fsClient:   fsClient,
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
		log.Printf("INFO: No recipients found for enquiry %s", payload.EnquiryID)
		return nil
	}

	message := s.buildMessage(payload)
	successCount := 0

	for _, recipient := range recipients {
		dedupeKey := ""
		leaseAcquired := true
		if s.config.EnableIdempotencyCheck {
			dedupeKey = buildRecipientDedupeKey(payload, recipient)
			leaseAcquired, err = s.acquireIdempotencyLease(ctx, dedupeKey, payload, recipient)
			if err != nil {
				log.Printf("WARN: idempotency lease error recipient=%s key=%s: %v", recipient.UserID, dedupeKey, err)
				// Fail-open: deliver notification even if idempotency store is temporarily unavailable.
				leaseAcquired = true
			}
			if !leaseAcquired {
				log.Printf("INFO: duplicate notification suppressed recipient=%s key=%s document=%s", recipient.UserID, dedupeKey, payload.DocumentID)
				continue
			}
		}

		recipientSuccessCount := 0
		firstMessageID := ""
		for _, token := range recipient.Tokens {
			if token == "" {
				continue
			}

			message.Token = token
			resp, err := s.msgClient.Send(ctx, message)
			if err != nil {
				log.Printf("WARN: Failed to send to token %s: %v", token, err)
				continue
			}

			successCount++
			recipientSuccessCount++
			if firstMessageID == "" {
				firstMessageID = resp
			}
		}

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
		return fmt.Errorf("failed to send notification to any recipient")
	}
	log.Printf("INFO: Sent %d notifications", successCount)
	return nil
}

// GetNotificationRecipients fetches all tokens for a payload
func (s *Service) GetNotificationRecipients(ctx context.Context, event *domain.ParsedFirestoreEvent,
	payload *domain.NotificationPayload, changes []domain.FieldChange) ([]*domain.NotificationRecipient, error) {

	recipients := []*domain.NotificationRecipient{}

	newFields := map[string]interface{}{}
	if event != nil && event.NewFields != nil {
		newFields = event.NewFields
	}

	userID := firstNonEmptyString(newFields, "userId", "customerId", "user_id", "createdBy")
	if userID == "" {
		userID = payload.UserID
	}

	sellerID := firstNonEmptyString(newFields, "sellerId", "shopId", "shop_id", "seller_id", "assignedTo", "assignedToId")
	if sellerID == "" {
		sellerID = payload.AssignedTo
	}

	target := resolveEnquiryRecipientTarget(newFields, changes, userID, sellerID)
	log.Printf(
		"INFO: enquiry recipient routing document=%s status=%q acceptedBy=%q rejectedBy=%q target=%q userID=%q sellerID=%q changedFields=%s",
		payload.DocumentID,
		firstNonEmptyString(newFields, "status"),
		firstNonEmptyString(newFields, "acceptedBy", "accepted_by"),
		firstNonEmptyString(newFields, "rejectedBy", "rejected_by"),
		target,
		userID,
		sellerID,
		joinChangedFieldNames(changes),
	)
	switch target {
	case "user":
		if userID != "" {
			tokens, _ := s.GetOwnerFCMTokens(ctx, "users", userID)
			if len(tokens) > 0 {
				recipients = append(recipients, &domain.NotificationRecipient{UserID: userID, Tokens: tokens, Type: "user"})
			}
		}
	case "seller":
		if sellerID != "" {
			tokens, _ := s.GetOwnerFCMTokens(ctx, "sellers", sellerID)
			if len(tokens) > 0 {
				recipients = append(recipients, &domain.NotificationRecipient{UserID: sellerID, Tokens: tokens, Type: "seller"})
			}
		}
	default:
		// Unknown routing state: don't fan out to both sides.
		log.Printf("INFO: recipient target unresolved for enquiry document=%s status=%v", payload.DocumentID, newFields["status"])
	}

	return s.deduplicateRecipients(recipients), nil
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
	for _, token := range tokens {
		if token == "" {
			continue
		}
		msg := s.buildMessage(&domain.NotificationPayload{Title: title, Body: body})
		msg.Token = token
		if len(data) > 0 {
			msg.Data = data
		}
		if _, err := s.msgClient.Send(ctx, msg); err != nil {
			log.Printf("WARN [SendToTokens]: failed to send to token %s: %v", token, err)
		}
	}
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
	docs, _ := coll.Documents(ctx).GetAll()
	tokens := []string{}
	for _, doc := range docs {
		data := doc.Data()
		if t, ok := data["token"].(string); ok && t != "" {
			active := true
			if a, ok := data["isActive"].(bool); ok {
				active = a
			}
			if active {
				tokens = append(tokens, t)
			}
		}
	}
	return tokens, nil
}

// resolveEnquiryRecipientTarget enforces one-way notification routing for
// enquiry negotiation updates.
//
// Returns:
//   - "user"   => notify only buyer/user
//   - "seller" => notify only seller
//   - ""       => skip when target can't be determined safely
func resolveEnquiryRecipientTarget(
	newFields map[string]interface{},
	changes []domain.FieldChange,
	userID, sellerID string,
) string {
	status := strings.ToLower(strings.TrimSpace(firstNonEmptyString(newFields, "status")))
	acceptedBy := strings.ToLower(strings.TrimSpace(firstNonEmptyString(newFields, "acceptedBy", "accepted_by")))
	rejectedBy := strings.ToLower(strings.TrimSpace(firstNonEmptyString(newFields, "rejectedBy", "rejected_by")))

	// Highest priority: if actor identity is present, always notify the opposite party.
	actorID := firstNonEmptyString(newFields, "updatedBy", "updatedById", "lastUpdatedBy", "modifiedBy", "changedBy", "actorId")
	if actorID != "" {
		if userID != "" && actorID == userID {
			return "seller"
		}
		if sellerID != "" && actorID == sellerID {
			return "user"
		}
	}

	actorRole := strings.ToLower(strings.TrimSpace(firstNonEmptyString(newFields, "updatedByRole", "updatedByType", "actorRole", "actorType")))
	switch actorRole {
	case "user", "customer", "client", "buyer":
		return "seller"
	case "seller", "vendor", "shop":
		return "user"
	}

	// Explicit completion routing from acceptedBy.
	if status == "completed_accepted" || status == "completed_rejected" {
		actor := acceptedBy
		if status == "completed_rejected" && rejectedBy != "" {
			actor = rejectedBy
		}
		switch actor {
		case "seller", "vendor", "shop":
			return "user"
		case "client", "customer", "user", "buyer":
			return "seller"
		}
	}

	// Status-based routing for negotiation states. These statuses represent the
	// party expected to take the next action, so notify that side only.
	switch status {
	case "pending_seller_price", "pending_seller_final", "seller_final_update":
		return "seller"
	case "pending_customer_price", "pending_customer_final":
		return "user"
	}

	// Fallback: new/initial enquiry (status empty, "pending", "new", or any
	// unrecognised value) — always notify the seller so they know about
	// incoming enquiries that don't yet have a negotiation-stage status.
	return "seller"
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

// Close only closes Firestore
func (s *Service) Close() error {
	if s.fsClient != nil {
		return s.fsClient.Close()
	}
	return nil
}

// helper for Android TTL
func ptrDuration(d time.Duration) *time.Duration { return &d }
