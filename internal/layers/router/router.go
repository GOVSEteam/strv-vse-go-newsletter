package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.uber.org/zap"

	editorHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/editor"
	newsletterHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/newsletter"
	postHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/post"
	subscriberHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/subscriber"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"

	"github.com/jackc/pgx/v5/pgxpool"
)

// RouterDependencies holds only essential dependencies for the newsletter service.
type RouterDependencies struct {
	DB                *pgxpool.Pool
	AuthClient        middleware.AuthClient
	NewsletterService service.NewsletterServiceInterface
	SubscriberService service.SubscriberServiceInterface
	PublishingService service.PublishingServiceInterface
	EditorService     service.EditorServiceInterface
	PasswordResetSvc  service.PasswordResetService
	EditorRepo        repository.EditorRepository
	Logger            *zap.SugaredLogger
}

// NewRouter creates a simple Chi router for the newsletter service.
func NewRouter(deps RouterDependencies) *chi.Mux {
	r := chi.NewRouter()

	// Essential middleware only
	r.Use(middleware.LoggingMiddleware(deps.Logger))
	r.Use(middleware.RecoveryMiddleware(deps.Logger))
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000", "https://yourdomain.com"}, // Specific origins
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Simple health check
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := deps.DB.Ping(r.Context()); err != nil {
			http.Error(w, "Database unavailable", http.StatusServiceUnavailable)
			return
		}
		w.Write([]byte("OK"))
	})

	// Serve OpenAPI specification
	r.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "docs/openapi.yaml")
	})

	// API routes
	r.Route("/api", func(r chi.Router) {
		// Public routes
		r.Post("/editor/signup", editorHandler.EditorSignUpHandler(deps.EditorService))
		r.Post("/editor/signin", editorHandler.EditorSignInHandler(deps.EditorService))
		r.Post("/editor/password-reset", editorHandler.PasswordResetRequestHandler(deps.PasswordResetSvc))
		r.Post("/newsletters/{newsletterID}/subscribe", subscriberHandler.SubscribeHandler(deps.SubscriberService))
		r.Get("/subscriptions/unsubscribe", subscriberHandler.UnsubscribeHandler(deps.SubscriberService))

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(deps.AuthClient, deps.EditorRepo))

			// Newsletter management
			r.Route("/newsletters", func(r chi.Router) {
				r.Get("/", newsletterHandler.ListHandler(deps.NewsletterService))
				r.Post("/", newsletterHandler.CreateHandler(deps.NewsletterService))
				r.Get("/{newsletterID}", newsletterHandler.GetByIDHandler(deps.NewsletterService))
				r.Patch("/{newsletterID}", newsletterHandler.UpdateHandler(deps.NewsletterService))
				r.Delete("/{newsletterID}", newsletterHandler.DeleteHandler(deps.NewsletterService))
				r.Get("/{newsletterID}/subscribers", subscriberHandler.ListSubscribersHandler(deps.SubscriberService))

				// Posts
				r.Post("/{newsletterID}/posts", postHandler.CreatePostHandler(deps.NewsletterService))
				r.Get("/{newsletterID}/posts", postHandler.ListPostsByNewsletterHandler(deps.NewsletterService))
			})

			// Individual post operations
			r.Route("/posts/{postID}", func(r chi.Router) {
				r.Get("/", postHandler.GetPostByIDHandler(deps.NewsletterService))
				r.Put("/", postHandler.UpdatePostHandler(deps.NewsletterService))
				r.Delete("/", postHandler.DeletePostHandler(deps.NewsletterService))
				r.Post("/publish", postHandler.PublishPostHandler(deps.PublishingService))
			})
		})
	})

	return r
}
