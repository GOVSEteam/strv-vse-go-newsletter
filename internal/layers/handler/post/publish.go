package post_handler

import (
	"errors"
	"net/http"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository" // For editorRepo
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/utils"
)

// PublishPostHandler handles requests to publish a post.
// POST /api/posts/{postID}/publish
func PublishPostHandler(
	publishingService service.PublishingServiceInterface,
	editorRepo repository.EditorRepository, // Needed to get editor's DB ID from Firebase UID
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		postID, err := utils.GetIDFromPath(r.URL.Path, "/api/posts/", "/publish")
		if err != nil {
			handler.JSONError(w, "Invalid post ID in path: "+err.Error(), http.StatusBadRequest)
			return
		}

		firebaseUID, err := auth.VerifyFirebaseJWT(r)
		if err != nil {
			handler.JSONError(w, "Authentication failed: "+err.Error(), http.StatusUnauthorized)
			return
		}

		// The publishingService.PublishPostToSubscribers will handle ownership checks internally.
		// It needs the editorFirebaseUID for that.
		err = publishingService.PublishPostToSubscribers(ctx, postID, firebaseUID)
		if err != nil {
			// Handle specific errors from publishing service if needed
			// e.g., if err == service.ErrPostNotFound, service.ErrForbidden, "post already published"
			// For now, a general error.
			// The error message from PublishingService might already be quite descriptive.
			if errors.Is(err, service.ErrPostNotFound) { // Assuming PublishingService might wrap this
				handler.JSONError(w, "Post not found: "+err.Error(), http.StatusNotFound)
			} else if errors.Is(err, service.ErrForbidden) { // Assuming PublishingService might wrap this
				handler.JSONError(w, "Forbidden: "+err.Error(), http.StatusForbidden)
			} else if err.Error() == "post is already published" { // Check for specific string if service returns it this way
				handler.JSONError(w, err.Error(), http.StatusConflict)
			} else {
				handler.JSONError(w, "Failed to publish post: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		handler.JSONResponse(w, map[string]string{"message": "Post published successfully and is being sent to subscribers."}, http.StatusOK)
	}
}
