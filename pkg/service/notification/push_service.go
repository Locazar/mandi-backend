package notification

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/v4/messaging"
)

// PushSender is the interface for sending FCM push notifications.
// Implementations are injected into the notification usecase.
type PushSender interface {
	// SendToTokens sends a notification to one or more device tokens directly.
	SendToTokens(ctx context.Context, tokens []string, title, body string, data map[string]string) error

	// SendToOwnerViaFirestore looks up tokens from Firestore and sends.
	// ownerCollection is "users" or "sellers".
	SendToOwnerViaFirestore(ctx context.Context, ownerCollection, ownerID, title, body string, data map[string]string) error

	// SaveTokenToFirestore persists a device token in Firestore so Cloud Functions can read it.
	// ownerCollection is "users" or "sellers".
	SaveTokenToFirestore(ctx context.Context, ownerCollection, ownerID, token, platform string) error

	// DeleteTokenFromFirestore removes an FCM token (e.g. on logout / token refresh).
	DeleteTokenFromFirestore(ctx context.Context, ownerCollection, ownerID, token string) error
}

// FCMPushService implements PushSender using Firebase Admin SDK.
// Firebase is initialised lazily on first use; the zero-value struct is valid.
type FCMPushService struct {
	msgClient *messaging.Client
	fsClient  *firestore.Client
	once      sync.Once
	initErr   error
}

// NewFCMPushService returns an uninitialised FCMPushService.
// Firebase clients are started on the first method call.
func NewFCMPushService() *FCMPushService {
	return &FCMPushService{}
}

// ensureInit initialises the Firebase clients exactly once.
// Uses the package-level shared Firebase App so that this service and
// FirestoreWatcher never attempt to create a second default App.
func (s *FCMPushService) ensureInit(ctx context.Context) error {
	s.once.Do(func() {
		var err error
		s.msgClient, err = sharedMessagingClient(ctx)
		if err != nil {
			s.initErr = fmt.Errorf("FCM messaging client: %w", err)
			return
		}
		// Firestore is optional – used for token storage sync
		s.fsClient, _ = sharedFirestoreClient(ctx)
	})
	return s.initErr
}

// SendToTokens sends a notification directly to the given device tokens.
func (s *FCMPushService) SendToTokens(
	ctx context.Context,
	tokens []string,
	title, body string,
	data map[string]string,
) error {
	if err := s.ensureInit(ctx); err != nil {
		return fmt.Errorf("FCM init failed: %w", err)
	}

	if len(tokens) == 0 {
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
			TTL:      ptrDuration(24 * time.Hour),
			Notification: &messaging.AndroidNotification{
				ChannelID: "high_importance_channel",
				Priority: messaging.PriorityHigh,
				Title:       title,
				Body:        body,
				ClickAction: "FLUTTER_NOTIFICATION_CLICK",
				Sound:       "default",
				DefaultSound: true,
			},
		},
		APNS: &messaging.APNSConfig{
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: title,
						Body:  body,
					},
					Sound: "default",
				},
			},
		},
	}

	resp, err := s.msgClient.SendEachForMulticast(ctx, msg)
	if err != nil {
		return fmt.Errorf("FCM multicast send: %w", err)
	}

	if resp.FailureCount > 0 {
		for i, r := range resp.Responses {
			if !r.Success {
				log.Printf("WARN: FCM send failed for token %s: %v", tokens[i], r.Error)
			}
		}
	}

	log.Printf("INFO: FCM sent %d/%d successfully", resp.SuccessCount, len(tokens))
	return nil
}

// isInvalidTokenError returns true when the FCM error indicates the token is
// no longer valid (unregistered / app uninstalled / wrong project).
func isInvalidTokenError(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "UNREGISTERED") ||
		strings.Contains(s, "Requested entity was not found") ||
		strings.Contains(s, "registration-token-not-registered")
}

// SendToOwnerViaFirestore fetches active FCM tokens for the owner from Firestore,
// delivers the notification, and automatically deactivates any invalid tokens.
func (s *FCMPushService) SendToOwnerViaFirestore(
	ctx context.Context,
	ownerCollection, ownerID, title, body string,
	data map[string]string,
) error {
	if err := s.ensureInit(ctx); err != nil {
		return fmt.Errorf("FCM init failed: %w", err)
	}
	if s.fsClient == nil {
		return fmt.Errorf("Firestore client not available")
	}

	tokens, err := s.getTokensFromFirestore(ctx, ownerCollection, ownerID)
	if err != nil {
		return fmt.Errorf("fetch tokens: %w", err)
	}
	if len(tokens) == 0 {
		log.Printf("INFO: no active FCM tokens for %s/%s", ownerCollection, ownerID)
		return nil
	}

	// Send and collect invalid tokens for cleanup.
	msg := &messaging.MulticastMessage{
		Tokens:       tokens,
		Notification: &messaging.Notification{Title: title, Body: body},
		Data:         data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			TTL:      ptrDuration(24 * time.Hour),
			Notification: &messaging.AndroidNotification{
				Title:       title,
				Body:        body,
				ClickAction: "FLUTTER_NOTIFICATION_CLICK",
				Sound:       "default",
			},
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

	resp, err := s.msgClient.SendEachForMulticast(ctx, msg)
	if err != nil {
		return fmt.Errorf("FCM multicast send: %w", err)
	}

	for i, r := range resp.Responses {
		if r.Success {
			continue
		}
		log.Printf("WARN: FCM send failed for token %s: %v", tokens[i], r.Error)
		if isInvalidTokenError(r.Error) {
			log.Printf("INFO: deactivating invalid FCM token for %s/%s", ownerCollection, ownerID)
			_ = s.deactivateToken(ctx, ownerCollection, ownerID, tokens[i])
		}
	}

	log.Printf("INFO: FCM sent %d/%d successfully to %s/%s", resp.SuccessCount, len(tokens), ownerCollection, ownerID)
	return nil
}

// deactivateToken marks a Firestore token document as isActive=false.
func (s *FCMPushService) deactivateToken(ctx context.Context, ownerCollection, ownerID, token string) error {
	docRef := s.fsClient.
		Collection(ownerCollection).
		Doc(ownerID).
		Collection("fcmTokens").
		Doc(token)
	_, err := docRef.Update(ctx, []firestore.Update{
		{Path: "isActive", Value: false},
	})
	return err
}

// SaveTokenToFirestore persists a device FCM token in Firestore.
// Path: {ownerCollection}/{ownerID}/fcmTokens/{token}
func (s *FCMPushService) SaveTokenToFirestore(
	ctx context.Context,
	ownerCollection, ownerID, token, platform string,
) error {
	if err := s.ensureInit(ctx); err != nil {
		return fmt.Errorf("FCM init failed: %w", err)
	}
	if s.fsClient == nil {
		return fmt.Errorf("Firestore client not available")
	}

	docRef := s.fsClient.
		Collection(ownerCollection).
		Doc(ownerID).
		Collection("fcmTokens").
		Doc(token)

	_, err := docRef.Set(ctx, map[string]interface{}{
		"token":     token,
		"platform":  platform,
		"isActive":  true,
		"updatedAt": firestore.ServerTimestamp,
	})
	return err
}

// DeleteTokenFromFirestore marks an FCM token as inactive in Firestore.
// Path: {ownerCollection}/{ownerID}/fcmTokens/{token}
func (s *FCMPushService) DeleteTokenFromFirestore(
	ctx context.Context,
	ownerCollection, ownerID, token string,
) error {
	if err := s.ensureInit(ctx); err != nil {
		return fmt.Errorf("FCM init failed: %w", err)
	}
	if s.fsClient == nil {
		return fmt.Errorf("Firestore client not available")
	}

	docRef := s.fsClient.
		Collection(ownerCollection).
		Doc(ownerID).
		Collection("fcmTokens").
		Doc(token)

	_, err := docRef.Update(ctx, []firestore.Update{
		{Path: "isActive", Value: false},
		{Path: "updatedAt", Value: firestore.ServerTimestamp},
	})
	return err
}

// getTokensFromFirestore retrieves all active FCM tokens for an owner from Firestore.
func (s *FCMPushService) getTokensFromFirestore(
	ctx context.Context,
	ownerCollection, ownerID string,
) ([]string, error) {
	coll := s.fsClient.Collection(ownerCollection).Doc(ownerID).Collection("fcmTokens")
	docs, err := coll.Where("isActive", "==", true).Documents(ctx).GetAll()
	if err != nil {
		return nil, err
	}

	tokens := make([]string, 0, len(docs))
	for _, doc := range docs {
		data := doc.Data()
		if t, ok := data["token"].(string); ok && t != "" {
			tokens = append(tokens, t)
		}
	}
	return tokens, nil
}
