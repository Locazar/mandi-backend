package usecase

import (
	"context"
	"fmt"
	"log"

	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	notificationSvc "github.com/rohit221990/mandi-backend/pkg/service/notification"
	"github.com/rohit221990/mandi-backend/pkg/service/token"
	usecase "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type fcmTokenUseCase struct {
	repo    interfaces.FcmTokenRepository
	fcmPush notificationSvc.PushSender
	token   token.TokenService
}

func NewFcmTokenUseCase(repo interfaces.FcmTokenRepository, tokenService token.TokenService) usecase.FcmTokenUseCase {
	return &fcmTokenUseCase{
		repo:    repo,
		fcmPush: notificationSvc.NewFCMPushService(),
		token:   tokenService,
	}
}

func (u *fcmTokenUseCase) DecodeTokenData(tokenString string) string {
	if u.token == nil {
		return ""
	}
	return u.token.DecodeTokenData(tokenString)
}

// SaveFcmToken persists the FCM token in Postgres (fcm_tokens table), syncs it
// to Firestore under sellers/{ownerID}/fcmTokens/{token}, and also writes into
// notification_device_tokens so the direct Postgres→FCM path works without Firestore.
//
// Owner resolution: ShopID takes precedence over AdminID.
func (u *fcmTokenUseCase) SaveFcmToken(fcmToken domain.FcmToken) (domain.FcmToken, error) {

	fmt.Printf("Saving FCM token: %s for owner_id: %s owner_type: %s\n", fcmToken.Token, fcmToken.OwnerID, fcmToken.OwnerType)
	fmt.Printf("Received FCM token data: ShopID=%s AdminID=%s OwnerID=%s OwnerType=%s\n", fcmToken.ShopID, fcmToken.AdminID, fcmToken.OwnerID, fcmToken.OwnerType)
	saved, err := u.repo.SaveFcmToken(fcmToken)
	if err != nil {
		return saved, err
	}

	// Resolve owner from canonical fields first, then compatibility fields.
	ownerID := saved.OwnerID
	ownerType := saved.OwnerType
	shopID := ""
	adminID := ""
	if ownerType == "" {
		ownerType = "seller"
	}
	if saved.ShopID != "" {
		shopID = saved.ShopID
	}
	if saved.AdminID != "" {
		adminID = saved.AdminID
	}
	if ownerID == "" {
		if saved.ShopID != "" {
			ownerID = saved.ShopID
		} else if saved.AdminID != "" {
			ownerID = saved.AdminID
		}
	}
	if adminID == "" && ownerType == "seller" {
		adminID = ownerID
	}
	if shopID == "" {
		if ownerType == "seller" {
			shopID = ownerID
		} else {
			shopID = adminID
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
	if syncErr := u.repo.UpsertDeviceToken(saved.Token, ownerID, ownerType, saved.Platform, shopID, adminID); syncErr != nil {
		log.Printf("WARN [SaveFcmToken]: notification_device_tokens upsert failed for %s %s: %v", ownerType, ownerID, syncErr)
	}

	return saved, nil
}

func (u *fcmTokenUseCase) UnregisterFcmToken(fcmToken domain.FcmToken) error {
	if err := u.repo.UnregisterFcmToken(fcmToken); err != nil {
		return err
	}

	ownerID := fcmToken.OwnerID
	ownerType := fcmToken.OwnerType
	if ownerType == "" {
		ownerType = "seller"
	}
	if ownerID == "" {
		if fcmToken.ShopID != "" {
			ownerID = fcmToken.ShopID
		} else if fcmToken.AdminID != "" {
			ownerID = fcmToken.AdminID
		}
	}

	if ownerID == "" || fcmToken.Token == "" {
		return nil
	}

	ctx := context.Background()
	if err := u.fcmPush.DeleteTokenFromFirestore(ctx, ownerType+"s", ownerID, fcmToken.Token); err != nil {
		log.Printf("WARN [UnregisterFcmToken]: Firestore cleanup failed for %s %s token=%s: %v", ownerType, ownerID, fcmToken.Token, err)
	}

	return nil
}
