package post_handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"
)

func DeletePostHandler(svc service.NewsletterServiceInterface) http.HandlerFunc {
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

		err := svc.DeletePost(r.Context(), editorID, postIDStr)
		if err != nil {
			commonHandler.JSONErrorSecure(w, err, "post delete")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
