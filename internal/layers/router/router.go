package router

import (
	// "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler" // Old monolithic handler, no longer needed directly here
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/editor"
	newsletterHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/newsletter" // New specific newsletter handlers
	subscriberHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/subscriber" // Added subscriber handler import
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/pkg/email" // Added email package
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/setup"      // Updated path
	"log"
	"net/http"
)

// SetupRouter configures the HTTP routes.
// It now takes individual handler functions for newsletter and editor, and a struct for subscriber.
func SetupRouter(
	mux *http.ServeMux, // Pass in the mux to attach routes
	// Subscriber Handler (struct with methods)
	subH *subscriberHandler.SubscriberHandler,
	// Newsletter Handlers (individual functions)
	createNewsletterH http.HandlerFunc,
	listNewslettersH http.HandlerFunc,
	updateNewsletterH http.HandlerFunc,
	deleteNewsletterH http.HandlerFunc,
	// Editor Handlers (individual functions)
	signUpH http.HandlerFunc,
	signInH http.HandlerFunc,
) {
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Newsletter routes
	mux.HandleFunc("POST /api/newsletters", createNewsletterH)
	mux.HandleFunc("GET /api/newsletters", listNewslettersH)
	mux.HandleFunc("PATCH /api/newsletters/{id}", updateNewsletterH)
	mux.HandleFunc("DELETE /api/newsletters/{id}", deleteNewsletterH)

	// Subscriber routes
	mux.HandleFunc("POST /api/newsletters/{newsletterID}/subscribe", subH.SubscribeToNewsletter)
	mux.HandleFunc("DELETE /api/newsletters/{newsletterID}/subscribers", subH.UnsubscribeFromNewsletter)
	mux.HandleFunc("GET /api/subscribers/confirm", subH.ConfirmSubscriptionHandler)

	// Editor/Auth routes
	mux.HandleFunc("POST /signup", signUpH)
	mux.HandleFunc("POST /signin", signInH)
}

// InitializeAndSetupRouter initializes all dependencies and sets up the router.
func InitializeAndSetupRouter() http.Handler {
	log.Println("Initializing database connection...")
	db := setup.ConnectDB() // Updated path
	if db == nil {
		log.Fatal("Failed to connect to database, cannot start server.")
	}
	log.Println("Database connection successful.")

	log.Println("Initializing Firebase...")
	setup.InitFirebase() // Updated path
	firestoreClient := setup.GetFirestoreClient() // Updated path
	if firestoreClient == nil {
		log.Fatal("Failed to initialize Firestore client, cannot start server.")
	}
	log.Println("Firebase Firestore client initialized.")

	// Initialize repositories
	newsletterRepo := repository.NewsletterRepo(db)
	editorRepo := repository.EditorRepo(db)
	subscriberRepo := repository.NewFirestoreSubscriberRepository(firestoreClient)

	// Initialize services
	emailSvc := email.NewConsoleEmailService() // Initialize EmailService

	var newsletterSvc service.NewsletterServiceInterface = service.NewsletterService(newsletterRepo)
	var editorSvc service.EditorService = service.NewEditorService(editorRepo)
	subscriberSvc := service.NewSubscriberService(subscriberRepo, newsletterRepo, emailSvc) // Pass emailSvc

	// Initialize handlers
	// Subscriber handler (struct)
	subH := subscriberHandler.NewSubscriberHandler(subscriberSvc)

	// Newsletter handlers (individual functions)
	createNewsletterH := newsletterHandler.CreateHandler(newsletterSvc, editorRepo)
	listNewslettersH := newsletterHandler.ListHandler(newsletterSvc, editorRepo) // Assuming ListHandler exists
	updateNewsletterH := newsletterHandler.UpdateHandler(newsletterSvc, editorRepo) // Assuming UpdateHandler exists
	deleteNewsletterH := newsletterHandler.DeleteHandler(newsletterSvc, editorRepo) // Assuming DeleteHandler exists

	// Editor handlers (individual functions)
	signUpH := editor.EditorSignUpHandler(editorSvc)
	signInH := editor.EditorSignInHandler(editorSvc) // Assuming EditorSignInHandler exists

	// Create the mux and pass it to SetupRouter
	mux := http.NewServeMux()
	SetupRouter(
		mux,
		subH,
		createNewsletterH,
		listNewslettersH,
		updateNewsletterH,
		deleteNewsletterH,
		signUpH,
		signInH,
	)

	log.Println("Router setup complete.")
	return mux
}
