package subscriber

import (
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
)

const (
	DefaultSubscriberLimit  = 10
	MaxSubscriberLimit      = 100
	DefaultSubscriberOffset = 0
)

// PaginatedSubscribersResponse defines the structure for paginated subscriber lists.
type PaginatedSubscribersResponse struct {
	Data   []models.Subscriber `json:"data"`
	Total  int                 `json:"total"`
	Limit  int                 `json:"limit"`
	Offset int                 `json:"offset"`
}

// ListSubscribersHandler handles requests for an editor to list active subscribers of their newsletter.
// GET /api/newsletters/{newsletterID}/subscribers
// Protected endpoint: Requires editor authentication.
func ListSubscribersHandler(subscriberService service.SubscriberServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		editorID := middleware.GetEditorIDFromContext(ctx)
		if editorID == "" {
			commonHandler.JSONError(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		newsletterIDStr := chi.URLParam(r, "newsletterID")
		if newsletterIDStr == "" {
			commonHandler.JSONError(w, "Newsletter ID is required in path", http.StatusBadRequest)
			return
		}

		// Parse pagination parameters
		limitStr := r.URL.Query().Get("limit")
		limit := DefaultSubscriberLimit
		if limitStr != "" {
			parsedLimit, err := strconv.Atoi(limitStr)
			if err != nil || parsedLimit <= 0 {
				commonHandler.JSONError(w, "Invalid limit parameter", http.StatusBadRequest)
				return
			}
			limit = parsedLimit
		}
		if limit > MaxSubscriberLimit {
			limit = MaxSubscriberLimit
		}

		offsetStr := r.URL.Query().Get("offset")
		offset := DefaultSubscriberOffset
		if offsetStr != "" {
			parsedOffset, err := strconv.Atoi(offsetStr)
			if err != nil || parsedOffset < 0 {
				commonHandler.JSONError(w, "Invalid offset parameter", http.StatusBadRequest)
				return
			}
			offset = parsedOffset
		}

		subscribers, total, err := subscriberService.ListActiveSubscribersByNewsletterID(ctx, editorID, newsletterIDStr, limit, offset)
		if err != nil {
			commonHandler.JSONErrorSecure(w, err, "subscriber list")
			return
		}

		response := PaginatedSubscribersResponse{
			Data:   subscribers,
			Total:  total,
			Limit:  limit,
			Offset: offset,
		}
		commonHandler.JSONResponse(w, response, http.StatusOK)
	}
}
