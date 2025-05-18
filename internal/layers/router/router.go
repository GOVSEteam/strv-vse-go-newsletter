package router

import (
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/editor"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/setup"
	"net/http"
)

func Router() http.Handler {
	mux := http.NewServeMux()

	db := setup.ConnectDB()

	newsletterRepo := repository.NewsletterRepo(db)
	newsletterService := service.NewsletterService(newsletterRepo)

	editorRepo := repository.EditorRepo(db)
	editorService := service.NewEditorService(editorRepo)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/api/newsletters", func(w http.ResponseWriter, r *http.Request) {
		handler.NewslettersHandler(w, r, newsletterService, editorRepo)
	})

	mux.HandleFunc("/signup", editor.EditorSignUpHandler(editorService))
	mux.HandleFunc("/signin", editor.EditorSignInHandler(editorService))

	return mux
}
