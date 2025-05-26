package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSubscriberRepository is a mock implementation of SubscriberRepository
type MockSubscriberRepository struct {
	mock.Mock
}

func (m *MockSubscriberRepository) CreateSubscriber(ctx context.Context, subscriber models.Subscriber) (string, error) {
	args := m.Called(ctx, subscriber)
	return args.String(0), args.Error(1)
}

func (m *MockSubscriberRepository) GetSubscriberByEmailAndNewsletterID(ctx context.Context, email string, newsletterID string) (*models.Subscriber, error) {
	args := m.Called(ctx, email, newsletterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscriber), args.Error(1)
}

func (m *MockSubscriberRepository) UpdateSubscriberStatus(ctx context.Context, subscriberID string, status models.SubscriberStatus) error {
	args := m.Called(ctx, subscriberID, status)
	return args.Error(0)
}

func (m *MockSubscriberRepository) UpdateSubscriberUnsubscribeToken(ctx context.Context, subscriberID string, newToken string) error {
	args := m.Called(ctx, subscriberID, newToken)
	return args.Error(0)
}

func (m *MockSubscriberRepository) GetSubscriberByConfirmationToken(ctx context.Context, token string) (*models.Subscriber, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscriber), args.Error(1)
}

func (m *MockSubscriberRepository) ConfirmSubscriber(ctx context.Context, subscriberID string, confirmationTime time.Time) error {
	args := m.Called(ctx, subscriberID, confirmationTime)
	return args.Error(0)
}

func (m *MockSubscriberRepository) GetSubscriberByUnsubscribeToken(ctx context.Context, token string) (*models.Subscriber, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscriber), args.Error(1)
}

func (m *MockSubscriberRepository) GetActiveSubscribersByNewsletterID(ctx context.Context, newsletterID string) ([]models.Subscriber, error) {
	args := m.Called(ctx, newsletterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Subscriber), args.Error(1)
}

// MockNewsletterRepository is a mock implementation of NewsletterRepository
type MockNewsletterRepository struct {
	mock.Mock
}

func (m *MockNewsletterRepository) GetNewsletterByID(newsletterID string) (*repository.Newsletter, error) {
	args := m.Called(newsletterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Newsletter), args.Error(1)
}

func (m *MockNewsletterRepository) ListNewslettersByEditorID(editorID string, limit int, offset int) ([]repository.Newsletter, int, error) {
	args := m.Called(editorID, limit, offset)
	return args.Get(0).([]repository.Newsletter), args.Int(1), args.Error(2)
}
func (m *MockNewsletterRepository) CreateNewsletter(editorID, name, description string) (*repository.Newsletter, error) {
	args := m.Called(editorID, name, description)
	return args.Get(0).(*repository.Newsletter), args.Error(1)
}
func (m *MockNewsletterRepository) GetNewsletterByIDAndEditorID(newsletterID string, editorID string) (*repository.Newsletter, error) {
	args := m.Called(newsletterID, editorID)
	return args.Get(0).(*repository.Newsletter), args.Error(1)
}
func (m *MockNewsletterRepository) UpdateNewsletter(newsletterID string, editorID string, name *string, description *string) (*repository.Newsletter, error) {
	args := m.Called(newsletterID, editorID, name, description)
	return args.Get(0).(*repository.Newsletter), args.Error(1)
}
func (m *MockNewsletterRepository) DeleteNewsletter(newsletterID string, editorID string) error {
	args := m.Called(newsletterID, editorID)
	return args.Error(0)
}
func (m *MockNewsletterRepository) GetNewsletterByNameAndEditorID(name string, editorID string) (*repository.Newsletter, error) {
	args := m.Called(name, editorID)
	return args.Get(0).(*repository.Newsletter), args.Error(1)
}

func TestSubscriberService_SubscribeToNewsletter(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name             string
		req              service.SubscribeToNewsletterRequest
		mockSetup        func(mockSubRepo *MockSubscriberRepository, mockNewsRepo *MockNewsletterRepository, mockEmailSvc *MockEmailService)
		expectedResponse *service.SubscribeToNewsletterResponse
		expectedError    error
	}{
		{
			name: "Success",
			req:  service.SubscribeToNewsletterRequest{Email: "test@example.com", NewsletterID: "news-123"},
			mockSetup: func(mockSubRepo *MockSubscriberRepository, mockNewsRepo *MockNewsletterRepository, mockEmailSvc *MockEmailService) {
				mockNewsRepo.On("GetNewsletterByID", "news-123").Return(&repository.Newsletter{ID: "news-123"}, nil).Once()
				mockSubRepo.On("GetSubscriberByEmailAndNewsletterID", ctx, "test@example.com", "news-123").Return(nil, nil).Once()
				mockSubRepo.On("CreateSubscriber", ctx, mock.AnythingOfType("models.Subscriber")).Return("sub-xyz", nil).Once()
				mockEmailSvc.On("SendConfirmationEmail", "test@example.com", "test@example.com", mock.AnythingOfType("string")).Return(nil).Once()
			},
			expectedResponse: &service.SubscribeToNewsletterResponse{
				SubscriberID: "sub-xyz",
				Email:        "test@example.com",
				NewsletterID: "news-123",
				Status:       models.SubscriberStatusActive, // Corrected status
			},
			expectedError: nil,
		},
		{
			name: "Fail - Empty Email",
			req:  service.SubscribeToNewsletterRequest{Email: "", NewsletterID: "news-123"},
			mockSetup: func(mockSubRepo *MockSubscriberRepository, mockNewsRepo *MockNewsletterRepository, mockEmailSvc *MockEmailService) {
			},
			expectedError: errors.New("email cannot be empty"),
		},
		{
			name: "Fail - Empty NewsletterID",
			req:  service.SubscribeToNewsletterRequest{Email: "test@example.com", NewsletterID: ""},
			mockSetup: func(mockSubRepo *MockSubscriberRepository, mockNewsRepo *MockNewsletterRepository, mockEmailSvc *MockEmailService) {
			},
			expectedError: errors.New("newsletter ID cannot be empty"),
		},
		{
			name: "Fail - Newsletter Not Found",
			req:  service.SubscribeToNewsletterRequest{Email: "test@example.com", NewsletterID: "unknown-news-id"},
			mockSetup: func(mockSubRepo *MockSubscriberRepository, mockNewsRepo *MockNewsletterRepository, mockEmailSvc *MockEmailService) {
				mockNewsRepo.On("GetNewsletterByID", "unknown-news-id").Return(nil, nil).Once()
			},
			expectedError: service.ErrNewsletterNotFound,
		},
		{
			name: "Fail - GetNewsletterByID Error",
			req:  service.SubscribeToNewsletterRequest{Email: "test@example.com", NewsletterID: "news-123"},
			mockSetup: func(mockSubRepo *MockSubscriberRepository, mockNewsRepo *MockNewsletterRepository, mockEmailSvc *MockEmailService) {
				mockNewsRepo.On("GetNewsletterByID", "news-123").Return(nil, errors.New("db error")).Once()
			},
			expectedError: errors.New("db error"),
		},
		{
			name: "Fail - Already Subscribed",
			req:  service.SubscribeToNewsletterRequest{Email: "test@example.com", NewsletterID: "news-123"},
			mockSetup: func(mockSubRepo *MockSubscriberRepository, mockNewsRepo *MockNewsletterRepository, mockEmailSvc *MockEmailService) {
				mockNewsRepo.On("GetNewsletterByID", "news-123").Return(&repository.Newsletter{ID: "news-123"}, nil).Once()
				mockSubRepo.On("GetSubscriberByEmailAndNewsletterID", ctx, "test@example.com", "news-123").Return(&models.Subscriber{ID: "sub-abc"}, nil).Once()
			},
			expectedError: service.ErrAlreadySubscribed,
		},
		{
			name: "Fail - GetSubscriberByEmailAndNewsletterID Error",
			req:  service.SubscribeToNewsletterRequest{Email: "test@example.com", NewsletterID: "news-123"},
			mockSetup: func(mockSubRepo *MockSubscriberRepository, mockNewsRepo *MockNewsletterRepository, mockEmailSvc *MockEmailService) {
				mockNewsRepo.On("GetNewsletterByID", "news-123").Return(&repository.Newsletter{ID: "news-123"}, nil).Once()
				mockSubRepo.On("GetSubscriberByEmailAndNewsletterID", ctx, "test@example.com", "news-123").Return(nil, errors.New("firestore error")).Once()
			},
			expectedError: errors.New("firestore error"),
		},
		{
			name: "Fail - CreateSubscriber Error",
			req:  service.SubscribeToNewsletterRequest{Email: "test@example.com", NewsletterID: "news-123"},
			mockSetup: func(mockSubRepo *MockSubscriberRepository, mockNewsRepo *MockNewsletterRepository, mockEmailSvc *MockEmailService) {
				mockNewsRepo.On("GetNewsletterByID", "news-123").Return(&repository.Newsletter{ID: "news-123"}, nil).Once()
				mockSubRepo.On("GetSubscriberByEmailAndNewsletterID", ctx, "test@example.com", "news-123").Return(nil, nil).Once()
				mockSubRepo.On("CreateSubscriber", ctx, mock.AnythingOfType("models.Subscriber")).Return("", errors.New("create failed")).Once()
			},
			expectedError: errors.New("failed to create subscriber: create failed"), // Updated expected error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSubRepo := new(MockSubscriberRepository)
			mockNewsRepo := new(MockNewsletterRepository)
			mockEmailSvc := new(MockEmailService)
			tt.mockSetup(mockSubRepo, mockNewsRepo, mockEmailSvc) // Pass mockEmailSvc

			svc := service.NewSubscriberService(mockSubRepo, mockNewsRepo, mockEmailSvc)
			resp, err := svc.SubscribeToNewsletter(ctx, tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
				assert.Equal(t, tt.expectedResponse.SubscriberID, resp.SubscriberID)
				assert.Equal(t, tt.expectedResponse.Email, resp.Email)
				assert.Equal(t, tt.expectedResponse.NewsletterID, resp.NewsletterID)
				assert.Equal(t, tt.expectedResponse.Status, resp.Status)
				// Can't directly compare time, but check if it's recent for the success case if needed
			}

			mockSubRepo.AssertExpectations(t)
			mockNewsRepo.AssertExpectations(t)
		})
	}
}

func TestSubscriberService_UnsubscribeFromNewsletter(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name          string
		req           service.UnsubscribeFromNewsletterRequest
		mockSetup     func(mockSubRepo *MockSubscriberRepository, mockNewsRepo *MockNewsletterRepository, mockEmailSvc *MockEmailService) // Added mockEmailSvc
		expectedError error
	}{
		{
			name: "Success - Unsubscribed",
			req:  service.UnsubscribeFromNewsletterRequest{Email: "test@example.com", NewsletterID: "news-123"},
			mockSetup: func(mockSubRepo *MockSubscriberRepository, mockNewsRepo *MockNewsletterRepository, mockEmailSvc *MockEmailService) { // Added mockEmailSvc
				mockSubRepo.On("GetSubscriberByEmailAndNewsletterID", ctx, "test@example.com", "news-123").
					Return(&models.Subscriber{ID: "sub-xyz", Email: "test@example.com", NewsletterID: "news-123", Status: models.SubscriberStatusActive}, nil).Once()
				mockSubRepo.On("UpdateSubscriberStatus", ctx, "sub-xyz", models.SubscriberStatusUnsubscribed).
					Return(nil).Once()
			},
			expectedError: nil,
		},
		{
			name: "Success - Already Unsubscribed",
			req:  service.UnsubscribeFromNewsletterRequest{Email: "test@example.com", NewsletterID: "news-123"},
			mockSetup: func(mockSubRepo *MockSubscriberRepository, mockNewsRepo *MockNewsletterRepository, mockEmailSvc *MockEmailService) { // Added mockEmailSvc
				mockSubRepo.On("GetSubscriberByEmailAndNewsletterID", ctx, "test@example.com", "news-123").
					Return(&models.Subscriber{ID: "sub-xyz", Email: "test@example.com", NewsletterID: "news-123", Status: models.SubscriberStatusUnsubscribed}, nil).Once()
				// UpdateSubscriberStatus should not be called
			},
			expectedError: nil,
		},
		{
			name: "Fail - Empty Email",
			req:  service.UnsubscribeFromNewsletterRequest{Email: "", NewsletterID: "news-123"},
			mockSetup: func(mockSubRepo *MockSubscriberRepository, mockNewsRepo *MockNewsletterRepository, mockEmailSvc *MockEmailService) {
			}, // Added mockEmailSvc
			expectedError: errors.New("email cannot be empty (deprecated method)"),
		},
		{
			name: "Fail - Empty NewsletterID",
			req:  service.UnsubscribeFromNewsletterRequest{Email: "test@example.com", NewsletterID: ""},
			mockSetup: func(mockSubRepo *MockSubscriberRepository, mockNewsRepo *MockNewsletterRepository, mockEmailSvc *MockEmailService) {
			}, // Added mockEmailSvc
			expectedError: errors.New("newsletter ID cannot be empty (deprecated method)"),
		},
		{
			name: "Fail - Subscription Not Found",
			req:  service.UnsubscribeFromNewsletterRequest{Email: "nonexistent@example.com", NewsletterID: "news-123"},
			mockSetup: func(mockSubRepo *MockSubscriberRepository, mockNewsRepo *MockNewsletterRepository, mockEmailSvc *MockEmailService) { // Added mockEmailSvc
				mockSubRepo.On("GetSubscriberByEmailAndNewsletterID", ctx, "nonexistent@example.com", "news-123").
					Return(nil, nil).Once()
			},
			expectedError: service.ErrSubscriptionNotFound,
		},
		{
			name: "Fail - GetSubscriberByEmailAndNewsletterID Error",
			req:  service.UnsubscribeFromNewsletterRequest{Email: "test@example.com", NewsletterID: "news-123"},
			mockSetup: func(mockSubRepo *MockSubscriberRepository, mockNewsRepo *MockNewsletterRepository, mockEmailSvc *MockEmailService) { // Added mockEmailSvc
				mockSubRepo.On("GetSubscriberByEmailAndNewsletterID", ctx, "test@example.com", "news-123").
					Return(nil, errors.New("firestore error")).Once()
			},
			expectedError: errors.New("firestore error"),
		},
		{
			name: "Fail - UpdateSubscriberStatus Error",
			req:  service.UnsubscribeFromNewsletterRequest{Email: "test@example.com", NewsletterID: "news-123"},
			mockSetup: func(mockSubRepo *MockSubscriberRepository, mockNewsRepo *MockNewsletterRepository, mockEmailSvc *MockEmailService) { // Added mockEmailSvc
				mockSubRepo.On("GetSubscriberByEmailAndNewsletterID", ctx, "test@example.com", "news-123").
					Return(&models.Subscriber{ID: "sub-xyz", Status: models.SubscriberStatusActive}, nil).Once()
				mockSubRepo.On("UpdateSubscriberStatus", ctx, "sub-xyz", models.SubscriberStatusUnsubscribed).
					Return(errors.New("update failed")).Once()
			},
			expectedError: errors.New("update failed"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSubRepo := new(MockSubscriberRepository)
			mockNewsRepo := new(MockNewsletterRepository)
			mockEmailSvc := new(MockEmailService)
			tt.mockSetup(mockSubRepo, mockNewsRepo, mockEmailSvc) // Pass mockEmailSvc

			svc := service.NewSubscriberService(mockSubRepo, mockNewsRepo, mockEmailSvc)
			err := svc.UnsubscribeFromNewsletter(ctx, tt.req)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tt.expectedError.Error())
			} else {
				assert.NoError(t, err)
			}

			mockSubRepo.AssertExpectations(t)
			mockNewsRepo.AssertExpectations(t) // Will assert if any unexpected calls were made
		})
	}
}

// MockEmailService is a mock implementation of EmailService
type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendConfirmationEmail(toEmail, recipientName, confirmationLink string) error {
	args := m.Called(toEmail, recipientName, confirmationLink)
	return args.Error(0)
}

func (m *MockEmailService) SendNewsletterIssue(toEmail, recipientName, subject, htmlContent, unsubscribeLink string) error {
	args := m.Called(toEmail, recipientName, subject, htmlContent, unsubscribeLink)
	return args.Error(0)
}
