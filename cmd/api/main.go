package main

import (
	"log"

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

	if server.Start(); err != nil {
		log.Fatal("failed to start server: ", err)
	}
}
