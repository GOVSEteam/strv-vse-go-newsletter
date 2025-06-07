package subscriber

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
)

// SubscribeRequest defines the expected JSON request body for subscribing.
// Note: NewsletterID is taken from the path, not the body.
type SubscribeRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// SubscribeResponse defines the JSON response for a successful subscription.
type SubscribeResponse struct {
	SubscriberID string                  `json:"subscriber_id"`
	Email        string                  `json:"email"`
	NewsletterID string                  `json:"newsletter_id"`
	Status       models.SubscriberStatus `json:"status"`
}

// SubscribeHandler handles requests to subscribe to a newsletter.
// POST /api/newsletters/{newsletterID}/subscribe
func SubscribeHandler(subscriberService service.SubscriberServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		newsletterIDStr := chi.URLParam(r, "newsletterID")
		if newsletterIDStr == "" {
			commonHandler.JSONError(w, "Newsletter ID is required in path", http.StatusBadRequest)
			return
		}

		var req SubscribeRequest
		if !commonHandler.ValidateAndRespond(w, r, &req) {
			return // Validation failed, response already sent
		}

		// Call service with Email from request body and NewsletterID from path
		subscriberModel, err := subscriberService.SubscribeToNewsletter(r.Context(), req.Email, newsletterIDStr)
		if err != nil {
			commonHandler.JSONErrorSecure(w, err, "subscriber subscribe")
			return
		}

		// The service returns *models.Subscriber. We should define a response struct
		// similar to service.SubscribeToNewsletterResponse or pass the model directly.
		// For consistency with other handlers, let's use a dedicated response struct.
		response := SubscribeResponse{
			SubscriberID: subscriberModel.ID,
			Email:        subscriberModel.Email,
			NewsletterID: subscriberModel.NewsletterID,
			Status:       subscriberModel.Status,
		}

		commonHandler.JSONResponse(w, response, http.StatusCreated)
	}
}

