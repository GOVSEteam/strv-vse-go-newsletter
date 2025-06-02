package post_handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"
)

// PublishPostHandler handles requests to publish a post.
// POST /api/posts/{postID}/publish
func PublishPostHandler(
	publishingService service.PublishingServiceInterface,
	// editorRepo is no longer needed as editorID comes from context
) http.HandlerFunc {
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
			statusCode := apperrors.ErrorToHTTPStatus(err)
			commonHandler.JSONError(w, err.Error(), statusCode)
			return
		}

		commonHandler.JSONResponse(w, map[string]string{"message": "Post published successfully and is being sent to subscribers."}, http.StatusOK)
	}
}
