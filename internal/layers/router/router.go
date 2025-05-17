package router

import (
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/db"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"net/http"
)

func Router() http.Handler {
	mux := http.NewServeMux()

	db := db.ConnectDB()

	newsletterRepo := repository.NewsletterRepo(db)
	newsletterService := service.NewsletterService(newsletterRepo)

	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	mux.HandleFunc("/api/newsletters", func(w http.ResponseWriter, r *http.Request) {
		handler.NewslettersHandler(w, r, newsletterService)
	})

	return mux
}
