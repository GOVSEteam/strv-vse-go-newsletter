package post_handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"
)

// UpdatePostRequest defines the expected request body for updating a post.
// Using pointers to distinguish between a field not provided and a field provided with an empty value.
type UpdatePostRequest struct {
	Title   *string `json:"title" validate:"omitempty,min=3,max=150"`
	Content *string `json:"content" validate:"omitempty,min=10"`
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
		if !commonHandler.ValidateAndRespond(w, r, &req) {
			return // Validation failed, response already sent
		}

		// Ensure at least one field is provided for update
		if req.Title == nil && req.Content == nil {
			commonHandler.JSONError(w, "At least one field (title or content) must be provided for update", http.StatusBadRequest)
			return
		}

		// The UpdatePost service method expects editorID, postID, and pointers for title and content.
		updatedPost, err := svc.UpdatePost(r.Context(), editorID, postIDStr, req.Title, req.Content)
		if err != nil {
			commonHandler.JSONErrorSecure(w, err, "post update")
			return
		}

		commonHandler.JSONResponse(w, updatedPost, http.StatusOK)
	}
}
