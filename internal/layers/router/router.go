package router

import (
	"net/http"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/editor"
	newsletterHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/newsletter" // Import specific newsletter handlers
	postHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/post"             // Import post handlers
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
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
	return mux
}
