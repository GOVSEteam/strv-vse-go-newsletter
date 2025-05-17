package http

import (
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/transport/http/handlers"
	"net/http"
)

func NewRouter() http.Handler {
	mux := http.NewServeMux()

	newsletterRepo := repository.NewInMemoryNewsletterRepo()
	newsletterService := service.NewNewsletterService(newsletterRepo)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/api/newsletters", func(w http.ResponseWriter, r *http.Request) {
		handlers.NewslettersHandler(w, r, newsletterService)
	})

	return mux
}
