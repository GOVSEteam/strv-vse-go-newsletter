package subscriber

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models" // For models.SubscriberStatus
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/utils"  // Retaining for ValidateEmail
)

// SubscribeRequest defines the expected JSON request body for subscribing.
// Note: NewsletterID is taken from the path, not the body.
type SubscribeRequest struct {
	Email string `json:"email"`
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
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			commonHandler.JSONError(w, "Invalid request body: "+err.Error(), http.StatusBadRequest)
			return
		}
		defer r.Body.Close()

		if err := utils.ValidateEmail(req.Email); err != nil {
			valErr := fmt.Errorf("email validation failed: %w", apperrors.ErrValidation)
			statusCode := apperrors.ErrorToHTTPStatus(valErr)
			commonHandler.JSONError(w, valErr.Error(), statusCode)
			return
		}

		// Call service with Email from request body and NewsletterID from path
		subscriberModel, err := subscriberService.SubscribeToNewsletter(r.Context(), req.Email, newsletterIDStr)
		if err != nil {
			statusCode := apperrors.ErrorToHTTPStatus(err)
			commonHandler.JSONError(w, err.Error(), statusCode)
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

