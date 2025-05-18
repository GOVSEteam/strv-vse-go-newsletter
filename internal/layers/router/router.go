package router

import (
	// "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler" // Old monolithic handler, no longer needed directly here
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/editor"
	newsletterHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/newsletter" // New specific newsletter handlers
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/setup"
	"net/http"
)

func Router() http.Handler {
	mux := http.NewServeMux()

	db := setup.ConnectDB()

	newsletterRepo := repository.NewsletterRepo(db)
	newsletterSvc := service.NewsletterService(newsletterRepo) // Renamed for clarity

	editorRepo := repository.EditorRepo(db)
	editorSvc := service.NewEditorService(editorRepo) // Renamed for clarity

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	// Newsletter routes
	mux.HandleFunc("POST /api/newsletters", newsletterHandler.CreateHandler(newsletterSvc, editorRepo))
	mux.HandleFunc("GET /api/newsletters", newsletterHandler.ListHandler(newsletterSvc, editorRepo))
	mux.HandleFunc("PATCH /api/newsletters/{id}", newsletterHandler.UpdateHandler(newsletterSvc, editorRepo))
	mux.HandleFunc("DELETE /api/newsletters/{id}", newsletterHandler.DeleteHandler(newsletterSvc, editorRepo))

	mux.HandleFunc("/signup", editor.EditorSignUpHandler(editorSvc))
	mux.HandleFunc("/signin", editor.EditorSignInHandler(editorSvc))

	return mux
}
