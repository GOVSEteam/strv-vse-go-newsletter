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
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/router"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/setup"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/worker"

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
	sugar.Infof("Main context created: %v (used for graceful shutdown and DI)", ctx)

	// Initialize Database Connection
	sugar.Info("Initializing database connection...")
	dbPool, err := setup.ConnectDB(ctx, cfg.GetDatabaseURL())
	if err != nil {
		sugar.Fatalf("Error connecting to database: %v", err)
	}
	sugar.Info("Database connection established")
	defer func() {
		sugar.Info("Closing database connection pool...")
		if dbPool != nil {
			dbPool.Close()
		}
		sugar.Info("Database connection pool closed.")
	}()

	// Initialize Firebase
	sugar.Info("Initializing Firebase app...")
	firebaseApp, err := setup.NewFirebaseApp(ctx, cfg.FirebaseServiceAccount)
	if err != nil {
		sugar.Fatalf("Error initializing Firebase app: %v", err)
	}
	sugar.Info("Firebase app initialized")

	firebaseAuthClient, err := setup.NewAuthClient(ctx, firebaseApp)
	if err != nil {
		sugar.Fatalf("Error initializing Firebase Auth client: %v", err)
	}
	sugar.Info("Firebase Auth client initialized")

	firestoreClient, err := setup.NewFirestoreClient(ctx, firebaseApp)
	if err != nil {
		sugar.Fatalf("Error initializing Firestore client: %v", err)
	}
	sugar.Info("Firestore client initialized")
	defer func() {
		sugar.Info("Closing Firestore client...")
		if firestoreClient != nil {
			if err := firestoreClient.Close(); err != nil {
				sugar.Errorf("Error closing Firestore client: %v", err)
			} else {
				sugar.Info("Firestore client closed.")
			}
		}
	}()

	// Initialize Repositories
	sugar.Info("Initializing repositories...")
	editorRepo := repository.EditorRepo(dbPool)
	newsletterRepo := repository.NewsletterRepo(dbPool)
	postRepo := repository.NewPostRepository(dbPool)
	subscriberRepo := repository.NewFirestoreSubscriberRepository(firestoreClient)
	sugar.Info("Repositories initialized")

	// Initialize Email Service and Worker
	sugar.Info("Initializing email service and worker...")
	emailServiceConfig := service.GmailEmailServiceConfig{
		From:     cfg.EmailFrom,
		Password: cfg.GoogleAppPassword,
		SMTPHost: "smtp.gmail.com", // Default, consider adding to config.Config
		SMTPPort: "587",            // Default, consider adding to config.Config
	}
	emailService, err := service.NewGmailEmailService(emailServiceConfig, zap.NewStdLog(logger))
	if err != nil {
		sugar.Fatalf("Error initializing email service: %v", err)
	}
	emailWorker := worker.NewEmailWorker(emailService, 100) // emailWorker is the EmailJobQueuer
	go emailWorker.Start(ctx, 5)
	sugar.Info("Email service and worker initialized")

	// Initialize Services
	sugar.Info("Initializing services...")
	firebasePasswordResetConfig := service.FirebasePasswordResetServiceConfig{
		APIKey:     cfg.FirebaseAPIKey,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		Logger:     zap.NewStdLog(logger),
	}
	passwordResetSvc, err := service.NewFirebasePasswordResetService(firebasePasswordResetConfig) // Renamed from authSvc to match router
	if err != nil {
		sugar.Fatalf("Error initializing password reset service: %v", err)
	}
	editorSvc := service.NewEditorService(editorRepo, firebaseAuthClient, &http.Client{Timeout: 10 * time.Second}, cfg.FirebaseAPIKey)
	newsletterSvc := service.NewNewsletterService(newsletterRepo, postRepo, editorRepo)
	subscriberSvc := service.NewSubscriberService(subscriberRepo, newsletterRepo, editorRepo, emailService, cfg.AppBaseURL)
	publishingSvc := service.NewPublishingService(newsletterSvc, subscriberSvc, emailWorker) // Added PublishingService
	sugar.Info("Services initialized")

	// Initialize Chi Router
	sugar.Info("Initializing Chi router...")
	routerDeps := router.RouterDependencies{
		DB:                  dbPool, // The router expects *sql.DB, but we have *pgxpool.Pool. This needs to be addressed.
		FirestoreClient:     firestoreClient,
		FirebaseAuthClient:  firebaseAuthClient,
		EditorService:       editorSvc,
		NewsletterService:   newsletterSvc,
		SubscriberService:   subscriberSvc,
		PublishingService:   publishingSvc,
		PasswordResetSvc:    passwordResetSvc,
		EmailJobQueuer:      emailWorker,
		EditorRepo:          editorRepo,
		Logger:              sugar, // Pass the sugared logger
	}
	mainRouter := router.NewRouter(routerDeps)
	sugar.Info("Chi router initialized")

	// HTTP Server Setup
	server := &http.Server{
		Addr:    fmt.Sprintf(":%d", cfg.Port),
		Handler: mainRouter, // Will be mainRouter (Chi router)
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
