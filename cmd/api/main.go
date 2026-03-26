package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/db"
	"github.com/rohit221990/mandi-backend/pkg/di"
)

func main() {

	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatal("Error to load the config: ", err)
	}

	// Connect to database and seed data
	dbConn, err := db.ConnectDatabase(cfg)
	if err != nil {
		log.Fatal("Failed to connect to database: ", err)
	}

	err = db.SeedProductItemFilters(dbConn)
	if err != nil {
		log.Printf("Warning: Failed to seed data: %v", err)
	}

	server, err := di.InitializeApi(cfg)
	if err != nil {
		log.Fatal("Failed to initialize the api: ", err)
	}

	// Start the Firestore watcher in the background.
	// It sends FCM push notifications when monitored document fields change.
	// Cancel the context on shutdown to stop all watcher goroutines cleanly.
	watcherCtx, cancelWatcher := context.WithCancel(context.Background())

	notifUC, initErr := di.InitializeNotificationUseCase(cfg)
	if initErr != nil {
		log.Printf("Warning: Could not initialize notification use-case for Firestore watcher: %v", initErr)
		cancelWatcher()
	} else {
		// nil rules → uses the four default e-commerce rules
		if startErr := notifUC.StartFirestoreWatcher(watcherCtx, nil); startErr != nil {
			log.Printf("Warning: Firestore watcher could not start (Firebase credentials may be missing): %v", startErr)
		}
	}

	// Graceful shutdown: stop watcher when OS signal is received
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-quit
		log.Println("Shutdown signal received — stopping Firestore watcher")
		cancelWatcher()
	}()

	if startErr := server.Start(); startErr != nil {
		log.Fatal("failed to start server: ", startErr)
	}
}
