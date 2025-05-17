package http

import (
	"encoding/json"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/service"
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
		newslettersHandler(w, r, newsletterService)
	})

	return mux
}

func newslettersHandler(w http.ResponseWriter, r *http.Request, svc service.NewsletterService) {
	switch r.Method {
	case http.MethodGet:
		list, err := svc.ListNewsletters()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(list)
	case http.MethodPost:
		var req struct{ Name string }
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := svc.CreateNewsletter(req.Name); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusCreated)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
