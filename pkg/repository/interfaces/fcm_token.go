package interfaces

import "github.com/rohit221990/mandi-backend/pkg/domain"

type FcmTokenRepository interface {
	SaveFcmToken(fcmToken domain.FcmToken) (domain.FcmToken, error)
	// UpsertDeviceToken writes into notification_device_tokens so the direct
	// Postgres→FCM path in SendPushNotification can find it.
	UpsertDeviceToken(token, ownerID, ownerType, platform string) error
}
