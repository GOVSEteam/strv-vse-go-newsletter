package newsletter

import (
	// "database/sql" // No longer directly checked
	"encoding/json"
	// "errors" // No longer directly checked
	"net/http"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"
	// models is implicitly used by the service return type, no direct handler use here for response struct
)

// UpdateNewsletterRequest defines the expected request body for updating a newsletter.
// Using pointers to distinguish between a field not provided and a field provided with an empty value.
type UpdateNewsletterRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
}

// UpdateHandler handles partial updates to a newsletter.
// It relies on AuthMiddleware for authentication and editor ID retrieval.
func UpdateHandler(svc service.NewsletterServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			commonHandler.JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		newsletterID := r.PathValue("newsletterID") // Requires Go 1.22+ and router pattern like /api/newsletters/{id}
		if newsletterID == "" {
			commonHandler.JSONError(w, "Newsletter ID is required in path", http.StatusBadRequest)
			return
		}

		editorAuthID := middleware.GetEditorIDFromContext(r.Context())
		if editorAuthID == "" {
			// This case should ideally be prevented by the AuthMiddleware.
			commonHandler.JSONError(w, "Unauthorized: editor ID not found in context", http.StatusUnauthorized)
			return
		}

		var req UpdateNewsletterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			commonHandler.JSONError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Input validation: if name is provided, it cannot be empty.
		// More comprehensive validation (e.g., length) is in the service layer.
		if req.Name != nil && *req.Name == "" {
			err := apperrors.ErrNameEmpty // Or a more specific validation error for update
			commonHandler.JSONError(w, err.Error(), apperrors.ErrorToHTTPStatus(err))
			return
		}
		// The service layer will handle the case where both req.Name and req.Description are nil
		// (e.g., by only updating timestamps or returning a validation error if no change is made).

		// The service UpdateNewsletter expects editorAuthID (e.g. FirebaseUID), newsletterID, and pointers for name/description.
		updatedNewsletter, err := svc.UpdateNewsletter(r.Context(), editorAuthID, newsletterID, req.Name, req.Description)
		if err != nil {
			statusCode := apperrors.ErrorToHTTPStatus(err)
			// log.Printf("Error updating newsletter %s: %v", newsletterID, err) // Example logging
			commonHandler.JSONError(w, err.Error(), statusCode)
			return
		}

		commonHandler.JSONResponse(w, updatedNewsletter, http.StatusOK)
	}
}
