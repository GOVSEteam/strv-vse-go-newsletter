package newsletter

import (
	// "database/sql" // No longer needed for direct error checking
	// "errors" // No longer needed for direct error checking
	"net/http"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"
)

// DeleteHandler handles the deletion of a newsletter.
// It relies on AuthMiddleware for authentication and extracting the editor's ID.
func DeleteHandler(svc service.NewsletterServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			commonHandler.JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		newsletterID := r.PathValue("newsletterID") // Requires Go 1.22+ and router pattern like /api/newsletters/{id}
		if newsletterID == "" {
			// This should ideally be caught by a route validation or a more specific check
			// if the router doesn't guarantee a non-empty {id}.
			commonHandler.JSONError(w, "Newsletter ID is required in path", http.StatusBadRequest)
			return
		}

		editorAuthID := middleware.GetEditorIDFromContext(r.Context())
		if editorAuthID == "" {
			// This case should ideally be prevented by the AuthMiddleware.
			commonHandler.JSONError(w, "Unauthorized: editor ID not found in context", http.StatusUnauthorized)
			return
		}

		// The service method DeleteNewsletter expects the editor's Firebase UID (or general auth ID)
		// and the newsletter ID.
		err := svc.DeleteNewsletter(r.Context(), editorAuthID, newsletterID)
		if err != nil {
			statusCode := apperrors.ErrorToHTTPStatus(err)
			// log.Printf("Error deleting newsletter %s: %v", newsletterID, err) // Example logging
			commonHandler.JSONError(w, err.Error(), statusCode)
			return
		}

		w.WriteHeader(http.StatusNoContent) // Success, no body
	}
}
