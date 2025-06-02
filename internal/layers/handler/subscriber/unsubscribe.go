package subscriber

import (
	"net/http"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	commonHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
)

// UnsubscribeHandler handles requests to unsubscribe from a newsletter using a token.
// GET /api/subscriptions/unsubscribe?token={token}
func UnsubscribeHandler(subscriberService service.SubscriberServiceInterface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			commonHandler.JSONError(w, "token query parameter is required", http.StatusBadRequest)
			return
		}

		err := subscriberService.UnsubscribeByToken(r.Context(), token)
		if err != nil {
			statusCode := apperrors.ErrorToHTTPStatus(err)
			// Provide a more generic message for token-related errors to avoid information leakage.
			message := err.Error()
			if statusCode == http.StatusBadRequest || statusCode == http.StatusNotFound {
				message = "Invalid or expired unsubscribe token."
			}
			commonHandler.JSONError(w, message, statusCode)
			return
		}

		commonHandler.JSONResponse(w, map[string]string{"message": "Successfully unsubscribed."}, http.StatusOK)
	}
}
