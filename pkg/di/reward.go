package di

import (
	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/db"
	"github.com/rohit221990/mandi-backend/pkg/repository"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
)

// InitializeRewardUseCase builds a RewardUseCase outside the Wire graph for the
// in-process sweep ticker in cmd/api/main.go (same pattern as the nudge sweeps).
func InitializeRewardUseCase(cfg config.Config) (*usecase.RewardUseCase, error) {
	gormDB, err := db.ConnectDatabase(cfg)
	if err != nil {
		return nil, err
	}
	// Object storage is optional here: only used to absolutise push image keys.
	cloudService, err := cloud.NewObjectStorageService(cfg)
	if err != nil {
		cloudService = nil
	}
	notificationRepo := repository.NewNotificationRepository(gormDB)
	notificationUC := usecase.NewNotificationUseCaseWithDB(notificationRepo, gormDB, cfg, cloudService)
	return usecase.NewRewardUseCase(repository.NewRewardRepository(gormDB), notificationUC), nil
}
