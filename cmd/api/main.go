package main

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/rohit221990/mandi-backend/pkg/api/middleware"
	"github.com/rohit221990/mandi-backend/pkg/config"
	"github.com/rohit221990/mandi-backend/pkg/db"
	"github.com/rohit221990/mandi-backend/pkg/di"
	applogger "github.com/rohit221990/mandi-backend/pkg/logger"
)

func main() {

	cfg, err := config.LoadConfig()

	if err != nil {
		log.Fatal("Error to load the config: ", err)
	}

	// Initialise structured logger as early as possible so all subsequent
	// startup messages are captured in the same format.
	// LOG_LEVEL env: debug | info | warn | error (default: info)
	// APP_ENV env: development | production (default: production)
	applogger.Init(applogger.Config{
		Level:       os.Getenv("LOG_LEVEL"),
		Development: os.Getenv("APP_ENV") == "development",
		LogDir:      "logs",
		ServiceName: "mandi-backend",
	})
	defer applogger.Sync()

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

	if cfg.Security.EnableTLS || cfg.Security.RateLimitingEnabled || cfg.Security.BruteForceProtectionEnabled || cfg.Security.HSTSMaxAge > 0 {
		server.Engine.Use(middleware.SecurityHeadersMiddleware(cfg))
		server.Engine.Use(middleware.SecurityMiddleware(cfg))
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

	if cfg.Security.EnableTLS {
		if cfg.Security.HTTPSRedirect {
			go func() {
				if err := runHTTPRedirectServer(cfg.Security.HTTPPort, cfg.Security.HTTPSPort); err != nil {
					log.Fatalf("failed to start HTTP redirect server: %v", err)
				}
			}()
		}
		if err := runHTTPSServer(server.Engine, cfg); err != nil {
			log.Fatal("failed to start TLS server: ", err)
		}
	} else {
		if startErr := server.Start(); startErr != nil {
			log.Fatal("failed to start server: ", startErr)
		}
	}
}

func runHTTPRedirectServer(httpPort, httpsPort string) error {
	redirect := func(w http.ResponseWriter, r *http.Request) {
		if r.TLS != nil {
			http.Error(w, "Use HTTPS", http.StatusBadRequest)
			return
		}

		host := r.Host
		if host == "" {
			host = "localhost"
		}
		if idx := strings.Index(host, ":"); idx != -1 {
			host = host[:idx]
		}
		if httpsPort != ":443" {
			host = fmt.Sprintf("%s%s", host, strings.TrimPrefix(httpsPort, ":"))
		}
		location := fmt.Sprintf("https://%s%s", host, r.URL.RequestURI())
		http.Redirect(w, r, location, http.StatusMovedPermanently)
	}

	return http.ListenAndServe(httpPort, http.HandlerFunc(redirect))
}

func runHTTPSServer(handler http.Handler, cfg config.Config) error {
	tlsConfig := &tls.Config{
		MinVersion:   parseTLSVersion(cfg.Security.TLSMinVersion),
		MaxVersion:   parseTLSVersion(cfg.Security.TLSMaxVersion),
		CipherSuites: parseCipherSuites(cfg.Security.CipherSuites),
	}

	server := &http.Server{
		Addr:         cfg.Security.HTTPSPort,
		Handler:      handler,
		TLSConfig:    tlsConfig,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server.ListenAndServeTLS(cfg.Security.TLSCertFile, cfg.Security.TLSKeyFile)
}

func parseTLSVersion(version string) uint16 {
	switch strings.TrimSpace(strings.ToLower(version)) {
	case "1.2", "tls1.2":
		return tls.VersionTLS12
	case "1.3", "tls1.3":
		return tls.VersionTLS13
	default:
		return tls.VersionTLS12
	}
}

func parseCipherSuites(names []string) []uint16 {
	if len(names) == 0 {
		return nil
	}
	var result []uint16
	for _, name := range names {
		switch strings.TrimSpace(strings.ToUpper(name)) {
		case "TLS_AES_128_GCM_SHA256":
			result = append(result, tls.TLS_AES_128_GCM_SHA256)
		case "TLS_AES_256_GCM_SHA384":
			result = append(result, tls.TLS_AES_256_GCM_SHA384)
		case "TLS_CHACHA20_POLY1305_SHA256":
			result = append(result, tls.TLS_CHACHA20_POLY1305_SHA256)
		case "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256":
			result = append(result, tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256)
		case "TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384":
			result = append(result, tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384)
		}
	}
	return result
}
