package di

import (
	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/db"
	"github.com/rohit221990/mandi-backend/pkg/repository"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
	usecaseinterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

// InitializeDistrictNudgeUseCase builds a DistrictNudgeUseCase independently
// of the main Wire graph — used by cmd/api/main.go's daily sweep ticker,
// mirroring InitializeOnboardingNudgeUseCase's role for its own ticker.
func InitializeDistrictNudgeUseCase(cfg config.Config) (usecaseinterfaces.DistrictNudgeUseCase, error) {
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

	districtNudgeRepo := repository.NewDistrictNudgeRepository(gormDB)
	return usecase.NewDistrictNudgeUseCase(districtNudgeRepo, notificationUC), nil
}
