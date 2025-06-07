package post_handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"
)

func GetPostByIDHandler(svc service.NewsletterServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// GetPostForEditor requires editorID, so we need to get it from context
		editorID := middleware.GetEditorIDFromContext(r.Context())
		if editorID == "" {
			// This implies that auth middleware did not run or failed to set editorID
			// which should ideally be caught by the middleware itself.
			// However, as a safeguard:
			commonHandler.JSONError(w, "Unauthorized: editor ID not available", http.StatusUnauthorized)
			return
		}

		postIDStr := chi.URLParam(r, "postID")
		if postIDStr == "" {
			commonHandler.JSONError(w, "Post ID is required in path", http.StatusBadRequest)
			return
		}

		// The service method GetPostForEditor should be used to ensure ownership.
		// It expects the editor's Auth ID (which is what GetEditorIDFromContext provides)
		// and the post ID.
		post, err := svc.GetPostForEditor(r.Context(), editorID, postIDStr)
		if err != nil {
			commonHandler.JSONErrorSecure(w, err, "post get")
			return
		}

		commonHandler.JSONResponse(w, post, http.StatusOK)
	}
}
