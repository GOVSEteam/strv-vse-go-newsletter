package service

import (
	"context"
	"testing"
	"time"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSubscriberRepository is a mock type for the SubscriberRepository type
type MockSubscriberRepository struct {
	mock.Mock
}

func (m *MockSubscriberRepository) GetSubscriberByEmailAndNewsletterID(ctx context.Context, email, newsletterID string) (*models.Subscriber, error) {
	args := m.Called(ctx, email, newsletterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscriber), args.Error(1)
}

func (m *MockSubscriberRepository) CreateSubscriber(ctx context.Context, subscriber models.Subscriber) (string, error) {
	args := m.Called(ctx, subscriber)
	return args.String(0), args.Error(1)
}

func (m *MockSubscriberRepository) ConfirmSubscriber(ctx context.Context, token string, confirmedAt time.Time) error {
	args := m.Called(ctx, token, confirmedAt)
	return args.Error(0)
}

func (m *MockSubscriberRepository) UpdateSubscriberStatus(ctx context.Context, subscriberID string, status models.SubscriberStatus) error {
	args := m.Called(ctx, subscriberID, status)
	return args.Error(0)
}

func (m *MockSubscriberRepository) GetSubscriberByConfirmationToken(ctx context.Context, token string) (*models.Subscriber, error) {
	args := m.Called(ctx, token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscriber), args.Error(1)
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

// Test Scenarios
func TestSubscribeToNewsletter_Success(t *testing.T) {
	ctx := context.Background()
	mockSubscriberRepo := new(MockSubscriberRepository)
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockEmailService := new(MockEmailService)

	subscriberService := NewSubscriberService(mockSubscriberRepo, mockNewsletterRepo, mockEmailService)

	req := SubscribeToNewsletterRequest{
		Email:        "test@example.com",
		NewsletterID: "newsletter-123",
	}
	newsletter := &repository.Newsletter{ID: req.NewsletterID, Name: "Test Newsletter"}

	mockNewsletterRepo.On("GetNewsletterByID", req.NewsletterID).Return(newsletter, nil)
	mockSubscriberRepo.On("GetSubscriberByEmailAndNewsletterID", ctx, req.Email, req.NewsletterID).Return(nil, nil)
	mockSubscriberRepo.On("CreateSubscriber", ctx, mock.AnythingOfType("models.Subscriber")).Return("sub-123", nil)
	mockEmailService.On("SendConfirmationEmail", req.Email, req.Email, mock.AnythingOfType("string")).Return(nil)

	response, err := subscriberService.SubscribeToNewsletter(ctx, req)

	assert.NoError(t, err)
	assert.NotNil(t, response)
	assert.Equal(t, "sub-123", response.SubscriberID)
	assert.Equal(t, req.Email, response.Email)
	assert.Equal(t, req.NewsletterID, response.NewsletterID)
	assert.Equal(t, models.SubscriberStatusPendingConfirmation, response.Status)
	mockSubscriberRepo.AssertExpectations(t)
	mockNewsletterRepo.AssertExpectations(t)
	mockEmailService.AssertExpectations(t)
}

func TestSubscribeToNewsletter_AlreadySubscribed(t *testing.T) {
	ctx := context.Background()
	mockSubscriberRepo := new(MockSubscriberRepository)
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockEmailService := new(MockEmailService)

	subscriberService := NewSubscriberService(mockSubscriberRepo, mockNewsletterRepo, mockEmailService)

	req := SubscribeToNewsletterRequest{
		Email:        "test@example.com",
		NewsletterID: "newsletter-123",
	}
	newsletter := &repository.Newsletter{ID: req.NewsletterID, Name: "Test Newsletter"}
	existingSubscriber := &models.Subscriber{ID: "sub-123", Email: req.Email, NewsletterID: req.NewsletterID, Status: models.SubscriberStatusActive}

	mockNewsletterRepo.On("GetNewsletterByID", req.NewsletterID).Return(newsletter, nil)
	mockSubscriberRepo.On("GetSubscriberByEmailAndNewsletterID", ctx, req.Email, req.NewsletterID).Return(existingSubscriber, nil)

	response, err := subscriberService.SubscribeToNewsletter(ctx, req)

	assert.Error(t, err)
	assert.Nil(t, response)
	assert.EqualError(t, err, ErrAlreadySubscribed.Error())
	mockSubscriberRepo.AssertExpectations(t)
	mockNewsletterRepo.AssertExpectations(t)
}

func TestConfirmSubscription_Success(t *testing.T) {
	ctx := context.Background()
	mockSubscriberRepo := new(MockSubscriberRepository)
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockEmailService := new(MockEmailService)

	subscriberService := NewSubscriberService(mockSubscriberRepo, mockNewsletterRepo, mockEmailService)

	req := ConfirmSubscriptionRequest{Token: "token-123"}
	subscriber := &models.Subscriber{
		ID:              "sub-123",
		Email:           "test@example.com",
		NewsletterID:    "newsletter-123",
		Status:          models.SubscriberStatusPendingConfirmation,
		TokenExpiryTime: time.Now().Add(time.Hour),
	}

	mockSubscriberRepo.On("GetSubscriberByConfirmationToken", ctx, req.Token).Return(subscriber, nil)
	mockSubscriberRepo.On("ConfirmSubscriber", ctx, subscriber.ID, mock.AnythingOfType("time.Time")).Return(nil)

	err := subscriberService.ConfirmSubscription(ctx, req)

	assert.NoError(t, err)
	mockSubscriberRepo.AssertExpectations(t)
}

func TestGetActiveSubscribersForNewsletter_Success(t *testing.T) {
	ctx := context.Background()
	mockSubscriberRepo := new(MockSubscriberRepository)
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockEmailService := new(MockEmailService)

	subscriberService := NewSubscriberService(mockSubscriberRepo, mockNewsletterRepo, mockEmailService)

	newsletterID := "newsletter-123"
	expectedSubscribers := []models.Subscriber{
		{ID: "sub-1", Email: "user1@example.com", NewsletterID: newsletterID, Status: models.SubscriberStatusActive},
		{ID: "sub-2", Email: "user2@example.com", NewsletterID: newsletterID, Status: models.SubscriberStatusActive},
	}

	newsletter := &repository.Newsletter{ID: newsletterID, Name: "Test Newsletter"}
	mockNewsletterRepo.On("GetNewsletterByID", newsletterID).Return(newsletter, nil)
	mockSubscriberRepo.On("GetActiveSubscribersByNewsletterID", ctx, newsletterID).Return(expectedSubscribers, nil)

	subscribers, err := subscriberService.GetActiveSubscribersForNewsletter(ctx, newsletterID)

	assert.NoError(t, err)
	assert.Equal(t, expectedSubscribers, subscribers)
	mockSubscriberRepo.AssertExpectations(t)
}
