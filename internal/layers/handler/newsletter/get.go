package newsletter

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"
)

// GetByIDHandler handles requests to fetch a specific newsletter by its ID.
// It relies on AuthMiddleware to provide the editor's authentication ID.
func GetByIDHandler(svc service.NewsletterServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		editorAuthID := middleware.GetEditorIDFromContext(r.Context())
		if editorAuthID == "" {
			commonHandler.JSONError(w, "Unauthorized: editor ID not found in context", http.StatusUnauthorized)
			return
		}

		newsletterID := chi.URLParam(r, "newsletterID")
		if newsletterID == "" {
			commonHandler.JSONError(w, "Newsletter ID is required in path", http.StatusBadRequest)
			return
		}

		newsletter, err := svc.GetNewsletterForEditor(r.Context(), editorAuthID, newsletterID)
		if err != nil {
			statusCode := apperrors.ErrorToHTTPStatus(err)
			commonHandler.JSONError(w, err.Error(), statusCode)
			return
		}

		commonHandler.JSONResponse(w, newsletter, http.StatusOK)
	}
} 