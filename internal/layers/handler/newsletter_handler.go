package handler

import (
	"encoding/json"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"net/http"
)

func NewslettersHandler(w http.ResponseWriter, r *http.Request, svc service.NewsletterService) {
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
