package di

import (
	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/db"
	"github.com/rohit221990/mandi-backend/pkg/repository"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
	usecaseinterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

// InitializeCustomerNudgeUseCase builds a CustomerNudgeUseCase independently
// of the main Wire graph, for the in-process sweep ticker in cmd/api.
func InitializeCustomerNudgeUseCase(cfg config.Config) (usecaseinterfaces.CustomerNudgeUseCase, error) {
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
	return usecase.NewCustomerNudgeUseCase(repository.NewCustomerNudgeRepository(gormDB), notificationUC), nil
}
