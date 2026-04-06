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

// SaveFcmToken persists the FCM token in Postgres (fcm_tokens table), syncs it
// to Firestore under sellers/{ownerID}/fcmTokens/{token}, and also writes into
// notification_device_tokens so the direct Postgres→FCM path works without Firestore.
//
// Owner resolution: ShopID takes precedence over AdminID.
func (u *fcmTokenUseCase) SaveFcmToken(fcmToken domain.FcmToken) (domain.FcmToken, error) {
	saved, err := u.repo.SaveFcmToken(fcmToken)
	if err != nil {
		return saved, err
	}

	// Resolve owner from canonical fields first, then compatibility fields.
	ownerID := saved.OwnerID
	ownerType := saved.OwnerType
	if ownerType == "" {
		ownerType = "seller"
	}
	if ownerID == "" {
		if saved.ShopID != 0 {
			ownerID = strconv.FormatUint(uint64(saved.ShopID), 10)
		} else if saved.AdminID != 0 {
			ownerID = strconv.FormatUint(uint64(saved.AdminID), 10)
		}
	}

	if ownerID == "" {
		log.Printf("WARN [SaveFcmToken]: token saved but no owner_id — skipping sync (token=%s)", saved.Token)
		return saved, nil
	}

	ctx := context.Background()

	// 1. Sync to Firestore so the Firestore watcher can deliver push notifications.
	if syncErr := u.fcmPush.SaveTokenToFirestore(ctx, ownerType+"s", ownerID, saved.Token, saved.Platform); syncErr != nil {
		log.Printf("WARN [SaveFcmToken]: Firestore sync failed for %s %s: %v", ownerType, ownerID, syncErr)
	}

	// 2. Upsert into notification_device_tokens so SendPushNotification finds the
	//    token via the Postgres path (no Firestore dependency required at send time).
	if syncErr := u.repo.UpsertDeviceToken(saved.Token, ownerID, ownerType, saved.Platform); syncErr != nil {
		log.Printf("WARN [SaveFcmToken]: notification_device_tokens upsert failed for %s %s: %v", ownerType, ownerID, syncErr)
	}

	return saved, nil
}
