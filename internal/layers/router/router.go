package router

import (
	"net/http"

	"log"
	"os"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/editor"
	newsletterHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/newsletter" // Import specific newsletter handlers
	postHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/post"             // Import post handlers
	subscriberHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/subscriber"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/pkg/email" // Added email package
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/setup"
)

func Router() http.Handler {
	mux := http.NewServeMux()

	db := setup.ConnectDB()

	editorRepo := repository.EditorRepo(db) // editorRepo needs to be initialized before newsletterService
	editorService := service.NewEditorService(editorRepo)

	newsletterRepo := repository.NewsletterRepo(db)
	postRepo := repository.NewPostRepository(db) // Instantiate PostRepository
	
	// Update NewsletterService instantiation
	newsletterService := service.NewNewsletterService(newsletterRepo, postRepo, editorRepo)

	// Initialize EmailService
	var emailSvc email.EmailService
	var err error
	if os.Getenv("RESEND_API_KEY") != "" {
		emailSvc, err = email.NewResendService()
		if err != nil {
			log.Fatalf("Error initializing Resend email service: %v", err)
		}
		log.Println("Using Resend email service")
	} else {
		emailSvc = email.NewConsoleEmailService()
		log.Println("Using Console email service (RESEND_API_KEY not set)")
	}

	// Initialize Firebase client for SubscriberRepository
	firestoreClient := setup.GetFirestoreClient() // Use the existing getter
	if firestoreClient == nil { // Should not happen if GetFirestoreClient initializes
		log.Fatal("Failed to get Firestore client")
	}
	subscriberRepo := repository.NewFirestoreSubscriberRepository(firestoreClient)

	// Initialize SubscriberService
	subscriberService := service.NewSubscriberService(subscriberRepo, newsletterRepo, emailSvc)


	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Editor routes - assuming these are still correct
	mux.HandleFunc("/editor/signup", editor.EditorSignUpHandler(editorService))
	mux.HandleFunc("/editor/signin", editor.EditorSignInHandler(editorService))
	// Assuming FirebasePasswordResetRequestHandler is independent or correctly defined elsewhere
	mux.HandleFunc("/editor/password-reset-request", editor.FirebasePasswordResetRequestHandler())

	// Handler for /api/newsletters (collection)
	mux.HandleFunc("/api/newsletters", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			newsletterHandler.ListHandler(newsletterService, editorRepo)(w, r)
		case http.MethodPost:
			newsletterHandler.CreateHandler(newsletterService, editorRepo)(w, r)
		default:
			http.Error(w, "Method not allowed for /api/newsletters collection", http.StatusMethodNotAllowed)
		}
	})

	// Handler for /api/newsletters/{id} (specific resource)
	// Note: Go 1.22+ http.ServeMux supports patterns like /api/newsletters/{id}/
	// For older Go versions or more complex routing, a third-party router is better.
	// Assuming Go 1.22+ for r.PathValue("id") in handlers.
	mux.HandleFunc("/api/newsletters/{id}", func(w http.ResponseWriter, r *http.Request) {
		// id := r.PathValue("id") // This is how handlers will get it.
		// No need to parse id here if handlers do it.
		switch r.Method {
		// case http.MethodGet:
			// TODO: Implement GetNewsletterByIDHandler
			// newsletterHandler.GetByIDHandler(newsletterService, editorRepo)(w, r)
		case http.MethodPatch: // Assuming PATCH for updates as per update.go
			newsletterHandler.UpdateHandler(newsletterService, editorRepo)(w, r)
		case http.MethodDelete:
			newsletterHandler.DeleteHandler(newsletterService, editorRepo)(w, r)
		default:
			http.Error(w, "Method not allowed for /api/newsletters/{id}", http.StatusMethodNotAllowed)
		}
	})

	// Post routes
	// Create a post for a specific newsletter & List posts for a specific newsletter
	mux.HandleFunc("/api/newsletters/{newsletterID}/posts", func(w http.ResponseWriter, r *http.Request) {
		// editorRepo is needed by CreatePostHandler for now, until auth middleware provides editor context directly
		// ListPostsByNewsletterHandler doesn't strictly need editorRepo if auth is just for access control via JWT
		// but CreatePostHandler passes it to the service layer implicitly via editorFirebaseUID.
		// For consistency, or if ListPostsByNewsletterHandler were to evolve to need it, it's passed.
		// However, the current CreatePostHandler doesn't take editorRepo. It takes svc (NewsletterServiceInterface).
		// The service method CreatePost takes editorFirebaseUID.
		// Let's ensure the handlers are called with what they expect.
		// CreatePostHandler(svc service.NewsletterServiceInterface, editorRepo repository.EditorRepository)
		// ListPostsByNewsletterHandler(svc service.NewsletterServiceInterface)
		// The editorRepo is not directly used by ListPostsByNewsletterHandler.
		// CreatePostHandler uses editorRepo to get editor's DB ID from Firebase UID.
		// This is a bit inconsistent. Ideally, auth middleware would provide editor's DB ID.
		// For now, CreatePostHandler needs editorRepo, ListPostsByNewsletterHandler does not.
		// The service methods themselves handle ownership checks using the firebaseUID.

		// The CreatePostHandler was defined to take editorRepo, but it's not used in its body.
		// The service method `CreatePost` takes `editorFirebaseUID`.
		// Let's adjust the handler signature if editorRepo is not truly needed by the handler itself.
		// Re-checking CreatePostHandler: it does NOT use editorRepo. It uses auth.VerifyFirebaseJWT.
		// The service layer uses editorRepo. So, no need to pass editorRepo to CreatePostHandler.
		// ListPostsByNewsletterHandler also does not need editorRepo.

		switch r.Method {
		case http.MethodPost:
			postHandler.CreatePostHandler(newsletterService, editorRepo)(w, r) // editorRepo is passed to CreatePostHandler as per its signature
		case http.MethodGet:
			postHandler.ListPostsByNewsletterHandler(newsletterService)(w, r)
		default:
			http.Error(w, "Method not allowed for /api/newsletters/{newsletterID}/posts", http.StatusMethodNotAllowed)
		}
	})

	// Get, Update, Delete a specific post by its ID
	mux.HandleFunc("/api/posts/{postID}", func(w http.ResponseWriter, r *http.Request) {
		// Similar to above, UpdatePostHandler and DeletePostHandler will use editorFirebaseUID from JWT.
		// GetPostByIDHandler also uses JWT for auth.
		// None of these post handlers directly need editorRepo if the service methods handle ownership via firebaseUID.
		// Re-checking signatures:
		// GetPostByIDHandler(svc service.NewsletterServiceInterface)
		// UpdatePostHandler(svc service.NewsletterServiceInterface)
		// DeletePostHandler(svc service.NewsletterServiceInterface)
		// So, editorRepo is not passed to these specific post handlers.
		switch r.Method {
		case http.MethodGet:
			postHandler.GetPostByIDHandler(newsletterService)(w, r)
		case http.MethodPut:
			postHandler.UpdatePostHandler(newsletterService)(w, r)
		case http.MethodDelete:
			postHandler.DeletePostHandler(newsletterService)(w, r)
		default:
			http.Error(w, "Method not allowed for /api/posts/{postID}", http.StatusMethodNotAllowed)
		}
	})

	// Subscriber routes
	// POST /api/newsletters/{id}/subscribe
	mux.HandleFunc("/api/newsletters/{newsletterID}/subscribe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			subscriberHandler.SubscribeHandler(subscriberService)(w, r)
		} else {
			http.Error(w, "Method not allowed for /api/newsletters/{newsletterID}/subscribe", http.StatusMethodNotAllowed)
		}
	})

	// GET /api/subscriptions/unsubscribe?token={token}
	mux.HandleFunc("/api/subscriptions/unsubscribe", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			subscriberHandler.UnsubscribeHandler(subscriberService)(w, r)
		} else {
			http.Error(w, "Method not allowed for /api/subscriptions/unsubscribe", http.StatusMethodNotAllowed)
		}
	})
	
	// POST /api/subscribers/confirm?token={token} - This is usually a GET, but current service expects POST-like ConfirmSubscriptionRequest
	// For consistency with how ConfirmSubscription is implemented (taking a request body implicitly via JSON unmarshal),
	// we might keep it as POST or refactor ConfirmSubscription to take token from query param.
	// The service method `ConfirmSubscription` takes `ConfirmSubscriptionRequest` which has a `Token` field.
	// A GET request would typically pass this in the query string.
	// Let's assume the handler will manage extracting it from query for a GET.
	// Or, if the client is expected to send a POST with JSON body `{"token": "value"}`, then POST is fine.
	// The current service method `SubscribeToNewsletter` generates a link like:
	// "http://localhost:8080/api/subscribers/confirm?token=" + confirmationToken
	// This implies a GET request where the token is a query parameter.
	mux.HandleFunc("/api/subscribers/confirm", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet { // Changed to GET to match typical confirmation link pattern
			subscriberHandler.ConfirmSubscriptionHandler(subscriberService)(w, r)
		} else {
			http.Error(w, "Method not allowed for /api/subscribers/confirm", http.StatusMethodNotAllowed)
		}
	})


	// TODO: Add API-SUB-003: /newsletters/{id}/subscribers (Protected GET for editors)
	// This will require auth middleware and integration with newsletterService/editorRepo for ownership check.
	mux.HandleFunc("/api/newsletters/{newsletterID}/subscribers", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// Pass newsletterRepo for ownership check and editorRepo for getting editor ID from JWT
			subscriberHandler.GetSubscribersHandler(subscriberService, newsletterRepo, editorRepo)(w, r)
		} else {
			http.Error(w, "Method not allowed for /api/newsletters/{newsletterID}/subscribers", http.StatusMethodNotAllowed)
		}
	})


	return mux
}
