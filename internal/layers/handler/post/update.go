package post_handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"
)

// UpdatePostRequest defines the expected request body for updating a post.
// Using pointers to distinguish between a field not provided and a field provided with an empty value.
type UpdatePostRequest struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
}

func UpdatePostHandler(svc service.NewsletterServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		editorID := middleware.GetEditorIDFromContext(r.Context())
		if editorID == "" {
			commonHandler.JSONError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		postIDStr := chi.URLParam(r, "postID")
		if postIDStr == "" {
			commonHandler.JSONError(w, "Post ID is required in path", http.StatusBadRequest)
			return
		}

		var req UpdatePostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			commonHandler.JSONError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Basic validation: if a field is provided, it must not be empty.
		// The service layer will handle more detailed validation (e.g., length constraints).
		if req.Title != nil && *req.Title == "" {
			commonHandler.JSONError(w, "Post title, if provided, cannot be empty", http.StatusBadRequest)
			return
		}
		if req.Content != nil && *req.Content == "" {
			commonHandler.JSONError(w, "Post content, if provided, cannot be empty", http.StatusBadRequest)
			return
		}
		// Ensure at least one field is provided for update.
		if req.Title == nil && req.Content == nil {
			commonHandler.JSONError(w, "At least one field (title or content) must be provided for update", http.StatusBadRequest)
			return
		}

		// The UpdatePost service method expects editorID, postID, and pointers for title and content.
		// It also expects a *time.Time for publishedAt, which is nil here as this handler is for content updates.
		updatedPost, err := svc.UpdatePost(r.Context(), editorID, postIDStr, req.Title, req.Content, nil)
		if err != nil {
			statusCode := apperrors.ErrorToHTTPStatus(err)
			commonHandler.JSONError(w, err.Error(), statusCode)
			return
		}

		commonHandler.JSONResponse(w, updatedPost, http.StatusOK)
	}
}
