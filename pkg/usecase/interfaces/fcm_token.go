package interfaces

import "github.com/rohit221990/mandi-backend/pkg/domain"

type FcmTokenUseCase interface {
	SaveFcmToken(fcmToken domain.FcmToken) (domain.FcmToken, error)
	UnregisterFcmToken(fcmToken domain.FcmToken) error
	DecodeTokenData(tokenString string) string
}
