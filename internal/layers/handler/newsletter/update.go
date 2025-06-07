package newsletter

import (
	"net/http"

	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"
	"github.com/go-chi/chi/v5"
)

// UpdateNewsletterRequest defines the expected request body for updating a newsletter.
// Using pointers to distinguish between a field not provided and a field provided with an empty value.
type UpdateNewsletterRequest struct {
	Name        *string `json:"name" validate:"omitempty,min=2,max=100"`
	Description *string `json:"description" validate:"omitempty,max=500"`
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
		if !commonHandler.ValidateAndRespond(w, r, &req) {
			return // Validation failed, response already sent
		}

		// Ensure at least one field is provided for update
		if req.Name == nil && req.Description == nil {
			commonHandler.JSONError(w, "At least one field (name or description) must be provided for update", http.StatusBadRequest)
			return
		}

		// The service UpdateNewsletter expects editorAuthID (e.g. FirebaseUID), newsletterID, and pointers for name/description.
		updatedNewsletter, err := svc.UpdateNewsletter(r.Context(), editorAuthID, newsletterID, req.Name, req.Description)
		if err != nil {
			commonHandler.JSONErrorSecure(w, err, "newsletter update")
			return
		}

		commonHandler.JSONResponse(w, updatedNewsletter, http.StatusOK)
	}
}
