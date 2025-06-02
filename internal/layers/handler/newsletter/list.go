package newsletter

import (
	"net/http"
	"strconv"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models" // For models.Newsletter in response
)

const (
	DefaultLimit  = 10
	MaxLimit      = 100
	DefaultOffset = 0
)

// PaginatedNewslettersResponse defines the structure for paginated newsletter lists.
// It now uses models.Newsletter.
type PaginatedNewslettersResponse struct {
	Data   []models.Newsletter `json:"data"`
	Total  int                 `json:"total"`
	Limit  int                 `json:"limit"`
	Offset int                 `json:"offset"`
}

// ListHandler handles requests to list newsletters for the authenticated editor.
// It relies on AuthMiddleware to provide the editor's ID.
func ListHandler(svc service.NewsletterServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		editorID := middleware.GetEditorIDFromContext(r.Context())
		if editorID == "" {
			// This case should ideally be prevented by the AuthMiddleware.
			commonHandler.JSONError(w, "Unauthorized: editor ID not found in context", http.StatusUnauthorized)
			return
		}

		// Parse query parameters for pagination
		var err error // Declare err here for use in parsing
		limitStr := r.URL.Query().Get("limit")
		limit := DefaultLimit
		if limitStr != "" {
			limit, err = strconv.Atoi(limitStr)
			if err != nil || limit <= 0 {
				commonHandler.JSONError(w, "Invalid limit parameter", http.StatusBadRequest)
				return
			}
		}
		if limit > MaxLimit {
			limit = MaxLimit
		}

		offsetStr := r.URL.Query().Get("offset")
		offset := DefaultOffset
		if offsetStr != "" {
			offset, err = strconv.Atoi(offsetStr)
			if err != nil || offset < 0 {
				commonHandler.JSONError(w, "Invalid offset parameter", http.StatusBadRequest)
				return
			}
		}

		// Service method ListNewslettersByEditorID expects the editor's database ID.
		newsletters, total, err := svc.ListNewslettersByEditorID(r.Context(), editorID, limit, offset)
		if err != nil {
			statusCode := apperrors.ErrorToHTTPStatus(err)
			// log.Printf("Error listing newsletters: %v", err) // Example logging
			commonHandler.JSONError(w, err.Error(), statusCode)
			return
		}

		response := PaginatedNewslettersResponse{
			Data:   newsletters, // This is now []models.Newsletter
			Total:  total,
			Limit:  limit,
			Offset: offset,
		}
		commonHandler.JSONResponse(w, response, http.StatusOK)
	}
}
