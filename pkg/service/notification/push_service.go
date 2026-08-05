package notification

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"firebase.google.com/go/v4/messaging"
)

// Sentinel errors that distinguish "there is simply no reachable device right
// now" from a genuine delivery failure. Callers can use errors.Is to treat the
// former as a successful no-op — e.g. an admin sending to a seller who is logged
// out, whose device tokens are all gone/unregistered, should not be an error.
var (
	// ErrNoActiveTokens means the owner has no registered device tokens.
	ErrNoActiveTokens = errors.New("no active FCM tokens")
	// ErrAllTokensUnreachable means every target token is permanently invalid
	// (unregistered / app uninstalled / logged out), so nothing was delivered.
	ErrAllTokensUnreachable = errors.New("all FCM tokens unreachable")
)

// PushSender is the interface for sending FCM push notifications.
// Implementations are injected into the notification usecase.
type PushSender interface {
	// SendToTokens sends a notification to one or more device tokens directly.
	SendToTokens(ctx context.Context, tokens []string, title, body string, data map[string]string) error

	// SendToOwnerViaFirestore looks up tokens from Firestore and sends.
	// ownerCollection is "users" or "sellers".
	SendToOwnerViaFirestore(ctx context.Context, ownerCollection, ownerID, title, body string, data map[string]string) error

	// SendToTopic broadcasts a notification to every device subscribed to a
	// topic (e.g. "all_users"). Device-level, so it reaches logged-out and
	// anonymous devices too, and scales to any audience with a single call.
	SendToTopic(ctx context.Context, topic, title, body string, data map[string]string) error

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
	imageURL := strings.TrimSpace(data["image_url"])
	if imageURL == "" {
		imageURL = strings.TrimSpace(data["product_image_url"])
	}

	msg := &messaging.MulticastMessage{
		Tokens: tokens,
		Notification: &messaging.Notification{
			Title:    title,
			Body:     body,
			ImageURL: imageURL,
		},
		Data: data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			TTL:      ptrDuration(24 * time.Hour),
			Notification: &messaging.AndroidNotification{
				ChannelID:    "high_importance_channel",
				Priority:     messaging.PriorityHigh,
				Title:        title,
				Body:         body,
				ClickAction:  "FLUTTER_NOTIFICATION_CLICK",
				ImageURL:     imageURL,
				Sound:        "default",
				DefaultSound: true,
			},
		},
		APNS: &messaging.APNSConfig{
			FCMOptions: &messaging.APNSFCMOptions{ImageURL: imageURL},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert: &messaging.ApsAlert{
						Title: title,
						Body:  body,
					},
					MutableContent: true,
					Sound:          "default",
				},
			},
		},
		Webpush: &messaging.WebpushConfig{
			Notification: &messaging.WebpushNotification{
				Title: title,
				Body:  body,
				Image: imageURL,
			},
		},
	}

	resp, err := s.msgClient.SendEachForMulticast(ctx, msg)
	if err != nil {
		return fmt.Errorf("FCM multicast send: %w", err)
	}

	invalidCount := 0
	if resp.FailureCount > 0 {
		for i, r := range resp.Responses {
			if !r.Success {
				log.Printf("WARN: FCM send failed for token %s: %v", tokens[i], r.Error)
				if isInvalidTokenError(r.Error) {
					invalidCount++
				}
			}
		}
	}

	log.Printf("INFO: FCM sent %d/%d successfully", resp.SuccessCount, len(tokens))
	if resp.SuccessCount == 0 {
		// Every token permanently invalid (owner logged out / uninstalled): signal
		// with a sentinel so the caller can prune them and treat it as "no
		// reachable device" instead of a hard failure.
		if invalidCount == len(tokens) {
			return fmt.Errorf("all %d token(s) unregistered: %w", len(tokens), ErrAllTokensUnreachable)
		}
		return fmt.Errorf("FCM multicast send: all %d token(s) failed, e.g. %v", len(tokens), resp.Responses[0].Error)
	}
	return nil
}

// isInvalidTokenError returns true when the FCM error indicates the token is
// no longer valid (unregistered / app uninstalled / wrong project).
func isInvalidTokenError(err error) bool {
	if err == nil {
		return false
	}
	// Typed SDK check first — the reliable way to detect a token FCM has dropped
	// (app uninstalled / logged out / token rotated). Covers the common
	// "NotRegistered" response that the string match below previously missed.
	if messaging.IsUnregistered(err) {
		return true
	}
	// Case-insensitive string fallback for wrapped errors where the typed value
	// was lost. Note "NotRegistered" does NOT contain "unregistered", so it must
	// be matched explicitly.
	s := strings.ToLower(err.Error())
	return strings.Contains(s, "unregistered") ||
		strings.Contains(s, "notregistered") ||
		strings.Contains(s, "not registered") ||
		strings.Contains(s, "not-registered") ||
		strings.Contains(s, "registration-token-not-registered") ||
		strings.Contains(s, "requested entity was not found")
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
		return fmt.Errorf("no active FCM tokens for %s/%s: %w", ownerCollection, ownerID, ErrNoActiveTokens)
	}

	imageURL := strings.TrimSpace(data["image_url"])
	if imageURL == "" {
		imageURL = strings.TrimSpace(data["product_image_url"])
	}

	// Send and collect invalid tokens for cleanup.
	msg := &messaging.MulticastMessage{
		Tokens:       tokens,
		Notification: &messaging.Notification{Title: title, Body: body, ImageURL: imageURL},
		Data:         data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			TTL:      ptrDuration(24 * time.Hour),
			Notification: &messaging.AndroidNotification{
				Title:       title,
				Body:        body,
				ClickAction: "FLUTTER_NOTIFICATION_CLICK",
				ImageURL:    imageURL,
				Sound:       "default",
			},
		},
		APNS: &messaging.APNSConfig{
			FCMOptions: &messaging.APNSFCMOptions{ImageURL: imageURL},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert:          &messaging.ApsAlert{Title: title, Body: body},
					MutableContent: true,
					Sound:          "default",
				},
			},
		},
		Webpush: &messaging.WebpushConfig{
			Notification: &messaging.WebpushNotification{
				Title: title,
				Body:  body,
				Image: imageURL,
			},
		},
	}

	resp, err := s.msgClient.SendEachForMulticast(ctx, msg)
	if err != nil {
		return fmt.Errorf("FCM multicast send: %w", err)
	}

	invalidCount := 0
	for i, r := range resp.Responses {
		if r.Success {
			continue
		}
		log.Printf("WARN: FCM send failed for token %s: %v", tokens[i], r.Error)
		if isInvalidTokenError(r.Error) {
			invalidCount++
			log.Printf("INFO: deactivating invalid FCM token for %s/%s", ownerCollection, ownerID)
			_ = s.deactivateToken(ctx, ownerCollection, ownerID, tokens[i])
		}
	}

	log.Printf("INFO: FCM sent %d/%d successfully to %s/%s", resp.SuccessCount, len(tokens), ownerCollection, ownerID)
	if resp.SuccessCount == 0 {
		// All invalid → the owner has no reachable device; the invalid tokens
		// were just deactivated above. Signal with a sentinel, not a hard error.
		if invalidCount == len(tokens) {
			return fmt.Errorf("all %d token(s) unregistered for %s/%s: %w", len(tokens), ownerCollection, ownerID, ErrAllTokensUnreachable)
		}
		return fmt.Errorf("FCM multicast send: all %d token(s) failed for %s/%s, e.g. %v", len(tokens), ownerCollection, ownerID, resp.Responses[0].Error)
	}
	return nil
}

// SendToTopic broadcasts a notification to every device subscribed to [topic].
// One FCM call reaches the whole audience regardless of login state.
func (s *FCMPushService) SendToTopic(
	ctx context.Context,
	topic, title, body string,
	data map[string]string,
) error {
	if err := s.ensureInit(ctx); err != nil {
		return fmt.Errorf("FCM init failed: %w", err)
	}
	if strings.TrimSpace(topic) == "" {
		return fmt.Errorf("topic is required")
	}
	if data == nil {
		data = map[string]string{}
	}
	data["timestamp"] = time.Now().UTC().Format(time.RFC3339)
	imageURL := strings.TrimSpace(data["image_url"])
	if imageURL == "" {
		imageURL = strings.TrimSpace(data["product_image_url"])
	}

	msg := &messaging.Message{
		Topic:        topic,
		Notification: &messaging.Notification{Title: title, Body: body, ImageURL: imageURL},
		Data:         data,
		Android: &messaging.AndroidConfig{
			Priority: "high",
			TTL:      ptrDuration(24 * time.Hour),
			Notification: &messaging.AndroidNotification{
				ChannelID:    "high_importance_channel",
				Priority:     messaging.PriorityHigh,
				Title:        title,
				Body:         body,
				ClickAction:  "FLUTTER_NOTIFICATION_CLICK",
				ImageURL:     imageURL,
				Sound:        "default",
				DefaultSound: true,
			},
		},
		APNS: &messaging.APNSConfig{
			FCMOptions: &messaging.APNSFCMOptions{ImageURL: imageURL},
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					Alert:          &messaging.ApsAlert{Title: title, Body: body},
					MutableContent: true,
					Sound:          "default",
				},
			},
		},
		Webpush: &messaging.WebpushConfig{
			Notification: &messaging.WebpushNotification{Title: title, Body: body, Image: imageURL},
		},
	}

	id, err := s.msgClient.Send(ctx, msg)
	if err != nil {
		return fmt.Errorf("FCM topic send to %q: %w", topic, err)
	}
	log.Printf("INFO: FCM broadcast sent to topic %q (message id %s)", topic, id)
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

// SaveTokenToFirestore persists a device FCM token in Firestore and deactivates
// all previous tokens for the same owner so that only one token is ever active.
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

	coll := s.fsClient.Collection(ownerCollection).Doc(ownerID).Collection("fcmTokens")

	// Deactivate all existing tokens that are NOT the current token.
	// This prevents duplicate notifications when the device refreshes its FCM token.
	existingDocs, err := coll.Where("isActive", "==", true).Documents(ctx).GetAll()
	if err != nil {
		log.Printf("WARN [SaveTokenToFirestore]: could not query existing tokens for %s/%s: %v", ownerCollection, ownerID, err)
	} else {
		for _, doc := range existingDocs {
			if doc.Ref.ID == token {
				continue // this is the token we're about to save — skip
			}
			if _, updateErr := doc.Ref.Update(ctx, []firestore.Update{
				{Path: "isActive", Value: false},
				{Path: "updatedAt", Value: firestore.ServerTimestamp},
			}); updateErr != nil {
				log.Printf("WARN [SaveTokenToFirestore]: failed to deactivate old token %s for %s/%s: %v", doc.Ref.ID, ownerCollection, ownerID, updateErr)
			}
		}
	}

	// Save (or overwrite) the new token as active.
	_, err = coll.Doc(token).Set(ctx, map[string]interface{}{
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
