package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rohit221990/mandi-backend/internal/config"
	"github.com/rohit221990/mandi-backend/internal/handlers"
	"github.com/rohit221990/mandi-backend/internal/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.RequestLogger())
	router.Use(middleware.CORSMiddleware(cfg))
	router.MaxMultipartMemory = 5 << 20 // 5MB limit

	// Register handlers (wires repos/services inside)
	authHandler, err := handlers.NewAuthHandler(cfg)
	if err != nil {
		log.Fatalf("init handlers: %v", err)
	}

	api := router.Group("/auth")
	{
		api.POST("/register", authHandler.Register)
		api.POST("/login", authHandler.Login)
		api.POST("/send-otp", authHandler.SendOTP)
		api.POST("/verify-otp", authHandler.VerifyOTP)
		api.POST("/set-pin", authHandler.SetPIN)
		api.POST("/refresh-token", authHandler.Refresh)
		api.POST("/logout", authHandler.Logout)
	}

	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("starting server on %s", srv.Addr)
	if err := srv.ListenAndServe(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
