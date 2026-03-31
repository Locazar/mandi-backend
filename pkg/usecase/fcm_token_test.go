package usecase

import (
	"context"
	"errors"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/rohit221990/mandi-backend/pkg/domain"
)

// ---------------------------------------------------------------------------
// Minimal hand-written mocks (avoids a mockgen dependency for this file)
// ---------------------------------------------------------------------------

// mockFcmTokenRepo implements interfaces.FcmTokenRepository.
type mockFcmTokenRepo struct {
	saveFn          func(domain.FcmToken) (domain.FcmToken, error)
	upsertFn        func(token, ownerID, ownerType, platform string) error
	upsertCalled    bool
	upsertToken     string
	upsertOwnerID   string
	upsertOwnerType string
}

func (m *mockFcmTokenRepo) SaveFcmToken(f domain.FcmToken) (domain.FcmToken, error) {
	return m.saveFn(f)
}

func (m *mockFcmTokenRepo) UpsertDeviceToken(token, ownerID, ownerType, platform string) error {
	m.upsertCalled = true
	m.upsertToken = token
	m.upsertOwnerID = ownerID
	m.upsertOwnerType = ownerType
	if m.upsertFn != nil {
		return m.upsertFn(token, ownerID, ownerType, platform)
	}
	return nil
}

// mockPushSender implements notification.PushSender.
type mockPushSender struct {
	sendToTokensCalled         bool
	savedToFirestore           bool
	savedFirestoreCollection   string
	savedFirestoreOwnerID      string
	savedFirestoreToken        string
	saveToFirestoreErr         error
	deleteFromFirestoreErr     error
	sendToOwnerViaFirestoreErr error
}

func (m *mockPushSender) SendToTokens(_ context.Context, _ []string, _, _ string, _ map[string]string) error {
	m.sendToTokensCalled = true
	return nil
}

func (m *mockPushSender) SendToOwnerViaFirestore(_ context.Context, _, _, _, _ string, _ map[string]string) error {
	return m.sendToOwnerViaFirestoreErr
}

func (m *mockPushSender) SaveTokenToFirestore(_ context.Context, collection, ownerID, token, _ string) error {
	m.savedToFirestore = true
	m.savedFirestoreCollection = collection
	m.savedFirestoreOwnerID = ownerID
	m.savedFirestoreToken = token
	return m.saveToFirestoreErr
}

func (m *mockPushSender) DeleteTokenFromFirestore(_ context.Context, _, _, _ string) error {
	return m.deleteFromFirestoreErr
}

// ---------------------------------------------------------------------------
// Helper – build the use-case with injected mocks
// ---------------------------------------------------------------------------

func newTestFcmTokenUseCase(repo *mockFcmTokenRepo, push *mockPushSender) *fcmTokenUseCase {
	return &fcmTokenUseCase{repo: repo, fcmPush: push}
}

// ---------------------------------------------------------------------------
// Tests
// ---------------------------------------------------------------------------

// TestSaveFcmToken_ShopID_SyncsToSellers verifies that when a token is submitted
// with a ShopID the usecase:
//  1. calls SaveFcmToken on the repository
//  2. syncs the token to Firestore under the "sellers" collection
//  3. upserts the token into notification_device_tokens with owner_type="seller"
func TestSaveFcmToken_ShopID_SyncsToSellers(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()

	const (
		testToken    = "fcm-token-abc123"
		testShopID   = uint(42)
		testPlatform = "android"
	)

	repo := &mockFcmTokenRepo{
		saveFn: func(f domain.FcmToken) (domain.FcmToken, error) {
			// Return the same struct (simulates DB save)
			return f, nil
		},
	}
	push := &mockPushSender{}

	uc := newTestFcmTokenUseCase(repo, push)

	saved, err := uc.SaveFcmToken(domain.FcmToken{
		Token:    testToken,
		Platform: testPlatform,
		ShopID:   testShopID,
	})

	if err != nil {
		t.Fatalf("SaveFcmToken returned unexpected error: %v", err)
	}
	if saved.Token != testToken {
		t.Errorf("expected token %q, got %q", testToken, saved.Token)
	}

	// Firestore sync must target "sellers", not "admins" or anything else.
	if !push.savedToFirestore {
		t.Fatal("SaveTokenToFirestore was NOT called — token will never be found by the watcher")
	}
	if push.savedFirestoreCollection != "sellers" {
		t.Errorf("Firestore collection = %q, want %q", push.savedFirestoreCollection, "sellers")
	}
	if push.savedFirestoreOwnerID != "42" {
		t.Errorf("Firestore ownerID = %q, want \"42\"", push.savedFirestoreOwnerID)
	}
	if push.savedFirestoreToken != testToken {
		t.Errorf("Firestore token = %q, want %q", push.savedFirestoreToken, testToken)
	}

	// Postgres notification_device_tokens must also be populated.
	if !repo.upsertCalled {
		t.Fatal("UpsertDeviceToken was NOT called — SendPushNotification will find 0 tokens in Postgres")
	}
	if repo.upsertOwnerID != "42" {
		t.Errorf("upsert ownerID = %q, want \"42\"", repo.upsertOwnerID)
	}
	if repo.upsertOwnerType != "seller" {
		t.Errorf("upsert ownerType = %q, want \"seller\"", repo.upsertOwnerType)
	}
	if repo.upsertToken != testToken {
		t.Errorf("upsert token = %q, want %q", repo.upsertToken, testToken)
	}
}

// TestSaveFcmToken_AdminID_FallsBackToSellers verifies that when only AdminID is
// set (no ShopID) the token is still synced to the "sellers" collection.
func TestSaveFcmToken_AdminID_FallsBackToSellers(t *testing.T) {
	const (
		testToken   = "fcm-token-admin-xyz"
		testAdminID = uint(7)
	)

	repo := &mockFcmTokenRepo{
		saveFn: func(f domain.FcmToken) (domain.FcmToken, error) { return f, nil },
	}
	push := &mockPushSender{}
	uc := newTestFcmTokenUseCase(repo, push)

	_, err := uc.SaveFcmToken(domain.FcmToken{Token: testToken, AdminID: testAdminID})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !push.savedToFirestore {
		t.Fatal("SaveTokenToFirestore was NOT called for AdminID case")
	}
	if push.savedFirestoreCollection != "sellers" {
		t.Errorf("Firestore collection = %q, want \"sellers\"", push.savedFirestoreCollection)
	}
	if repo.upsertOwnerType != "seller" {
		t.Errorf("upsert ownerType = %q, want \"seller\"", repo.upsertOwnerType)
	}
}

// TestSaveFcmToken_NoOwnerID_SkipsSync verifies that a token with neither ShopID
// nor AdminID is stored in Postgres but does NOT trigger a Firestore sync or an
// notification_device_tokens upsert.
func TestSaveFcmToken_NoOwnerID_SkipsSync(t *testing.T) {
	repo := &mockFcmTokenRepo{
		saveFn: func(f domain.FcmToken) (domain.FcmToken, error) { return f, nil },
	}
	push := &mockPushSender{}
	uc := newTestFcmTokenUseCase(repo, push)

	_, err := uc.SaveFcmToken(domain.FcmToken{Token: "orphan-token"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if push.savedToFirestore {
		t.Error("SaveTokenToFirestore should NOT be called when both ShopID and AdminID are 0")
	}
	if repo.upsertCalled {
		t.Error("UpsertDeviceToken should NOT be called when both ShopID and AdminID are 0")
	}
}

// TestSaveFcmToken_RepoError_Propagates verifies that a DB error from SaveFcmToken
// short-circuits the function (no Firestore/Postgres sync attempted).
func TestSaveFcmToken_RepoError_Propagates(t *testing.T) {
	dbErr := errors.New("db: connection refused")
	repo := &mockFcmTokenRepo{
		saveFn: func(f domain.FcmToken) (domain.FcmToken, error) { return domain.FcmToken{}, dbErr },
	}
	push := &mockPushSender{}
	uc := newTestFcmTokenUseCase(repo, push)

	_, err := uc.SaveFcmToken(domain.FcmToken{Token: "t", ShopID: 1})
	if !errors.Is(err, dbErr) {
		t.Errorf("expected db error, got %v", err)
	}
	if push.savedToFirestore {
		t.Error("Firestore sync should not run when repo.SaveFcmToken fails")
	}
}

// TestSaveFcmToken_FirestoreError_DoesNotBlock verifies that a Firestore sync
// failure is logged but does not surface as an error to the caller.
func TestSaveFcmToken_FirestoreError_DoesNotBlock(t *testing.T) {
	repo := &mockFcmTokenRepo{
		saveFn: func(f domain.FcmToken) (domain.FcmToken, error) { return f, nil },
	}
	push := &mockPushSender{saveToFirestoreErr: errors.New("firebase: quota exceeded")}
	uc := newTestFcmTokenUseCase(repo, push)

	_, err := uc.SaveFcmToken(domain.FcmToken{Token: "t", ShopID: 99})
	if err != nil {
		t.Errorf("Firestore error must not bubble up; got: %v", err)
	}
	// The Postgres upsert must still happen even if Firestore failed.
	if !repo.upsertCalled {
		t.Error("UpsertDeviceToken must still be called even when Firestore sync fails")
	}
}

// TestSaveFcmToken_ShopID_TakesPrecedence verifies that when both ShopID and
// AdminID are set, ShopID is used as the owner.
func TestSaveFcmToken_ShopID_TakesPrecedence(t *testing.T) {
	repo := &mockFcmTokenRepo{
		saveFn: func(f domain.FcmToken) (domain.FcmToken, error) { return f, nil },
	}
	push := &mockPushSender{}
	uc := newTestFcmTokenUseCase(repo, push)

	_, err := uc.SaveFcmToken(domain.FcmToken{Token: "t", ShopID: 10, AdminID: 99})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if push.savedFirestoreOwnerID != "10" {
		t.Errorf("expected Firestore ownerID \"10\" (ShopID), got %q", push.savedFirestoreOwnerID)
	}
	if repo.upsertOwnerID != "10" {
		t.Errorf("expected upsert ownerID \"10\" (ShopID), got %q", repo.upsertOwnerID)
	}
}
