package usecase

import (
	"github.com/rohit221990/mandi-backend/pkg/domain"
	"github.com/rohit221990/mandi-backend/pkg/repository/interfaces"
	usecase "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

type fcmTokenUseCase struct {
	repo interfaces.FcmTokenRepository
}

func NewFcmTokenUseCase(repo interfaces.FcmTokenRepository) usecase.FcmTokenUseCase {
	return &fcmTokenUseCase{repo}
}

func (u *fcmTokenUseCase) SaveFcmToken(fcmToken domain.FcmToken) (domain.FcmToken, error) {
	return u.repo.SaveFcmToken(fcmToken)
}
