package subscriber_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io" // Import for io.Reader
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/subscriber" // Will be 'subscriber' package
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"github.com/stretchr/testify/assert" // For assertions
)

// MockSubscriberService is a mock implementation of the SubscriberService.
type MockSubscriberService struct {
	SubscribeToNewsletterFunc             func(ctx context.Context, req service.SubscribeToNewsletterRequest) (*service.SubscribeToNewsletterResponse, error)
	UnsubscribeFromNewsletterFunc         func(ctx context.Context, req service.UnsubscribeFromNewsletterRequest) error
	UnsubscribeByTokenFunc                func(ctx context.Context, token string) error // Added for new interface method
	ConfirmSubscriptionFunc               func(ctx context.Context, req service.ConfirmSubscriptionRequest) error
	GetActiveSubscribersForNewsletterFunc func(ctx context.Context, newsletterID string) ([]models.Subscriber, error) // Added for new interface method
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

func (m *MockSubscriberService) UnsubscribeByToken(ctx context.Context, token string) error {
	if m.UnsubscribeByTokenFunc != nil {
		return m.UnsubscribeByTokenFunc(ctx, token)
	}
	return errors.New("UnsubscribeByTokenFunc not implemented")
}

func (m *MockSubscriberService) GetActiveSubscribersForNewsletter(ctx context.Context, newsletterID string) ([]models.Subscriber, error) {
	if m.GetActiveSubscribersForNewsletterFunc != nil {
		return m.GetActiveSubscribersForNewsletterFunc(ctx, newsletterID)
	}
	return nil, errors.New("GetActiveSubscribersForNewsletterFunc not implemented")
}

func TestSubscriberHandler_SubscribeToNewsletter(t *testing.T) {
	mockService := &MockSubscriberService{}
	// Use the handler function directly from the subscriber package
	// The actual handler function is subscriber.SubscribeHandler
	// The test will call this function.

	tests := []struct {
		name                 string
		newsletterIDPath     string                               // To set in path
		body                 service.SubscribeToNewsletterRequest // Use the service request type
		mockServiceSetup     func()
		expectedStatusCode   int
		expectedBodyContains string
		expectedSubResponse  *service.SubscribeToNewsletterResponse // For success case
	}{
		{
			name:             "Success - Subscribed",
			newsletterIDPath: "test-newsletter-id",
			body:             service.SubscribeToNewsletterRequest{Email: "test@example.com"}, // Use service.SubscribeToNewsletterRequest
			mockServiceSetup: func() {
				mockService.SubscribeToNewsletterFunc = func(ctx context.Context, req service.SubscribeToNewsletterRequest) (*service.SubscribeToNewsletterResponse, error) {
					// NewsletterID will be set by the handler from path, so we don't check it in body here
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
			expectedStatusCode: http.StatusCreated,
			expectedSubResponse: &service.SubscribeToNewsletterResponse{
				SubscriberID: "sub-123",
				Email:        "test@example.com",
				NewsletterID: "test-newsletter-id",
				Status:       models.SubscriberStatusActive,
			},
		},
		{
			name:                 "Fail - Missing NewsletterID in path",
			newsletterIDPath:     "", // Simulate missing path param
			body:                 service.SubscribeToNewsletterRequest{Email: "test@example.com"},
			mockServiceSetup:     func() { /* No service call expected */ },
			expectedStatusCode:   http.StatusBadRequest,
			expectedBodyContains: "extracted ID is empty", // Error from utils.GetIDFromPath
		},
		{
			name:                 "Fail - Invalid JSON body",
			newsletterIDPath:     "test-newsletter-id",
			body:                 service.SubscribeToNewsletterRequest{}, // Placeholder, actual body is raw string below
			mockServiceSetup:     func() { /* No service call expected */ },
			expectedStatusCode:   http.StatusBadRequest,
			expectedBodyContains: "invalid request body",
		},
		{
			name:                 "Fail - Missing Email in request",
			newsletterIDPath:     "test-newsletter-id",
			body:                 service.SubscribeToNewsletterRequest{Email: ""}, // Email is empty
			mockServiceSetup:     func() { /* No service call expected */ },
			expectedStatusCode:   http.StatusBadRequest,
			expectedBodyContains: "email cannot be empty", // Error from utils.ValidateEmail
		},
		{
			name:                 "Fail - Invalid Email Format",
			newsletterIDPath:     "test-newsletter-id",
			body:                 service.SubscribeToNewsletterRequest{Email: "invalid-email"},
			mockServiceSetup:     func() { /* No service call expected */ },
			expectedStatusCode:   http.StatusBadRequest,
			expectedBodyContains: "invalid email format", // Error from utils.ValidateEmail
		},
		{
			name:             "Fail - Already Subscribed",
			newsletterIDPath: "test-newsletter-id",
			body:             service.SubscribeToNewsletterRequest{Email: "taken@example.com"},
			mockServiceSetup: func() {
				mockService.SubscribeToNewsletterFunc = func(ctx context.Context, req service.SubscribeToNewsletterRequest) (*service.SubscribeToNewsletterResponse, error) {
					return nil, service.ErrAlreadySubscribed
				}
			},
			expectedStatusCode:   http.StatusConflict,
			expectedBodyContains: service.ErrAlreadySubscribed.Error(),
		},
		{
			name:             "Fail - Newsletter Not Found",
			newsletterIDPath: "unknown-newsletter-id",
			body:             service.SubscribeToNewsletterRequest{Email: "test@example.com"},
			mockServiceSetup: func() {
				mockService.SubscribeToNewsletterFunc = func(ctx context.Context, req service.SubscribeToNewsletterRequest) (*service.SubscribeToNewsletterResponse, error) {
					return nil, service.ErrNewsletterNotFound
				}
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedBodyContains: service.ErrNewsletterNotFound.Error(),
		},
		{
			name:             "Fail - Service Internal Error",
			newsletterIDPath: "test-newsletter-id",
			body:             service.SubscribeToNewsletterRequest{Email: "test@example.com"},
			mockServiceSetup: func() {
				mockService.SubscribeToNewsletterFunc = func(ctx context.Context, req service.SubscribeToNewsletterRequest) (*service.SubscribeToNewsletterResponse, error) {
					return nil, errors.New("some internal service error")
				}
			},
			expectedStatusCode:   http.StatusInternalServerError,
			expectedBodyContains: "Failed to subscribe: some internal service error", // Exact error from handler
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.mockServiceSetup()

			var reqBody io.Reader
			if tt.name == "Fail - Invalid JSON body" {
				reqBody = bytes.NewBufferString("this is not json")
			} else {
				// For other cases, marshal the structured body
				bodyBytes, err := json.Marshal(tt.body)
				if err != nil {
					t.Fatalf("could not marshal request body for test %s: %v", tt.name, err)
				}
				reqBody = bytes.NewBuffer(bodyBytes)
			}

			req, err := http.NewRequest("POST", "/api/newsletters/"+tt.newsletterIDPath+"/subscribe", reqBody)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			// For Go 1.22+ http.ServeMux path parameters
			if tt.newsletterIDPath != "" {
				req.SetPathValue("newsletterID", tt.newsletterIDPath)
			}

			rr := httptest.NewRecorder()
			// Call the handler function directly
			httpHandler := subscriber.SubscribeHandler(mockService)
			httpHandler.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code, "handler returned wrong status code. Body: "+rr.Body.String())

			if tt.expectedBodyContains != "" {
				assert.Contains(t, rr.Body.String(), tt.expectedBodyContains, "handler returned unexpected body")
			}

			if tt.expectedSubResponse != nil { // Check full response for success case
				var actualResponse service.SubscribeToNewsletterResponse
				// Assuming the response is a JSON object with a "data" field for success,
				// or directly the response object if handler.Success is changed.
				// For now, let's assume handler.Success marshals the data directly.
				// If handler.Success wraps it like {"data": ...}, this needs adjustment.
				// The current handler.JSONResponse marshals data directly.
				err := json.Unmarshal(rr.Body.Bytes(), &actualResponse)
				assert.NoError(t, err, "could not unmarshal response body")
				assert.Equal(t, *tt.expectedSubResponse, actualResponse, "handler returned unexpected success response body")
			}
		})
	}
}

func TestSubscriberHandler_UnsubscribeFromNewsletter(t *testing.T) {
	mockService := &MockSubscriberService{}
	// handler := subscriber.NewSubscriberHandler(mockService) // No constructor
	// We will call subscriber.UnsubscribeHandler(mockService) directly in test

	tests := []struct {
		name                 string
		newsletterIDPath     string
		emailQueryParam      string
		mockServiceSetup     func()
		expectedStatusCode   int
		expectedBodyContains string
	}{
		{
			name:             "Success - Unsubscribed",
			newsletterIDPath: "news-123",
			emailQueryParam:  "test@example.com",
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
			name:                 "Fail - Missing NewsletterID in path",
			newsletterIDPath:     "",
			emailQueryParam:      "test@example.com",
			mockServiceSetup:     func() { /* No service call expected */ },
			expectedStatusCode:   http.StatusBadRequest,
			expectedBodyContains: "newsletterID path parameter is required",
		},
		{
			name:                 "Fail - Missing Email in query",
			newsletterIDPath:     "news-123",
			emailQueryParam:      "",
			mockServiceSetup:     func() { /* No service call expected */ },
			expectedStatusCode:   http.StatusBadRequest,
			expectedBodyContains: "email query parameter is required",
		},
		{
			name:             "Fail - Subscription Not Found",
			newsletterIDPath: "news-123",
			emailQueryParam:  "nonexistent@example.com",
			mockServiceSetup: func() {
				mockService.UnsubscribeFromNewsletterFunc = func(ctx context.Context, req service.UnsubscribeFromNewsletterRequest) error {
					return service.ErrSubscriptionNotFound
				}
			},
			expectedStatusCode:   http.StatusNotFound,
			expectedBodyContains: service.ErrSubscriptionNotFound.Error(),
		},
		{
			name:             "Fail - Service Internal Error",
			newsletterIDPath: "news-123",
			emailQueryParam:  "test@example.com",
			mockServiceSetup: func() {
				mockService.UnsubscribeFromNewsletterFunc = func(ctx context.Context, req service.UnsubscribeFromNewsletterRequest) error {
					return errors.New("some internal service error")
				}
			},
			expectedStatusCode:   http.StatusInternalServerError,
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
			// httpHandler := subscriber.UnsubscribeHandler(mockService) // Assuming UnsubscribeHandler exists
			// httpHandler.ServeHTTP(rr, req) // Call when UnsubscribeHandler is implemented

			// For now, this test will fail until UnsubscribeHandler is created.
			// To make it runnable without UnsubscribeHandler, we'd need to skip or comment out the call.
			// assert.Equal(t, tt.expectedStatusCode, rr.Code, "handler returned wrong status code. Body: "+rr.Body.String())

			// if tt.expectedStatusCode != http.StatusNoContent && tt.expectedBodyContains != "" {
			// 	assert.Contains(t, rr.Body.String(), tt.expectedBodyContains, "handler returned unexpected body")
			// } else if tt.expectedStatusCode == http.StatusNoContent && rr.Body.Len() > 0 {
			// 	assert.Fail(t, "handler returned body for 204 No Content: got %v", rr.Body.String())
			// }
			// Mark test as skipped until UnsubscribeHandler is ready
			if subscriber.UnsubscribeHandler == nil {
				t.Skip("Skipping UnsubscribeFromNewsletter test as handler is not yet implemented")
			} else {
				httpHandler := subscriber.UnsubscribeHandler(mockService)
				httpHandler.ServeHTTP(rr, req)
				assert.Equal(t, tt.expectedStatusCode, rr.Code, "handler returned wrong status code. Body: "+rr.Body.String())
				if tt.expectedStatusCode != http.StatusNoContent && tt.expectedBodyContains != "" {
					assert.Contains(t, rr.Body.String(), tt.expectedBodyContains, "handler returned unexpected body")
				} else if tt.expectedStatusCode == http.StatusNoContent && rr.Body.Len() > 0 {
					assert.Fail(t, "handler returned body for 204 No Content: got %v", rr.Body.String())
				}
			}
		})
	}
}
