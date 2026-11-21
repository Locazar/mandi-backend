package di

import (
	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/db"
	"github.com/rohit221990/mandi-backend/pkg/repository"
	"github.com/rohit221990/mandi-backend/pkg/service/cloud"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
)

// InitializeShopVisitDigest builds the nightly shop-visit digest independently
// of the main Wire graph — used by cmd/api/main.go's 23:50 IST scheduler.
func InitializeShopVisitDigest(cfg config.Config) (*usecase.ShopVisitDigest, error) {
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
	return usecase.NewShopVisitDigest(gormDB, notificationUC), nil
}
