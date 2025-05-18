package newsletter

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
)

// UpdateNewsletterRequest defines the expected request body for updating a newsletter.
// Using pointers to distinguish between a field not provided and a field provided with an empty value.
type UpdateNewsletterRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

func UpdateHandler(svc service.NewsletterServiceInterface, editorRepo repository.EditorRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
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

		var req UpdateNewsletterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			commonHandler.JSONError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Input validation
		if req.Name != nil && *req.Name == "" {
			commonHandler.JSONError(w, "Newsletter name, if provided, cannot be empty", http.StatusBadRequest)
			return
		}
		// Ensure at least one field is provided for update if that's a requirement.
		// For now, an empty PATCH request will just bump updated_at.
		if req.Name == nil && req.Description == nil {
			// commonHandler.JSONError(w, "At least one field (name or description) must be provided for update", http.StatusBadRequest)
			// return
			// Or, allow this to simply touch updated_at. The RFC implies rename/description update.
			// Let's proceed, it will just update `updated_at` if both are nil.
		}

		updatedNewsletter, err := svc.UpdateNewsletter(r.Context(), newsletterID, editor.ID, req.Name, req.Description)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				commonHandler.JSONError(w, "Newsletter not found or you don't have permission to update it", http.StatusNotFound) // Or StatusForbidden
			} else if errors.Is(err, service.ErrNewsletterNameTaken) {
				commonHandler.JSONError(w, service.ErrNewsletterNameTaken.Error(), http.StatusConflict)
			} else {
				// Log the full error for server-side debugging
				// log.Printf("Error updating newsletter %s: %v", newsletterID, err)
				commonHandler.JSONError(w, "Failed to update newsletter: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		commonHandler.JSONResponse(w, updatedNewsletter, http.StatusOK)
	}
}
