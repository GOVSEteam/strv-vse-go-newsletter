package newsletter

import (
	"encoding/json"
	"net/http"

	"errors"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler" // Alias for common handler package
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
)

type CreateNewsletterRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func CreateHandler(svc service.NewsletterServiceInterface, editorRepo repository.EditorRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			commonHandler.JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		firebaseUID, err := auth.VerifyFirebaseJWT(r)
		if err != nil {
			commonHandler.JSONError(w, "Invalid or missing token", http.StatusUnauthorized)
			return
		}

		editor, err := editorRepo.GetEditorByFirebaseUID(firebaseUID)
		if err != nil {
			// Consider if this should be 403 Forbidden or 404 Not Found depending on GetEditorByFirebaseUID behavior
			commonHandler.JSONError(w, "Editor not found or not authorized", http.StatusForbidden)
			return
		}

		var req CreateNewsletterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			commonHandler.JSONError(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Basic validation (more can be added in service layer if needed)
		if req.Name == "" {
			commonHandler.JSONError(w, "Newsletter name cannot be empty", http.StatusBadRequest)
			return
		}

		newsletter, err := svc.CreateNewsletter(r.Context(), editor.ID, req.Name, req.Description)
		if err != nil {
			if errors.Is(err, service.ErrNewsletterNameTaken) {
				commonHandler.JSONError(w, service.ErrNewsletterNameTaken.Error(), http.StatusConflict)
			} else {
				// TODO: Differentiate further errors from service (e.g., other validation errors vs. internal errors)
				commonHandler.JSONError(w, "Failed to create newsletter: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		commonHandler.JSONResponse(w, newsletter, http.StatusCreated)
	}
}
