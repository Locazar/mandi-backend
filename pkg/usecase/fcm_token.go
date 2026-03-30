package usecase

import (
	"context"
	"log"
	"strconv"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	notificationSvc "github.com/rohit221990/mandi-backend/pkg/service/notification"
	usecase "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type fcmTokenUseCase struct {
	repo    interfaces.FcmTokenRepository
	fcmPush notificationSvc.PushSender
}

func NewFcmTokenUseCase(repo interfaces.FcmTokenRepository) usecase.FcmTokenUseCase {
	return &fcmTokenUseCase{
		repo:    repo,
		fcmPush: notificationSvc.NewFCMPushService(),
	}
}

// SaveFcmToken persists the FCM token in Postgres and syncs it to Firestore so
// the backend Firestore watcher can dispatch push notifications when monitored
// documents (e.g. enquiries) are created or updated.
func (u *fcmTokenUseCase) SaveFcmToken(fcmToken domain.FcmToken) (domain.FcmToken, error) {
	saved, err := u.repo.SaveFcmToken(fcmToken)
	if err != nil {
		return saved, err
	}

	ctx := context.Background()
	if saved.ShopID != 0 {
		ownerID := strconv.FormatUint(uint64(saved.ShopID), 10)
		if syncErr := u.fcmPush.SaveTokenToFirestore(ctx, "sellers", ownerID, saved.Token, saved.Platform); syncErr != nil {
			log.Printf("WARN [SaveFcmToken]: Firestore sync failed for seller %s: %v", ownerID, syncErr)
		}
	}
	if saved.AdminID != 0 {
		ownerID := strconv.FormatUint(uint64(saved.AdminID), 10)
		if syncErr := u.fcmPush.SaveTokenToFirestore(ctx, "admins", ownerID, saved.Token, saved.Platform); syncErr != nil {
			log.Printf("WARN [SaveFcmToken]: Firestore sync failed for admin %s: %v", ownerID, syncErr)
		}
	}

	return saved, nil
}
