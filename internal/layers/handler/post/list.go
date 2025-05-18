package post_handler

import (
	"net/http"
	"strconv"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	globalHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models" // For models.Post
	"github.com/google/uuid"
)

const (
	DefaultPostLimit  = 10
	MaxPostLimit      = 100
	DefaultPostOffset = 0
)

type PaginatedPostsResponse struct {
	Data   []*models.Post `json:"data"`
	Total  int            `json:"total"`
	Limit  int            `json:"limit"`
	Offset int            `json:"offset"`
}

func ListPostsByNewsletterHandler(svc service.NewsletterServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			globalHandler.JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// TODO: Replace with auth middleware (if listing posts is protected)
		// For now, let's assume it's protected as per API-POST-001
		_, err := auth.VerifyFirebaseJWT(r)
		if err != nil {
			globalHandler.JSONError(w, "Invalid or missing token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		newsletterIDStr := r.PathValue("newsletterID")
		if newsletterIDStr == "" {
			globalHandler.JSONError(w, "Newsletter ID is required in path", http.StatusBadRequest)
			return
		}
		newsletterID, err := uuid.Parse(newsletterIDStr)
		if err != nil {
			globalHandler.JSONError(w, "Invalid newsletter ID format", http.StatusBadRequest)
			return
		}

		limitStr := r.URL.Query().Get("limit")
		limit := DefaultPostLimit
		if limitStr != "" {
			parsedLimit, err := strconv.Atoi(limitStr)
			if err != nil || parsedLimit <= 0 {
				globalHandler.JSONError(w, "Invalid limit parameter", http.StatusBadRequest)
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
				globalHandler.JSONError(w, "Invalid offset parameter", http.StatusBadRequest)
				return
			}
			offset = parsedOffset
		}

		posts, total, err := svc.ListPostsByNewsletter(r.Context(), newsletterID, limit, offset)
		if err != nil {
			// Consider specific errors like service.ErrServiceNewsletterNotFound if the newsletter itself doesn't exist
			globalHandler.JSONError(w, "Failed to list posts: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := PaginatedPostsResponse{
			Data:   posts,
			Total:  total,
			Limit:  limit,
			Offset: offset,
		}
		globalHandler.JSONResponse(w, response, http.StatusOK)
	}
}
