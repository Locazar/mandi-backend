package usecase

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/rohit221990/mandi-backend/pkg/api/handler/request"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	notificationSvc "github.com/rohit221990/mandi-backend/pkg/service/notification"
)

// stubNotifRepo embeds the interface so unimplemented methods aren't needed;
// only the two GetActiveTokensByOwner/DeleteDeviceToken paths are exercised.
type stubNotifRepo struct {
	interfaces.NotificationRepository
	tokens  []string
	getErr  error
	deleted []string
}

func (s *stubNotifRepo) GetActiveTokensByOwner(_ context.Context, _, _ string) ([]string, error) {
	return s.tokens, s.getErr
}

func (s *stubNotifRepo) DeleteDeviceToken(_ context.Context, _, _, token string) error {
	s.deleted = append(s.deleted, token)
	return nil
}

// stubPush embeds PushSender; only the two send methods are overridden.
type stubPush struct {
	notificationSvc.PushSender
	sendTokensErr    error
	firestoreErr     error
	sendTokensCalled bool
	firestoreCalled  bool
}

func (s *stubPush) SendToTokens(_ context.Context, _ []string, _, _ string, _ map[string]string) error {
	s.sendTokensCalled = true
	return s.sendTokensErr
}

func (s *stubPush) SendToOwnerViaFirestore(_ context.Context, _, _, _, _ string, _ map[string]string) error {
	s.firestoreCalled = true
	return s.firestoreErr
}

func req() request.SendPushRequest {
	return request.SendPushRequest{OwnerID: "adm_x", OwnerType: "seller", Title: "t", Body: "b"}
}

// A wrapped ErrAllTokensUnreachable, as the real service returns.
func allUnreachable(n int) error {
	return fmt.Errorf("all %d token(s) unregistered: %w", n, notificationSvc.ErrAllTokensUnreachable)
}

// A wrapped ErrNoActiveTokens, as the real service returns.
func noActive() error {
	return fmt.Errorf("no active FCM tokens for sellers/adm_x: %w", notificationSvc.ErrNoActiveTokens)
}

func TestSendPushNotification_LoggedOutSeller_IsSuccessNoOpAndPrunes(t *testing.T) {
	// Reported scenario: 2 stale Postgres tokens (all NotRegistered) + no active
	// Firestore tokens. Admin's send must succeed and the stale tokens pruned.
	repo := &stubNotifRepo{tokens: []string{"tokA", "tokB"}}
	push := &stubPush{sendTokensErr: allUnreachable(2), firestoreErr: noActive()}
	uc := &notificationUseCase{notificationRepo: repo, fcmPush: push}

	if err := uc.SendPushNotification(context.Background(), req()); err != nil {
		t.Fatalf("expected success (no-op) for logged-out seller, got error: %v", err)
	}
	if len(repo.deleted) != 2 {
		t.Errorf("expected both stale tokens pruned, deleted = %v", repo.deleted)
	}
	if !push.firestoreCalled {
		t.Error("expected Firestore fallback to be attempted")
	}
}

func TestSendPushNotification_NoDevicesAnywhere_IsSuccess(t *testing.T) {
	// No Postgres tokens at all + no active Firestore tokens → successful no-op.
	repo := &stubNotifRepo{tokens: nil}
	push := &stubPush{firestoreErr: noActive()}
	uc := &notificationUseCase{notificationRepo: repo, fcmPush: push}

	if err := uc.SendPushNotification(context.Background(), req()); err != nil {
		t.Fatalf("expected success for owner with no devices, got: %v", err)
	}
	if push.sendTokensCalled {
		t.Error("SendToTokens should not be called when there are no Postgres tokens")
	}
}

func TestSendPushNotification_GenuineTransportError_IsSurfaced(t *testing.T) {
	// A real FCM transport error (not a sentinel) must still fail the request.
	repo := &stubNotifRepo{tokens: nil}
	push := &stubPush{firestoreErr: errors.New("FCM multicast send: context deadline exceeded")}
	uc := &notificationUseCase{notificationRepo: repo, fcmPush: push}

	if err := uc.SendPushNotification(context.Background(), req()); err == nil {
		t.Fatal("expected a genuine transport error to be surfaced, got nil")
	}
}

func TestSendPushNotification_PostgresDeliverySucceeds_SkipsFirestore(t *testing.T) {
	repo := &stubNotifRepo{tokens: []string{"tokA"}}
	push := &stubPush{sendTokensErr: nil}
	uc := &notificationUseCase{notificationRepo: repo, fcmPush: push}

	if err := uc.SendPushNotification(context.Background(), req()); err != nil {
		t.Fatalf("expected success, got: %v", err)
	}
	if push.firestoreCalled {
		t.Error("Firestore fallback should not run when Postgres delivery succeeds")
	}
	if len(repo.deleted) != 0 {
		t.Error("no tokens should be pruned on a successful send")
	}
}
