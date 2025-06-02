package newsletter

import (
	"encoding/json"
	"net/http"

	// "errors" // No longer needed for direct comparison if using apperrors mapping

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware" // Assuming GetEditorIDFromContext is here
	// For the response model
	// For the response model
)

type CreateNewsletterRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

// CreateHandler handles the creation of new newsletters.
// It relies on AuthMiddleware to have placed the editor's ID in the context.
func CreateHandler(svc service.NewsletterServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		editorID := middleware.GetEditorIDFromContext(r.Context())
		if editorID == "" {
			// This case should ideally be prevented by the AuthMiddleware.
			// If reached, it implies a middleware configuration issue or bypass.
			commonHandler.JSONError(w, "Unauthorized: editor ID not found in context", http.StatusUnauthorized)
			return
		}

		var req CreateNewsletterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			commonHandler.JSONError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Basic request validation can remain here, or be fully in service.
		// Service layer handles more comprehensive validation (e.g. length, specific formats)
		if req.Name == "" {
			// Example of direct validation error before calling service.
			// Alternatively, let the service validate and return apperrors.ErrValidation.
			err := apperrors.ErrNameEmpty // Use a predefined validation error
			commonHandler.JSONError(w, err.Error(), apperrors.ErrorToHTTPStatus(err))
			return
		}

		// The service CreateNewsletter expects editor's database ID.
		newsletter, err := svc.CreateNewsletter(r.Context(), editorID, req.Name, req.Description)
		if err != nil {
			statusCode := apperrors.ErrorToHTTPStatus(err)
			// It's good practice to log the original error on the server, especially for non-client errors.
			// log.Printf("Error creating newsletter: %v", err) // Example logging
			commonHandler.JSONError(w, err.Error(), statusCode) // err.Error() should provide a user-friendly message from the service/apperror
			return
		}

		// Ensure the response is models.Newsletter, which the service now returns.
		commonHandler.JSONResponse(w, newsletter, http.StatusCreated)
	}
}
