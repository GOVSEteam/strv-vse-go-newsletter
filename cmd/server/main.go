// Newsletter API
//
// A comprehensive API for managing newsletters, posts, and subscribers
//
//	Title: Newsletter API
//	Description: A comprehensive API for managing newsletters, posts, and subscribers
//	Version: 1.0.0
//	Host: localhost:8080 // Will be replaced by config.AppBaseURL
//	BasePath: /
//	Schemes: http, https
//
//	SecurityDefinitions:
//	  BearerAuth:
//	    type: apiKey
//	    in: header
//	    name: Authorization
//	    description: Firebase JWT token (add 'Bearer ' prefix)
//
// @contact.name Newsletter API Support
// @contact.email support@newsletter-api.com
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/config"
	// "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler" // Placeholder
	// "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository" // Placeholder
	// "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/router" // Placeholder for new chi router
	// "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service" // Placeholder
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/setup"
	// "github.com/GOVSEteam/strv-vse-go-newsletter/internal/worker" // Placeholder

	_ "github.com/joho/godotenv/autoload" // Autoload .env file
	"go.uber.org/zap"
	// "github.com/jackc/pgx/v5/pgxpool" // Placeholder for pgxpool
)

var _ = setup.ConnectDB // Dummy use of setup package until its functions are fully integrated

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize structured logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}
	defer logger.Sync() // Flushes buffer, if any
	sugar := logger.Sugar()
	sugar.Info("Logger initialized")
	sugar.Infof("Application configuration loaded: %+v", cfg)

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sugar.Infof("Main context created: %v (used for graceful shutdown and DI)", ctx) // Dummy use

	// Initialize Database Connection (Placeholder for Step 27)
	sugar.Info("Initializing database connection...")
	// dbPool, err := setup.ConnectDB(ctx, cfg.GetDatabaseURL()) // ConnectDB will be refactored to return *pgxpool.Pool and error
	// The existing setup.ConnectDB() returns *sql.DB and panics on error, no ctx or URL yet.
	// For now, let's assume it returns a generic closable interface to allow compilation.
	type closable interface{ Close() error }
	var dbPool closable // This is a temporary placeholder
	// sqlDB := setup.ConnectDB() // Call existing function, will be replaced
	// dbPool = sqlDB // Assign to the interface
	sugar.Infof("dbPool placeholder: %v (actual setup in Step 27)", dbPool) // Dummy use

	// if err != nil { // This error handling will be used when ConnectDB is refactored
	// 	sugar.Fatalf("Error connecting to database: %v", err)
	// }
	sugar.Info("Database connection established (placeholder)")
	// // Close database connection on shutdown (pgxpool.Pool has Close method)
	// defer func() {
	// 	sugar.Info("Closing database connection pool...")
	// 	if dbPool != nil {
	// 	 dbPool.Close() // This will be *pgxpool.Pool in the future
	// 	}
	// 	sugar.Info("Database connection pool closed (placeholder).")
	// }()

	// Initialize Firebase (Placeholder for Step 28)
	sugar.Info("Initializing Firebase app...")
	// firebaseApp, err := setup.NewFirebaseApp(ctx, cfg.FirebaseServiceAccount)
	// if err != nil {
	// 	sugar.Fatalf("Error initializing Firebase app: %v", err)
	// }
	// sugar.Info("Firebase app initialized")

	// firebaseAuthClient, err := setup.NewAuthClient(ctx, firebaseApp)
	// if err != nil {
	// 	sugar.Fatalf("Error initializing Firebase Auth client: %v", err)
	// }
	// sugar.Info("Firebase Auth client initialized")

	// firestoreClient, err := setup.NewFirestoreClient(ctx, firebaseApp)
	// if err != nil {
	// 	sugar.Fatalf("Error initializing Firestore client: %v", err)
	// }
	// sugar.Info("Firestore client initialized")
	var firestoreClient closable // Placeholder for *firestore.Client
	sugar.Infof("firestoreClient placeholder: %v (actual setup in Step 28)", firestoreClient) // Dummy use
	// // Close Firestore client on shutdown
	// defer func() {
	// 	sugar.Info("Closing Firestore client...")
	// 	if firestoreClient != nil {
	// 	 if err := firestoreClient.Close(); err != nil {
	// 		 sugar.Errorf("Error closing Firestore client: %v", err)
	// 	 } else {
	// 		 sugar.Info("Firestore client closed (placeholder).")
	// 	 }
	// 	}
	// }()

	// Initialize Repositories (Placeholder for Steps 13-16)
	sugar.Info("Initializing repositories...")
	// editorRepo := repository.NewEditorRepository(dbPool) // Example placeholder
	// newsletterRepo := repository.NewNewsletterRepository(dbPool) // Example placeholder
	// postRepo := repository.NewPostRepository(dbPool) // Example placeholder
	// subscriberRepo := repository.NewSubscriberRepository(firestoreClient) // Example placeholder
	sugar.Info("Repositories initialized (placeholders)")

	// Initialize Email Service and Worker (Placeholders for Step 9 and 25)
	sugar.Info("Initializing email service and worker...")
	// emailService := service.NewGmailEmailService(cfg) // Example placeholder
	// emailWorker := worker.NewEmailWorker(emailService, 100) // Example placeholder
	// go emailWorker.Start(ctx, 5) // Start worker pool
	sugar.Info("Email service and worker initialized (placeholders)")

	// Initialize Services (Placeholder for Steps 17-20, and 10)
	sugar.Info("Initializing services...")
	// authSvc := service.NewFirebasePasswordResetService(cfg, &http.Client{}) // Example placeholder
	// editorSvc := service.NewEditorService(editorRepo, firebaseAuthClient) // Example placeholder
	// newsletterSvc := service.NewNewsletterService(newsletterRepo) // Example placeholder
	// postSvc := service.NewPostService(postRepo) // Example placeholder
	// subscriberSvc := service.NewSubscriberService(subscriberRepo, emailService) // Example placeholder
	sugar.Info("Services initialized (placeholders)")

	// Initialize Handlers (Placeholder for Steps 21-24, and 30)
	sugar.Info("Initializing handlers...")
	// healthHandler := handler.NewHealthHandler(dbPool, firestoreClient) // Example placeholder
	// editorHandler := handler.NewEditorHandler(editorSvc, authSvc) // Example placeholder
	// newsletterHandler := handler.NewNewsletterHandler(newsletterSvc) // Example placeholder
	// postHandler := handler.NewPostHandler(postSvc) // Example placeholder
	// subscriberHandler := handler.NewSubscriberHandler(subscriberSvc) // Example placeholder
	sugar.Info("Handlers initialized (placeholders)")

	// Initialize Chi Router with Middleware and Handlers (Placeholder for Step 26)
	sugar.Info("Initializing Chi router...")
	// mainRouter := router.NewChiRouter( // This will be the new Chi router setup
	// 	logger.Sugar(),
	// 	cfg, // For CORS, rate limiting config
	// 	firebaseAuthClient, // For auth middleware
	// 	editorRepo, // For auth middleware
	// 	healthHandler,
	// 	editorHandler,
	// 	newsletterHandler,
	// 	postHandler,
	// 	subscriberHandler,
	// ) // Example placeholder
	// Temp placeholder for existing router to allow compilation until Step 26
	tempRouter := http.NewServeMux()
	tempRouter.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "pong")
	})
	sugar.Info("Chi router initialized (placeholder - using temp simple router)")

	// HTTP Server Setup
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: tempRouter, // Will be mainRouter (Chi router)
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful Shutdown Handling
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan

		sugar.Info("Shutdown signal received, initiating graceful shutdown...")
		cancel() // Signal cancellation to other parts of the application

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			sugar.Fatalf("HTTP server Shutdown: %v", err)
		}
		sugar.Info("HTTP server shut down gracefully")
	}()

	sugar.Infof("Starting server on port %d...", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		sugar.Fatalf("Could not start server: %v", err)
	}

	sugar.Info("Server stopped.")
}
