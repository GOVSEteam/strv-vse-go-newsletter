package newsletter

import (
	"net/http"

	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"
)

type CreateNewsletterRequest struct {
	Name        string `json:"name" validate:"required,min=2,max=100"`
	Description string `json:"description" validate:"max=500"`
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
		if !commonHandler.ValidateAndRespond(w, r, &req) {
			return // Validation failed, response already sent
		}

		// The service CreateNewsletter expects editor's database ID.
		newsletter, err := svc.CreateNewsletter(r.Context(), editorID, req.Name, req.Description)
		if err != nil {
			commonHandler.JSONErrorSecure(w, err, "newsletter creation")
			return
		}

		// Ensure the response is models.Newsletter, which the service now returns.
		commonHandler.JSONResponse(w, newsletter, http.StatusCreated)
	}
}
