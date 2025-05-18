package post_handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	globalHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/google/uuid"
)

// UpdatePostRequest defines the expected request body for updating a post.
// Using pointers to distinguish between a field not provided and a field provided with an empty value.
type UpdatePostRequest struct {
	Title   *string `json:"title"`
	Content *string `json:"content"`
}

func UpdatePostHandler(svc service.NewsletterServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// API-POST-001 specifies PUT, but PATCH is often preferred for partial updates.
		// Sticking to PUT as per spec for now.
		if r.Method != http.MethodPut {
			globalHandler.JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// TODO: Replace with auth middleware
		editorFirebaseUID, err := auth.VerifyFirebaseJWT(r)
		if err != nil {
			globalHandler.JSONError(w, "Invalid or missing token: "+err.Error(), http.StatusUnauthorized)
			return
		}

		postIDStr := r.PathValue("postID") // Assuming pattern like /posts/{postID}
		if postIDStr == "" {
			globalHandler.JSONError(w, "Post ID is required in path", http.StatusBadRequest)
			return
		}
		postID, err := uuid.Parse(postIDStr)
		if err != nil {
			globalHandler.JSONError(w, "Invalid post ID format", http.StatusBadRequest)
			return
		}

		var req UpdatePostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			globalHandler.JSONError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Basic validation
		if req.Title != nil && *req.Title == "" {
			globalHandler.JSONError(w, "Post title, if provided, cannot be empty", http.StatusBadRequest)
			return
		}
		if req.Content != nil && *req.Content == "" {
			globalHandler.JSONError(w, "Post content, if provided, cannot be empty", http.StatusBadRequest)
			return
		}
		if req.Title == nil && req.Content == nil {
			globalHandler.JSONError(w, "At least one field (title or content) must be provided for update", http.StatusBadRequest)
			return
		}

		updatedPost, err := svc.UpdatePost(r.Context(), editorFirebaseUID, postID, req.Title, req.Content)
		if err != nil {
			if errors.Is(err, service.ErrPostNotFound) {
				globalHandler.JSONError(w, "Post not found", http.StatusNotFound)
			} else if errors.Is(err, service.ErrForbidden) {
				globalHandler.JSONError(w, "Forbidden: You do not own this post or editor not found.", http.StatusForbidden)
			} else {
				globalHandler.JSONError(w, "Failed to update post: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		globalHandler.JSONResponse(w, updatedPost, http.StatusOK)
	}
}
