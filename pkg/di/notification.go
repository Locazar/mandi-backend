package di

import (
	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/db"
	"github.com/rohit221990/mandi-backend/pkg/repository"
	"github.com/rohit221990/mandi-backend/pkg/usecase"
	usecaseinterfaces "github.com/rohit221990/mandi-backend/pkg/usecase/interfaces"
)

// InitializeNotificationUseCase builds a NotificationUseCase independently of
// the main Wire graph.  It is used by main.go to start the Firestore watcher
// without requiring changes to wire.go / wire_gen.go.
func InitializeNotificationUseCase(cfg config.Config) (usecaseinterfaces.NotificationUseCase, error) {
	gormDB, err := db.ConnectDatabase(cfg)
	if err != nil {
		return nil, err
	}

	notificationRepo := repository.NewNotificationRepository(gormDB)
	return usecase.NewNotificationUseCaseWithDB(notificationRepo, gormDB), nil
}
