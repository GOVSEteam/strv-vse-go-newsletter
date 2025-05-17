package handler

import (
	"encoding/json"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"net/http"
)

func NewslettersHandler(w http.ResponseWriter, r *http.Request, svc service.NewsletterServiceInterface, editorRepo repository.EditorRepository) {
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
		// Require JWT
		firebaseUID, err := auth.VerifyFirebaseJWT(r)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "invalid or missing token"})
			return
		}
		editor, err := editorRepo.GetEditorByFirebaseUID(firebaseUID)
		if err != nil {
			w.WriteHeader(http.StatusForbidden)
			json.NewEncoder(w).Encode(map[string]string{"error": "editor not found"})
			return
		}
		var req struct {
			Name        string `json:"name"`
			Description string `json:"description"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Name == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		newsletter, err := svc.CreateNewsletter(editor.ID, req.Name, req.Description)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(newsletter)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
