package interfaces

import "github.com/rohit221990/mandi-backend/pkg/domain"

type FcmTokenUseCase interface {
	SaveFcmToken(fcmToken domain.FcmToken) (domain.FcmToken, error)
}
