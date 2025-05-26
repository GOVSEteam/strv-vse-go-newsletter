package post_handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	globalHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository" // For EditorRepository
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/google/uuid"
)

type CreatePostRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func CreatePostHandler(svc service.NewsletterServiceInterface, editorRepo repository.EditorRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			globalHandler.JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		newsletterIDStr := r.PathValue("newsletterID") // Assuming Go 1.22+ and pattern like /newsletters/{newsletterID}/posts
		if newsletterIDStr == "" {
			globalHandler.JSONError(w, "Newsletter ID is required in path", http.StatusBadRequest)
			return
		}
		newsletterID, err := uuid.Parse(newsletterIDStr)
		if err != nil {
			globalHandler.JSONError(w, "Invalid newsletter ID format", http.StatusBadRequest)
			return
		}

		// TODO: Replace with auth middleware
		editorFirebaseUID, err := auth.VerifyFirebaseJWT(r)
		if err != nil {
			globalHandler.JSONError(w, "Invalid or missing token: "+err.Error(), http.StatusUnauthorized)
			return
		}
		// The service layer will use editorFirebaseUID to check ownership via editorRepo
		// No need to call editorRepo here directly if service handles it.

		var req CreatePostRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			globalHandler.JSONError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}

		// Title and Content validation removed from handler, will be done by service
		// if req.Title == "" {
		// 	globalHandler.JSONError(w, "Post title cannot be empty", http.StatusBadRequest)
		// 	return
		// }
		// if req.Content == "" {
		// 	globalHandler.JSONError(w, "Post content cannot be empty", http.StatusBadRequest)
		// 	return
		// }

		post, err := svc.CreatePost(r.Context(), editorFirebaseUID, newsletterID, req.Title, req.Content)
		if err != nil {
			if errors.Is(err, service.ErrForbidden) {
				globalHandler.JSONError(w, "Forbidden: You do not own this newsletter or editor not found.", http.StatusForbidden)
			} else if errors.Is(err, service.ErrServiceNewsletterNotFound) { // Using the renamed error
				globalHandler.JSONError(w, "Newsletter not found.", http.StatusNotFound)
			} else if errors.Is(err, service.ErrPostTitleEmpty) || errors.Is(err, service.ErrPostContentEmpty) {
				globalHandler.JSONError(w, err.Error(), http.StatusBadRequest)
			} else {
				globalHandler.JSONError(w, "Failed to create post: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		globalHandler.JSONResponse(w, post, http.StatusCreated)
	}
}
