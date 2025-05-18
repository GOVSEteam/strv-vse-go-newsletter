package post_handler

import (
	"errors"
	"net/http"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	globalHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/google/uuid"
)

func GetPostByIDHandler(svc service.NewsletterServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			globalHandler.JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// TODO: Replace with auth middleware
		_, err := auth.VerifyFirebaseJWT(r)
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

		post, err := svc.GetPostByID(r.Context(), postID)
		if err != nil {
			if errors.Is(err, service.ErrPostNotFound) {
				globalHandler.JSONError(w, "Post not found", http.StatusNotFound)
			} else {
				globalHandler.JSONError(w, "Failed to get post: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		globalHandler.JSONResponse(w, post, http.StatusOK)
	}
}
