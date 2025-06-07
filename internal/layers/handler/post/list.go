package post_handler

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models" // For models.Post
)

const (
	DefaultPostLimit  = 10
	MaxPostLimit      = 100
	DefaultPostOffset = 0
)

type PaginatedPostsResponse struct {
	Data   []models.Post `json:"data"`
	Total  int           `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

func ListPostsByNewsletterHandler(svc service.NewsletterServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Ensure editor is authenticated, although not directly used by ListPostsByNewsletterID, it's good practice for grouped routes
		editorID := middleware.GetEditorIDFromContext(r.Context())
		if editorID == "" {
			commonHandler.JSONError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		newsletterIDStr := chi.URLParam(r, "newsletterID")
		if newsletterIDStr == "" {
			commonHandler.JSONError(w, "Newsletter ID is required in path", http.StatusBadRequest)
			return
		}

		// Validation of newsletterID format (UUID) is implicitly handled by service/repository layer if it attempts to parse/use it.
		// No need to parse to uuid.UUID here if the service layer expects a string and handles validation.

		limitStr := r.URL.Query().Get("limit")
		limit := DefaultPostLimit
		if limitStr != "" {
			parsedLimit, err := strconv.Atoi(limitStr)
			if err != nil || parsedLimit <= 0 {
				commonHandler.JSONError(w, "Invalid limit parameter", http.StatusBadRequest)
				return
			}
			limit = parsedLimit
		}
		if limit > MaxPostLimit {
			limit = MaxPostLimit
		}

		offsetStr := r.URL.Query().Get("offset")
		offset := DefaultPostOffset
		if offsetStr != "" {
			parsedOffset, err := strconv.Atoi(offsetStr)
			if err != nil || parsedOffset < 0 {
				commonHandler.JSONError(w, "Invalid offset parameter", http.StatusBadRequest)
				return
			}
			offset = parsedOffset
		}

		posts, total, err := svc.ListPostsByNewsletterID(r.Context(), newsletterIDStr, limit, offset)
		if err != nil {
			commonHandler.JSONErrorSecure(w, err, "post list")
			return
		}

		response := PaginatedPostsResponse{
			Data:   posts,
			Total:  total,
			Limit:  limit,
			Offset: offset,
		}
		commonHandler.JSONResponse(w, response, http.StatusOK)
	}
}
