package notification

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/db"
	"firebase.google.com/go/v4/messaging"
	"github.com/rohit221990/mandi-backend/pkg/domain"
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
			if s.config.EnableIdempotencyCheck {
				if err := s.recordNotification(ctx, recipient.UserID, payload.DocumentID, resp); err != nil {
					log.Printf("WARN: Failed to record notification: %v", err)
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

	if payload.UserID != "" {
		tokens, _ := s.GetUserFCMTokens(ctx, payload.UserID)
		if len(tokens) > 0 {
			recipients = append(recipients, &domain.NotificationRecipient{
				UserID: payload.UserID, Tokens: tokens, Type: "user",
			})
		}
	}

	if payload.AssignedTo != "" && payload.AssignedTo != payload.UserID {
		tokens, _ := s.GetUserFCMTokens(ctx, payload.AssignedTo)
		if len(tokens) > 0 {
			recipients = append(recipients, &domain.NotificationRecipient{
				UserID: payload.AssignedTo, Tokens: tokens, Type: "admin",
			})
		}
	}

	return s.deduplicateRecipients(recipients), nil
}

// GetUserFCMTokens fetches active tokens for a user
func (s *Service) GetUserFCMTokens(ctx context.Context, userID string) ([]string, error) {
	if s.fsClient == nil {
		return []string{}, nil
	}

	coll := s.fsClient.Collection("enquiry").Doc(userID).Collection("fcmTokens")
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

// buildMessage constructs Firebase message
func (s *Service) buildMessage(payload *domain.NotificationPayload) *messaging.Message {
	data := map[string]string{
		"enquiryId": payload.EnquiryID,
		"documentId": payload.DocumentID,
		"timestamp": payload.Timestamp,
		"actionUrl": payload.ActionURL,
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