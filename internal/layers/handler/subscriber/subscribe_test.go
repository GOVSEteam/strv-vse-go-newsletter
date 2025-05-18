package subscriber_handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	subscriberHandler "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/subscriber"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
)

// MockSubscriberService is a mock implementation of the SubscriberService.
type MockSubscriberService struct {
	SubscribeToNewsletterFunc   func(ctx context.Context, req service.SubscribeToNewsletterRequest) (*service.SubscribeToNewsletterResponse, error)
	UnsubscribeFromNewsletterFunc func(ctx context.Context, req service.UnsubscribeFromNewsletterRequest) error
	ConfirmSubscriptionFunc     func(ctx context.Context, req service.ConfirmSubscriptionRequest) error
}

func (m *MockSubscriberService) SubscribeToNewsletter(ctx context.Context, req service.SubscribeToNewsletterRequest) (*service.SubscribeToNewsletterResponse, error) {
	if m.SubscribeToNewsletterFunc != nil {
		return m.SubscribeToNewsletterFunc(ctx, req)
	}
	return nil, errors.New("SubscribeToNewsletterFunc not implemented")
}

func (m *MockSubscriberService) UnsubscribeFromNewsletter(ctx context.Context, req service.UnsubscribeFromNewsletterRequest) error {
	if m.UnsubscribeFromNewsletterFunc != nil {
		return m.UnsubscribeFromNewsletterFunc(ctx, req)
	}
	return errors.New("UnsubscribeFromNewsletterFunc not implemented")
}

func (m *MockSubscriberService) ConfirmSubscription(ctx context.Context, req service.ConfirmSubscriptionRequest) error {
	if m.ConfirmSubscriptionFunc != nil {
		return m.ConfirmSubscriptionFunc(ctx, req)
	}
	return errors.New("ConfirmSubscriptionFunc not implemented")
}

func TestSubscriberHandler_SubscribeToNewsletter(t *testing.T) {
	mockService := &MockSubscriberService{}
	handler := subscriberHandler.NewSubscriberHandler(mockService)

	tests := []struct {
		name                   string
		newsletterIDPath       string // To set in path
		body                   interface{}
		mockServiceSetup       func()
		expectedStatusCode     int
		expectedBodyContains   string
		expectedSubResponse    *service.SubscribeToNewsletterResponse // For success case
	}{
		{
			name:                "Success - Subscribed",
			newsletterIDPath:    "test-newsletter-id",
			body:                subscriberHandler.SubscribeRequest{Email: "test@example.com"},
			mockServiceSetup: func() {
				mockService.SubscribeToNewsletterFunc = func(ctx context.Context, req service.SubscribeToNewsletterRequest) (*service.SubscribeToNewsletterResponse, error) {
					if req.Email == "test@example.com" && req.NewsletterID == "test-newsletter-id" {
						return &service.SubscribeToNewsletterResponse{
							SubscriberID: "sub-123",
							Email:        "test@example.com",
							NewsletterID: "test-newsletter-id",
							Status:       models.SubscriberStatusActive,
						}, nil
					}
					return nil, errors.New("unexpected input to mock service")
				}
			},
			expectedStatusCode:  http.StatusCreated,
			expectedSubResponse: &service.SubscribeToNewsletterResponse{
				SubscriberID: "sub-123",
				Email:        "test@example.com",
				NewsletterID: "test-newsletter-id",
				Status:       models.SubscriberStatusActive,
			},
		},
		{
			name:                "Fail - Missing NewsletterID in path",
			newsletterIDPath:    "", // Simulate missing path param
			body:                subscriberHandler.SubscribeRequest{Email: "test@example.com"},
			mockServiceSetup:    func() { /* No service call expected */ },
			expectedStatusCode:  http.StatusBadRequest,
			expectedBodyContains: "newsletterID path parameter is required",
		},
		{
			name:                "Fail - Invalid JSON body",
			newsletterIDPath:    "test-newsletter-id",
			body:                "not-json",
			mockServiceSetup:    func() { /* No service call expected */ },
			expectedStatusCode:  http.StatusBadRequest,
			expectedBodyContains: "invalid request body",
		},
		{
			name:                "Fail - Missing Email in request",
			newsletterIDPath:    "test-newsletter-id",
			body:                subscriberHandler.SubscribeRequest{Email: ""},
			mockServiceSetup:    func() { /* No service call expected */ },
			expectedStatusCode:  http.StatusBadRequest,
			expectedBodyContains: "email is required in request body",
		},
		{
			name:                "Fail - Already Subscribed",
			newsletterIDPath:    "test-newsletter-id",
			body:                subscriberHandler.SubscribeRequest{Email: "taken@example.com"},
			mockServiceSetup: func() {
				mockService.SubscribeToNewsletterFunc = func(ctx context.Context, req service.SubscribeToNewsletterRequest) (*service.SubscribeToNewsletterResponse, error) {
					return nil, service.ErrAlreadySubscribed
				}
			},
			expectedStatusCode:  http.StatusConflict,
			expectedBodyContains: service.ErrAlreadySubscribed.Error(),
		},
		{
			name:                "Fail - Newsletter Not Found",
			newsletterIDPath:    "unknown-newsletter-id",
			body:                subscriberHandler.SubscribeRequest{Email: "test@example.com"},
			mockServiceSetup: func() {
				mockService.SubscribeToNewsletterFunc = func(ctx context.Context, req service.SubscribeToNewsletterRequest) (*service.SubscribeToNewsletterResponse, error) {
					return nil, service.ErrNewsletterNotFound
				}
			},
			expectedStatusCode:  http.StatusNotFound,
			expectedBodyContains: service.ErrNewsletterNotFound.Error(),
		},
		{
			name:                "Fail - Service Internal Error",
			newsletterIDPath:    "test-newsletter-id",
			body:                subscriberHandler.SubscribeRequest{Email: "test@example.com"},
			mockServiceSetup: func() {
				mockService.SubscribeToNewsletterFunc = func(ctx context.Context, req service.SubscribeToNewsletterRequest) (*service.SubscribeToNewsletterResponse, error) {
					return nil, errors.New("some internal service error")
				}
			},
			expectedStatusCode:  http.StatusInternalServerError,
			expectedBodyContains: "failed to subscribe to newsletter: some internal service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockServiceSetup()

			bodyBytes, _ := json.Marshal(tt.body)
			req, err := http.NewRequest("POST", "/api/newsletters/"+tt.newsletterIDPath+"/subscribe", bytes.NewBuffer(bodyBytes))
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			// For Go 1.22+ http.ServeMux path parameters
			if tt.newsletterIDPath != "" {
				req.SetPathValue("newsletterID", tt.newsletterIDPath)
			}

			rr := httptest.NewRecorder()
			handler.SubscribeToNewsletter(rr, req)

			if status := rr.Code; status != tt.expectedStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatusCode)
				t.Errorf("response body: %s", rr.Body.String()) // Log body for easier debugging
			}

			if tt.expectedBodyContains != "" {
				if !bytes.Contains(rr.Body.Bytes(), []byte(tt.expectedBodyContains)) {
					t.Errorf("handler returned unexpected body: got %v want to contain %v", rr.Body.String(), tt.expectedBodyContains)
				}
			}

			if tt.expectedSubResponse != nil { // Check full response for success case
				var actualResponse service.SubscribeToNewsletterResponse
				if err := json.Unmarshal(rr.Body.Bytes(), &actualResponse); err != nil {
					t.Fatalf("could not unmarshal response body: %v", err)
				}
				if actualResponse.SubscriberID != tt.expectedSubResponse.SubscriberID ||
					actualResponse.Email != tt.expectedSubResponse.Email ||
					actualResponse.NewsletterID != tt.expectedSubResponse.NewsletterID ||
					actualResponse.Status != tt.expectedSubResponse.Status {
					t.Errorf("handler returned unexpected success response body:\ngot %+v\nwant %+v", actualResponse, *tt.expectedSubResponse)
				}
			}
		})
	}
}

func TestSubscriberHandler_UnsubscribeFromNewsletter(t *testing.T) {
	mockService := &MockSubscriberService{}
	handler := subscriberHandler.NewSubscriberHandler(mockService)

	tests := []struct {
		name                 string
		newsletterIDPath     string
		emailQueryParam      string
		mockServiceSetup     func()
		expectedStatusCode   int
		expectedBodyContains string
	}{
		{
			name:               "Success - Unsubscribed",
			newsletterIDPath:   "news-123",
			emailQueryParam:    "test@example.com",
			mockServiceSetup: func() {
				mockService.UnsubscribeFromNewsletterFunc = func(ctx context.Context, req service.UnsubscribeFromNewsletterRequest) error {
					if req.Email == "test@example.com" && req.NewsletterID == "news-123" {
						return nil
					}
					return errors.New("unexpected input to mock service for unsubscribe")
				}
			},
			expectedStatusCode: http.StatusNoContent,
		},
		{
			name:               "Fail - Missing NewsletterID in path",
			newsletterIDPath:   "",
			emailQueryParam:    "test@example.com",
			mockServiceSetup:   func() { /* No service call expected */ },
			expectedStatusCode: http.StatusBadRequest,
			expectedBodyContains: "newsletterID path parameter is required",
		},
		{
			name:               "Fail - Missing Email in query",
			newsletterIDPath:   "news-123",
			emailQueryParam:    "",
			mockServiceSetup:   func() { /* No service call expected */ },
			expectedStatusCode: http.StatusBadRequest,
			expectedBodyContains: "email query parameter is required",
		},
		{
			name:               "Fail - Subscription Not Found",
			newsletterIDPath:   "news-123",
			emailQueryParam:    "nonexistent@example.com",
			mockServiceSetup: func() {
				mockService.UnsubscribeFromNewsletterFunc = func(ctx context.Context, req service.UnsubscribeFromNewsletterRequest) error {
					return service.ErrSubscriptionNotFound
				}
			},
			expectedStatusCode: http.StatusNotFound,
			expectedBodyContains: service.ErrSubscriptionNotFound.Error(),
		},
		{
			name:               "Fail - Service Internal Error",
			newsletterIDPath:   "news-123",
			emailQueryParam:    "test@example.com",
			mockServiceSetup: func() {
				mockService.UnsubscribeFromNewsletterFunc = func(ctx context.Context, req service.UnsubscribeFromNewsletterRequest) error {
					return errors.New("some internal service error")
				}
			},
			expectedStatusCode: http.StatusInternalServerError,
			expectedBodyContains: "failed to unsubscribe: some internal service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockServiceSetup()

			url := "/api/newsletters/" + tt.newsletterIDPath + "/subscribers"
			if tt.emailQueryParam != "" {
				url += "?email=" + tt.emailQueryParam
			}

			req, err := http.NewRequest("DELETE", url, nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			if tt.newsletterIDPath != "" {
				req.SetPathValue("newsletterID", tt.newsletterIDPath)
			}

			rr := httptest.NewRecorder()
			handler.UnsubscribeFromNewsletter(rr, req)

			if status := rr.Code; status != tt.expectedStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatusCode)
				t.Errorf("response body: %s", rr.Body.String()) 
			}

			if tt.expectedStatusCode != http.StatusNoContent && tt.expectedBodyContains != "" {
				if !bytes.Contains(rr.Body.Bytes(), []byte(tt.expectedBodyContains)) {
					t.Errorf("handler returned unexpected body: got %v want to contain %v", rr.Body.String(), tt.expectedBodyContains)
				}
			} else if tt.expectedStatusCode == http.StatusNoContent && rr.Body.Len() > 0 {
				t.Errorf("handler returned body for 204 No Content: got %v", rr.Body.String())
			}
		})
	}
}

func TestSubscriberHandler_ConfirmSubscriptionHandler(t *testing.T) {
	mockService := &MockSubscriberService{}
	handler := subscriberHandler.NewSubscriberHandler(mockService)

	tests := []struct {
		name                 string
		tokenQueryParam      string
		mockServiceSetup     func()
		expectedStatusCode   int
		expectedBodyContains string
	}{
		{
			name:            "Success - Subscription Confirmed",
			tokenQueryParam: "valid-token",
			mockServiceSetup: func() {
				mockService.ConfirmSubscriptionFunc = func(ctx context.Context, req service.ConfirmSubscriptionRequest) error {
					if req.Token == "valid-token" {
						return nil
					}
					return errors.New("unexpected token in mock service for confirm")
				}
			},
			expectedStatusCode:   http.StatusOK,
			expectedBodyContains: "Subscription confirmed successfully",
		},
		{
			name:               "Fail - Missing Token",
			tokenQueryParam:    "",
			mockServiceSetup:   func() { /* No service call expected */ },
			expectedStatusCode: http.StatusBadRequest,
			expectedBodyContains: "token query parameter is required",
		},
		{
			name:            "Fail - Invalid or Expired Token",
			tokenQueryParam: "invalid-token",
			mockServiceSetup: func() {
				mockService.ConfirmSubscriptionFunc = func(ctx context.Context, req service.ConfirmSubscriptionRequest) error {
					return service.ErrInvalidOrExpiredToken
				}
			},
			expectedStatusCode:   http.StatusBadRequest, // As per current handler mapping
			expectedBodyContains: service.ErrInvalidOrExpiredToken.Error(),
		},
		{
			name:            "Fail - Already Confirmed",
			tokenQueryParam: "valid-token-already-confirmed",
			mockServiceSetup: func() {
				mockService.ConfirmSubscriptionFunc = func(ctx context.Context, req service.ConfirmSubscriptionRequest) error {
					return service.ErrAlreadyConfirmed
				}
			},
			expectedStatusCode:   http.StatusConflict,
			expectedBodyContains: service.ErrAlreadyConfirmed.Error(),
		},
		{
			name:            "Fail - Service Internal Error",
			tokenQueryParam: "any-token",
			mockServiceSetup: func() {
				mockService.ConfirmSubscriptionFunc = func(ctx context.Context, req service.ConfirmSubscriptionRequest) error {
					return errors.New("some internal service error")
				}
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedBodyContains: "failed to confirm subscription: some internal service error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockServiceSetup()

			url := "/api/subscribers/confirm"
			if tt.tokenQueryParam != "" {
				url += "?token=" + tt.tokenQueryParam
			}

			req, err := http.NewRequest("GET", url, nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			rr := httptest.NewRecorder()
			handler.ConfirmSubscriptionHandler(rr, req)

			if status := rr.Code; status != tt.expectedStatusCode {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatusCode)
				t.Errorf("response body: %s", rr.Body.String())
			}

			if tt.expectedBodyContains != "" {
				if !bytes.Contains(rr.Body.Bytes(), []byte(tt.expectedBodyContains)) {
					t.Errorf("handler returned unexpected body: got %v want to contain %v", rr.Body.String(), tt.expectedBodyContains)
				}
			}
		})
	}
} 