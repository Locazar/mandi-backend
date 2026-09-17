package di

import (
	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/db"
	"github.com/rohit221990/mandi-backend/pkg/repository"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
	usecaseinterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

// InitializeOnboardingNudgeUseCase builds an OnboardingNudgeUseCase
// independently of the main Wire graph — used by cmd/onboarding-nudges,
// mirroring InitializeNotificationUseCase's role for cmd/enquiry-autoreject.
func InitializeOnboardingNudgeUseCase(cfg config.Config) (usecaseinterfaces.OnboardingNudgeUseCase, error) {
	gormDB, err := db.ConnectDatabase(cfg)
	if err != nil {
		return nil, err
	}

	// Object storage is optional here, same reasoning as the notification
	// watcher: only needed to absolutise push image keys.
	cloudService, err := cloud.NewObjectStorageService(cfg)
	if err != nil {
		cloudService = nil
	}

	notificationRepo := repository.NewNotificationRepository(gormDB)
	notificationUC := usecase.NewNotificationUseCaseWithDB(notificationRepo, gormDB, cfg, cloudService)

	onboardingNudgeRepo := repository.NewOnboardingNudgeRepository(gormDB)
	return usecase.NewOnboardingNudgeUseCase(onboardingNudgeRepo, notificationUC), nil
}
