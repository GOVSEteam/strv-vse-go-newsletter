package newsletter

import (
	"net/http"
	"strconv"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository" // Need this for the response struct
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
)

const (
	DefaultLimit  = 10
	MaxLimit      = 100
	DefaultOffset = 0
)

type PaginatedNewslettersResponse struct {
	Data       []repository.Newsletter `json:"data"`
	Total      int                     `json:"total"`
	Limit      int                     `json:"limit"`
	Offset     int                     `json:"offset"`
}

func ListHandler(svc service.NewsletterServiceInterface, editorRepo repository.EditorRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			commonHandler.JSONError(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		firebaseUID, err := auth.VerifyFirebaseJWT(r)
		if err != nil {
			commonHandler.JSONError(w, "Invalid or missing token", http.StatusUnauthorized)
			return
		}

		editor, err := editorRepo.GetEditorByFirebaseUID(firebaseUID)
		if err != nil {
			commonHandler.JSONError(w, "Editor not found or not authorized", http.StatusForbidden)
			return
		}

		// Parse query parameters for pagination
		limitStr := r.URL.Query().Get("limit")
		limit := DefaultLimit
		if limitStr != "" {
			limit, err = strconv.Atoi(limitStr)
			if err != nil || limit <= 0 {
				commonHandler.JSONError(w, "Invalid limit parameter", http.StatusBadRequest)
				return
			}
		}
		if limit > MaxLimit {
			limit = MaxLimit
		}

		offsetStr := r.URL.Query().Get("offset")
		offset := DefaultOffset
		if offsetStr != "" {
			offset, err = strconv.Atoi(offsetStr)
			if err != nil || offset < 0 {
				commonHandler.JSONError(w, "Invalid offset parameter", http.StatusBadRequest)
				return
			}
		}

		newsletters, total, err := svc.ListNewslettersByEditorID(editor.ID, limit, offset)
		if err != nil {
			commonHandler.JSONError(w, "Failed to list newsletters: "+err.Error(), http.StatusInternalServerError)
			return
		}

		response := PaginatedNewslettersResponse{
			Data:   newsletters,
			Total:  total,
			Limit:  limit,
			Offset: offset,
		}
		commonHandler.JSONResponse(w, response, http.StatusOK)
	}
} 