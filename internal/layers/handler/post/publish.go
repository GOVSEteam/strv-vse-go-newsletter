package post_handler

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"
)

// PublishPostHandler handles requests to publish a post.
// POST /api/posts/{postID}/publish
func PublishPostHandler(publishingService service.PublishingServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		editorID := middleware.GetEditorIDFromContext(ctx)
		if editorID == "" {
			commonHandler.JSONError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		postIDStr := chi.URLParam(r, "postID")
		if postIDStr == "" {
			commonHandler.JSONError(w, "Post ID is required in path", http.StatusBadRequest)
			return
		}

		// The publishingService.PublishPostToSubscribers will handle ownership checks internally.
		// It needs the editorID (Firebase UID) for that.
		err := publishingService.PublishPostToSubscribers(ctx, postIDStr, editorID)
		if err != nil {
			if errors.Is(err, service.ErrPostAlreadyPublished) {
				commonHandler.JSONError(w, err.Error(), http.StatusConflict)
				return
			}

			// 500 errors
			commonHandler.JSONErrorSecure(w, err, "post publish")
			return
		}

		commonHandler.JSONResponse(w, map[string]string{"message": "Post published successfully and is being sent to subscribers."}, http.StatusOK)
	}
}
