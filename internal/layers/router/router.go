package router

import (
	"net/http"

	"firebase.google.com/go/v4/auth"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
	"go.uber.org/zap"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/editor"
	newsletterHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/newsletter"
	postHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/post"
	subscriberHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/subscriber"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/pkg/email" // Added email package
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/setup"

	"github.com/go-chi/chi/v5"
	"github.com/swaggo/http-swagger"
)

func Router() http.Handler {
	r := chi.NewRouter()

// NewRouter creates a simple Chi router for the newsletter service.
func NewRouter(deps RouterDependencies) *chi.Mux {
	r := chi.NewRouter()

	editorRepo := repository.EditorRepo(db)
	editorService := service.NewEditorService(editorRepo)

	newsletterRepo := repository.NewsletterRepo(db)
	postRepo := repository.NewPostRepository(db)
	newsletterService := service.NewNewsletterService(newsletterRepo, postRepo, editorRepo)

	var emailSvc email.EmailService
	var err error
	if os.Getenv("GOOGLE_APP_PASSWORD") != "" {
		emailSvc, err = email.NewGmailService()
		if err != nil {
			log.Fatalf("Error initializing Gmail email service: %v", err)
		}
		log.Println("Using Gmail email service")
	} else {
		emailSvc = email.NewConsoleEmailService()
		log.Println("Using Console email service (GOOGLE_APP_PASSWORD not set)")
	}

	firestoreClient := setup.GetFirestoreClient()
	if firestoreClient == nil {
		log.Fatal("Failed to get Firestore client")
	}
	subscriberRepo := repository.NewFirestoreSubscriberRepository(firestoreClient)
	subscriberService := service.NewSubscriberService(subscriberRepo, newsletterRepo, emailSvc)
	publishingService := service.NewPublishingService(newsletterService, subscriberService, emailSvc)

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	r.Get("/swagger/*", func(w http.ResponseWriter, r *http.Request) {
		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := r.Host
		specURL := fmt.Sprintf("%s://%s/docs/openapi.yaml", scheme, host)
		httpSwagger.Handler(httpSwagger.URL(specURL))(w, r)
	})

	r.Get("/docs/*", func(w http.ResponseWriter, r *http.Request) {
		scheme := "http"
		if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
			scheme = "https"
		}
		host := r.Host
		specURL := fmt.Sprintf("%s://%s/docs/openapi.yaml", scheme, host)
		httpSwagger.Handler(httpSwagger.URL(specURL))(w, r)
	})

	r.Get("/docs/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/x-yaml")
		http.ServeFile(w, r, "docs/openapi.yaml")
	})

	r.Route("/editor", func(r chi.Router) {
		r.Post("/signup", editor.EditorSignUpHandler(editorService))
		r.Post("/signin", editor.EditorSignInHandler(editorService))
		r.Post("/password-reset-request", editor.FirebasePasswordResetRequestHandler())
	})

	r.Route("/api/newsletters", func(r chi.Router) {
		r.Get("/", newsletterHandler.ListHandler(newsletterService, editorRepo))
		r.Post("/", newsletterHandler.CreateHandler(newsletterService, editorRepo))
		r.Route("/{newsletterID}", func(r chi.Router) {
			r.Patch("/", newsletterHandler.UpdateHandler(newsletterService, editorRepo))
			r.Delete("/", newsletterHandler.DeleteHandler(newsletterService, editorRepo))
			r.Get("/subscribers", subscriberHandler.GetSubscribersHandler(subscriberService, newsletterRepo, editorRepo))
			r.Post("/subscribe", subscriberHandler.SubscribeHandler(subscriberService))
			r.Route("/posts", func(r chi.Router) {
				r.Get("/", postHandler.ListPostsByNewsletterHandler(newsletterService))
				r.Post("/", postHandler.CreatePostHandler(newsletterService, editorRepo))
			})
		})
	})

	r.Route("/api/posts", func(r chi.Router) {
		r.Route("/{postID}", func(r chi.Router) {
			r.Get("/", postHandler.GetPostByIDHandler(newsletterService))
			r.Put("/", postHandler.UpdatePostHandler(newsletterService))
			r.Delete("/", postHandler.DeletePostHandler(newsletterService))
			r.Post("/publish", postHandler.PublishPostHandler(publishingService, editorRepo))
		})
	})

	r.Get("/api/subscriptions/unsubscribe", subscriberHandler.UnsubscribeHandler(subscriberService))

	return r
}
