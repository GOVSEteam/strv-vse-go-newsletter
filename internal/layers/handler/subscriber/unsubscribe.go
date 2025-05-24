package subscriber

import (
	"net/http"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
)

// UnsubscribeHandler handles requests to unsubscribe from a newsletter using a token.
// GET /api/subscriptions/unsubscribe?token={token}
func UnsubscribeHandler(subscriberService service.SubscriberServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			handler.JSONError(w, "token query parameter is required", http.StatusBadRequest)
			return
		}

		// The UnsubscribeFromNewsletter service method needs to be updated
		// to accept a token instead of Email & NewsletterID.
		// For now, we'll create a temporary request struct for the existing service method signature,
		// assuming the service will be refactored.
		// OR, we create a new service method like UnsubscribeByToken.
		// Let's assume the service will be updated to have:
		// UnsubscribeByToken(ctx context.Context, token string) error
		// And the current UnsubscribeFromNewsletter will be deprecated or removed.

		// For the purpose of this handler, we'll assume a new service method or refactored one.
		// Let's define a new request for a hypothetical service method for now.
		// This part will need to align with service layer changes.
		
		// req := service.UnsubscribeByTokenRequest{Token: token} // Hypothetical
		// err := subscriberService.UnsubscribeByToken(r.Context(), req) // Hypothetical

		// Given the current service.UnsubscribeFromNewsletterRequest expects Email and NewsletterID,
		// and the task is to implement token-based unsubscription,
		// the service layer MUST be updated first.
		// For now, this handler cannot fully call the existing service method correctly for token-based unsubscribe.
		// We will proceed by calling a *new* (yet to be implemented) service method.
		// This highlights the dependency: Handler depends on Service Layer.

		err := subscriberService.UnsubscribeByToken(r.Context(), token) // Assuming this method will be added to the service
		if err != nil {
			if err == service.ErrSubscriptionNotFound || err == service.ErrInvalidOrExpiredToken { // Assuming service might return this for bad tokens
				handler.JSONError(w, "Invalid or expired unsubscribe token.", http.StatusBadRequest)
			} else {
				handler.JSONError(w, "Failed to unsubscribe: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// Consider what to return. A simple success message or redirect.
		// For an API, a 200 OK or 204 No Content is typical.
		// A user-facing page might show a success message.
		// For now, a simple JSON response.
		handler.JSONResponse(w, map[string]string{"message": "Successfully unsubscribed."}, http.StatusOK)
	}
}
