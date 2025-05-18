package newsletter

import (
	"database/sql"
	"errors"
	"net/http"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
)

func DeleteHandler(svc service.NewsletterServiceInterface, editorRepo repository.EditorRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			commonHandler.JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		newsletterID := r.PathValue("id") // Requires Go 1.22+ and router pattern like /api/newsletters/{id}
		if newsletterID == "" {
			commonHandler.JSONError(w, "Newsletter ID is required in path", http.StatusBadRequest)
			return
		}

		firebaseUID, err := auth.VerifyFirebaseJWT(r)
		if err != nil {
			commonHandler.JSONError(w, "Invalid or missing token", http.StatusUnauthorized)
			return
		}

		editor, err := editorRepo.GetEditorByFirebaseUID(firebaseUID)
		if err != nil {
			commonHandler.JSONError(w, "Editor not found or not authorized", http.StatusForbidden)
			return
		}

		err = svc.DeleteNewsletter(r.Context(), newsletterID, editor.ID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				commonHandler.JSONError(w, "Newsletter not found or you don't have permission to delete it", http.StatusNotFound) // Or StatusForbidden
			} else {
				// log.Printf("Error deleting newsletter %s: %v", newsletterID, err)
				commonHandler.JSONError(w, "Failed to delete newsletter: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		w.WriteHeader(http.StatusNoContent) // Success, no body
	}
}
