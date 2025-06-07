package post_handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"
)

type CreatePostRequest struct {
	Title   string `json:"title" validate:"required,min=3,max=150"`
	Content string `json:"content" validate:"required,min=10"`
}

func CreatePostHandler(svc service.NewsletterServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newsletterIDStr := chi.URLParam(r, "newsletterID")
		if newsletterIDStr == "" {
			commonHandler.JSONError(w, "Newsletter ID is required in path", http.StatusBadRequest)
			return
		}

		editorID := middleware.GetEditorIDFromContext(r.Context())
		if editorID == "" {
			commonHandler.JSONError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req CreatePostRequest
		if !commonHandler.ValidateAndRespond(w, r, &req) {
			return // Validation failed, response already sent
		}

		post, err := svc.CreatePost(r.Context(), editorID, newsletterIDStr, req.Title, req.Content)
		if err != nil {
			commonHandler.JSONErrorSecure(w, err, "post creation")
			return
		}

		commonHandler.JSONResponse(w, post, http.StatusCreated)
	}
}
