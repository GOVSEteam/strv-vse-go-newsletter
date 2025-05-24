package subscriber

import (
	"encoding/json"
	"net/http"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler" // For Error and Success responses
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/utils" // For GetIDFromPath
)

// SubscribeHandler handles requests to subscribe to a newsletter.
// POST /api/newsletters/{newsletterID}/subscribe
func SubscribeHandler(subscriberService service.SubscriberServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newsletterID, err := utils.GetIDFromPath(r.URL.Path, "/api/newsletters/", "/subscribe")
		if err != nil {
			handler.JSONError(w, err.Error(), http.StatusBadRequest)
			return
		}

		var req service.SubscribeToNewsletterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			handler.JSONError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		req.NewsletterID = newsletterID // Set newsletterID from path

		if err := utils.ValidateEmail(req.Email); err != nil {
			handler.JSONError(w, err.Error(), http.StatusBadRequest)
			return
		}
		
		resp, err := subscriberService.SubscribeToNewsletter(r.Context(), req)
		if err != nil {
			// Specific error handling based on service errors
			if err == service.ErrAlreadySubscribed {
				handler.JSONError(w, err.Error(), http.StatusConflict)
			} else if err == service.ErrNewsletterNotFound {
				handler.JSONError(w, err.Error(), http.StatusNotFound)
			} else {
				handler.JSONError(w, "Failed to subscribe: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		handler.JSONResponse(w, resp, http.StatusCreated)
	}
}
