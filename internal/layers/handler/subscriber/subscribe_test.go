package subscriber_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	h "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/subscriber"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSubscriberService is a mock implementation of SubscriberServiceInterface
type MockSubscriberService struct {
	mock.Mock
}

func (m *MockSubscriberService) SubscribeToNewsletter(ctx context.Context, req service.SubscribeToNewsletterRequest) (*service.SubscribeToNewsletterResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*service.SubscribeToNewsletterResponse), args.Error(1)
}

func (m *MockSubscriberService) ConfirmSubscription(ctx context.Context, req service.ConfirmSubscriptionRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockSubscriberService) UnsubscribeFromNewsletter(ctx context.Context, req service.UnsubscribeFromNewsletterRequest) error {
	args := m.Called(ctx, req)
	return args.Error(0)
}

func (m *MockSubscriberService) UnsubscribeByToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockSubscriberService) ListSubscribersByNewsletterID(ctx context.Context, newsletterID string, limit int, offset int) ([]service.SubscriberResponse, int, error) {
	args := m.Called(ctx, newsletterID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]service.SubscriberResponse), args.Int(1), args.Error(2)
}

func TestSubscribeHandler(t *testing.T) {
	mockService := new(MockSubscriberService)
	httpHandler := h.SubscribeHandler(mockService)

	t.Run("Success", func(t *testing.T) {
		reqBody := service.SubscribeToNewsletterRequest{
			Email: "test@example.com",
		}
		expectedResponse := &service.SubscribeToNewsletterResponse{
			Message: "Subscription successful",
		}

		mockService.On("SubscribeToNewsletter", mock.Anything, mock.MatchedBy(func(req service.SubscribeToNewsletterRequest) bool {
			return req.Email == "test@example.com" && req.NewsletterID == "newsletter-123"
		})).Return(expectedResponse, nil).Once()

		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters/newsletter-123/subscribe", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusCreated, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Error - Invalid Email", func(t *testing.T) {
		reqBody := service.SubscribeToNewsletterRequest{
			Email: "invalid-email",
		}

		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters/newsletter-123/subscribe", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusBadRequest, rr.Code)
	})

	t.Run("Error - Already Subscribed", func(t *testing.T) {
		reqBody := service.SubscribeToNewsletterRequest{
			Email: "test@example.com",
		}

		mockService.On("SubscribeToNewsletter", mock.Anything, mock.Anything).Return(nil, service.ErrAlreadySubscribed).Once()

		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters/newsletter-123/subscribe", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusConflict, rr.Code)
		mockService.AssertExpectations(t)
	})

	t.Run("Error - Newsletter Not Found", func(t *testing.T) {
		reqBody := service.SubscribeToNewsletterRequest{
			Email: "test@example.com",
		}

		mockService.On("SubscribeToNewsletter", mock.Anything, mock.Anything).Return(nil, service.ErrNewsletterNotFound).Once()

		bodyBytes, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/newsletters/newsletter-123/subscribe", bytes.NewReader(bodyBytes))
		rr := httptest.NewRecorder()

		httpHandler.ServeHTTP(rr, req)

		assert.Equal(t, http.StatusNotFound, rr.Code)
		mockService.AssertExpectations(t)
	})
} 