package subscriber_test

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	h "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/subscriber"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestConfirmSubscriptionHandler(t *testing.T) {
	mockService := new(MockSubscriberService)
	httpHandler := h.ConfirmSubscriptionHandler(mockService)

	t.Run("Success", func(t *testing.T) {
		token := "valid-token-123"
		mockService.On("ConfirmSubscription", mock.Anything, service.ConfirmSubscriptionRequest{Token: token}).Return(nil).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/subscribers/confirm?token="+token, nil)
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusOK, rr.Code)
		assert.Contains(t, rr.Body.String(), "confirmed successfully")
		mockService.AssertExpectations(t)
	})

	t.Run("Error - Missing Token", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/subscribers/confirm", nil)
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "token query parameter is required")
	})

	t.Run("Error - Invalid Token", func(t *testing.T) {
		token := "invalid-token"
		mockService.On("ConfirmSubscription", mock.Anything, service.ConfirmSubscriptionRequest{Token: token}).Return(service.ErrInvalidOrExpiredToken).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/subscribers/confirm?token="+token, nil)
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
		assert.Contains(t, rr.Body.String(), "Invalid or expired confirmation token")
		mockService.AssertExpectations(t)
	})

	t.Run("Error - Already Confirmed", func(t *testing.T) {
		token := "already-confirmed-token"
		mockService.On("ConfirmSubscription", mock.Anything, service.ConfirmSubscriptionRequest{Token: token}).Return(service.ErrAlreadyConfirmed).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/subscribers/confirm?token="+token, nil)
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
		assert.Contains(t, rr.Body.String(), "already confirmed")
		mockService.AssertExpectations(t)
	})

	t.Run("Error - Service Error", func(t *testing.T) {
		token := "service-error-token"
		mockService.On("ConfirmSubscription", mock.Anything, service.ConfirmSubscriptionRequest{Token: token}).Return(errors.New("database error")).Once()

		req := httptest.NewRequest(http.MethodGet, "/api/subscribers/confirm?token="+token, nil)
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusInternalServerError, rr.Code)
		assert.Contains(t, rr.Body.String(), "Failed to confirm subscription")
		mockService.AssertExpectations(t)
	})
} 