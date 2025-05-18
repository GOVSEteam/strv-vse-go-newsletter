package subscriber_handler

import (
	"encoding/json"
	"errors"
	"net/http"

	// "github.com/go-chi/chi/v5" // No longer using chi here directly, assuming Go 1.22+ ServeMux

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
)

// SubscriberHandler handles HTTP requests for subscriber operations.
type SubscriberHandler struct {
	subscriberService *service.SubscriberService
}

// NewSubscriberHandler creates a new SubscriberHandler.
func NewSubscriberHandler(ss *service.SubscriberService) *SubscriberHandler {
	return &SubscriberHandler{
		subscriberService: ss,
	}
}

// SubscribeRequest is the expected request body for subscribing to a newsletter.
type SubscribeRequest struct {
	Email string `json:"email"`
}

// SubscribeToNewsletter handles the POST /api/newsletters/{newsletterID}/subscribe request.
func (h *SubscriberHandler) SubscribeToNewsletter(w http.ResponseWriter, r *http.Request) {
	newsletterID := r.PathValue("newsletterID") // Use r.PathValue for Go 1.22+ ServeMux
	if newsletterID == "" {
		handler.JSONError(w, "newsletterID path parameter is required", http.StatusBadRequest)
		return
	}

	var req SubscribeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		handler.JSONError(w, "invalid request body: "+err.Error(), http.StatusBadRequest)
		return
	}

	if req.Email == "" { // Additional validation, though service also checks
		handler.JSONError(w, "email is required in request body", http.StatusBadRequest)
		return
	}

	serviceReq := service.SubscribeToNewsletterRequest{
		Email:        req.Email,
		NewsletterID: newsletterID,
	}

	subResponse, err := h.subscriberService.SubscribeToNewsletter(r.Context(), serviceReq)
	if err != nil {
		if errors.Is(err, service.ErrAlreadySubscribed) {
			handler.JSONError(w, err.Error(), http.StatusConflict)
		} else if errors.Is(err, service.ErrNewsletterNotFound) {
			handler.JSONError(w, err.Error(), http.StatusNotFound)
		} else {
			// For other errors from the service layer
			handler.JSONError(w, "failed to subscribe to newsletter: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	handler.JSONResponse(w, subResponse, http.StatusCreated)
}

// UnsubscribeFromNewsletter handles the DELETE /api/newsletters/{newsletterID}/subscribers?email={email} request.
func (h *SubscriberHandler) UnsubscribeFromNewsletter(w http.ResponseWriter, r *http.Request) {
	newsletterID := r.PathValue("newsletterID")
	if newsletterID == "" {
		handler.JSONError(w, "newsletterID path parameter is required", http.StatusBadRequest)
		return
	}

	email := r.URL.Query().Get("email")
	if email == "" {
		handler.JSONError(w, "email query parameter is required", http.StatusBadRequest)
		return
	}

	serviceReq := service.UnsubscribeFromNewsletterRequest{
		Email:        email,
		NewsletterID: newsletterID,
	}

	err := h.subscriberService.UnsubscribeFromNewsletter(r.Context(), serviceReq)
	if err != nil {
		if errors.Is(err, service.ErrSubscriptionNotFound) {
			handler.JSONError(w, err.Error(), http.StatusNotFound)
		} else if errors.Is(err, service.ErrNewsletterNotFound) { // Though service currently doesn't return this for unsubscribe
			handler.JSONError(w, err.Error(), http.StatusNotFound) 
		} else {
			// For other errors from the service layer (e.g., validation, DB issues)
			handler.JSONError(w, "failed to unsubscribe: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent) // Successfully unsubscribed
}

// ConfirmSubscriptionHandler handles the GET /api/subscribers/confirm?token={token} request.
func (h *SubscriberHandler) ConfirmSubscriptionHandler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query().Get("token")
	if token == "" {
		handler.JSONError(w, "token query parameter is required", http.StatusBadRequest)
		return
	}

	serviceReq := service.ConfirmSubscriptionRequest{
		Token: token,
	}

	err := h.subscriberService.ConfirmSubscription(r.Context(), serviceReq)
	if err != nil {
		if errors.Is(err, service.ErrInvalidOrExpiredToken) {
			handler.JSONError(w, err.Error(), http.StatusBadRequest) // Or 404 if we prefer for not found/expired
		} else if errors.Is(err, service.ErrAlreadyConfirmed) {
			handler.JSONError(w, err.Error(), http.StatusConflict)
		} else {
			// For other errors from the service layer
			handler.JSONError(w, "failed to confirm subscription: "+err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// On successful confirmation, you might redirect to a success page on your frontend,
	// or simply return a success message.
	handler.JSONResponse(w, map[string]string{"message": "Subscription confirmed successfully"}, http.StatusOK)
} 