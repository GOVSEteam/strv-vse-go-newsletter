package subscriber

import (
	"net/http"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
)

// ConfirmSubscriptionHandler handles requests to confirm a subscription using a token.
// GET /api/subscribers/confirm?token={token}
func ConfirmSubscriptionHandler(subscriberService service.SubscriberServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			handler.JSONError(w, "token query parameter is required", http.StatusBadRequest)
			return
		}

		req := service.ConfirmSubscriptionRequest{Token: token}
		err := subscriberService.ConfirmSubscription(r.Context(), req)
		if err != nil {
			if err == service.ErrInvalidOrExpiredToken {
				handler.JSONError(w, "Invalid or expired confirmation token.", http.StatusBadRequest)
			} else if err == service.ErrAlreadyConfirmed {
				handler.JSONError(w, "Subscription is already confirmed.", http.StatusConflict)
			} else {
				handler.JSONError(w, "Failed to confirm subscription: "+err.Error(), http.StatusInternalServerError)
			}
			return
		}

		// For an API, a 200 OK with a success message is fine.
		// A user-facing flow might redirect to a success page.
		handler.JSONResponse(w, map[string]string{"message": "Subscription confirmed successfully."}, http.StatusOK)
	}
}
