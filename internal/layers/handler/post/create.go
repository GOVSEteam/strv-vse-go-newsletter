package post_handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"
	"github.com/google/uuid"
)

type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func CreatePostHandler(svc service.NewsletterServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newsletterIDStr := chi.URLParam(r, "newsletterID")
		if newsletterIDStr == "" {
			commonHandler.JSONError(w, "Newsletter ID is required in path", http.StatusBadRequest)
			return
		}
		newsletterID, err := uuid.Parse(newsletterIDStr)
		if err != nil {
			commonHandler.JSONError(w, "Invalid newsletter ID format", http.StatusBadRequest)
			return
		}

		editorID := middleware.GetEditorIDFromContext(r.Context())
		if editorID == "" {
			commonHandler.JSONError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		var req CreatePostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			commonHandler.JSONError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		if req.Title == "" {
			commonHandler.JSONError(w, "Post title cannot be empty", http.StatusBadRequest)
			return
		}
		if req.Content == "" {
			commonHandler.JSONError(w, "Post content cannot be empty", http.StatusBadRequest)
			return
		}

		post, err := svc.CreatePost(r.Context(), editorID, newsletterID.String(), req.Title, req.Content)
		if err != nil {
			statusCode := apperrors.ErrorToHTTPStatus(err)
			commonHandler.JSONError(w, err.Error(), statusCode)
			return
		}

		commonHandler.JSONResponse(w, post, http.StatusCreated)
	}
}
