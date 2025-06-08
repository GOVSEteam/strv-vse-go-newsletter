// Newsletter API - A comprehensive API for managing newsletters, posts, and subscribers
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
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/setup"

	_ "github.com/joho/godotenv/autoload"
	"go.uber.org/zap"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Error loading configuration: %v", err)
	}

	// Initialize logger
	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatalf("Error initializing logger: %v", err)
	}
	defer logger.Sync()
	sugar := logger.Sugar()

	// Create context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize Database Connection
	dbPool, err := setup.ConnectDB(ctx, cfg.GetDatabaseURL())
	if err != nil {
		sugar.Fatalf("Error connecting to database: %v", err)
	}
	defer dbPool.Close()

	// Initialize Firebase
	firebaseApp, err := setup.NewFirebaseApp(ctx, cfg.FirebaseServiceAccount)
	if err != nil {
		sugar.Fatalf("Error initializing Firebase app: %v", err)
	}

	firebaseAuthClient, err := setup.NewAuthClient(ctx, firebaseApp)
	if err != nil {
		sugar.Fatalf("Error initializing Firebase Auth client: %v", err)
	}

	firestoreClient, err := setup.NewFirestoreClient(ctx, firebaseApp)
	if err != nil {
		sugar.Fatalf("Error initializing Firestore client: %v", err)
	}
	defer func() {
		if firestoreClient != nil {
			firestoreClient.Close()
		}
	}()

	// Initialize Repositories
	editorRepo := repository.NewPostgresEditorRepo(dbPool)
	newsletterRepo := repository.NewPostgresNewsletterRepo(dbPool)
	postRepo := repository.NewPostRepository(dbPool)
	subscriberRepo := repository.NewFirestoreSubscriberRepository(firestoreClient)

	// Initialize Email Service (direct sending)
	emailServiceConfig := service.GmailEmailServiceConfig{
		From:     cfg.EmailFrom,
		Password: cfg.GoogleAppPassword,
		SMTPHost: cfg.SMTPHost,
		SMTPPort: cfg.SMTPPort,
	}
	emailService, err := service.NewGmailEmailService(emailServiceConfig, zap.NewStdLog(logger))
	if err != nil {
		sugar.Fatalf("Error initializing email service: %v", err)
	}

	// Initialize Services
	firebasePasswordResetConfig := service.FirebasePasswordResetServiceConfig{
		APIKey:     cfg.FirebaseAPIKey,
		HTTPClient: &http.Client{Timeout: 10 * time.Second},
		Logger:     zap.NewStdLog(logger),
	}
	passwordResetSvc, err := service.NewFirebasePasswordResetService(firebasePasswordResetConfig)
	if err != nil {
		sugar.Fatalf("Error initializing password reset service: %v", err)
	}
	editorSvc := service.NewEditorService(editorRepo, setup.NewFirebaseAuthAdapter(firebaseAuthClient), &http.Client{Timeout: 10 * time.Second}, cfg.FirebaseAPIKey)
	newsletterSvc := service.NewNewsletterService(newsletterRepo, postRepo)
	subscriberSvc := service.NewSubscriberService(subscriberRepo, newsletterRepo, editorRepo, emailService, cfg.AppBaseURL)
	publishingSvc := service.NewPublishingService(newsletterSvc, subscriberSvc, emailService, cfg)

	// Initialize Router
	routerDeps := router.RouterDependencies{
		DB:                dbPool,
		AuthClient:        middleware.NewFirebaseAuthAdapter(firebaseAuthClient),
		NewsletterService: newsletterSvc,
		SubscriberService: subscriberSvc,
		PublishingService: publishingSvc,
		EditorService:     editorSvc,
		PasswordResetSvc:  passwordResetSvc,
		EditorRepo:        editorRepo,
		Logger:            sugar,
	}
	mainRouter := router.NewRouter(routerDeps)

	// HTTP Server Setup
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Port),
		Handler:      mainRouter,
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
		
		// Create shutdown context with timeout
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer shutdownCancel()

		// Stop accepting new HTTP requests (this blocks until server is down)
		sugar.Info("Shutting down HTTP server...")
		if err := server.Shutdown(shutdownCtx); err != nil {
			sugar.Errorf("HTTP server shutdown error: %v", err)
		} else {
			sugar.Info("HTTP server shut down gracefully")
		}
	}()

	sugar.Infof("Starting server on port %d...", cfg.Port)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		sugar.Fatalf("Could not start server: %v", err)
	}

	sugar.Info("Server stopped.")
}
