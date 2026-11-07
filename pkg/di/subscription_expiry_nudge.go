package di

import (
	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/db"
	"github.com/rohit221990/mandi-backend/pkg/repository"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
	usecaseinterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

// InitializeSubscriptionExpiryNudgeUseCase builds a
// SubscriptionExpiryNudgeUseCase independently of the main Wire graph —
// used by cmd/api/main.go's daily sweep ticker.
func InitializeSubscriptionExpiryNudgeUseCase(cfg config.Config) (usecaseinterfaces.SubscriptionExpiryNudgeUseCase, error) {
	gormDB, err := db.ConnectDatabase(cfg)
	if err != nil {
		return nil, err
	}

	cloudService, err := cloud.NewObjectStorageService(cfg)
	if err != nil {
		cloudService = nil
	}

	notificationRepo := repository.NewNotificationRepository(gormDB)
	notificationUC := usecase.NewNotificationUseCaseWithDB(notificationRepo, gormDB, cfg, cloudService)

	expiryNudgeRepo := repository.NewSubscriptionExpiryNudgeRepository(gormDB)
	return usecase.NewSubscriptionExpiryNudgeUseCase(expiryNudgeRepo, notificationUC), nil
}
