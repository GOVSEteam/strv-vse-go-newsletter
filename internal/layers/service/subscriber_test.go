package service

import (
	"context"
	"testing"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockSubscriberRepository is a mock type for the SubscriberRepository type
type MockSubscriberRepository struct {
	mock.Mock
}

func (m *MockSubscriberRepository) GetSubscriberByEmailAndNewsletterID(email, newsletterID string) (*repository.Subscriber, error) {
	args := m.Called(email, newsletterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Subscriber), args.Error(1)
}

func (m *MockSubscriberRepository) CreateSubscriber(email, newsletterID, confirmationToken string) (*repository.Subscriber, error) {
	args := m.Called(email, newsletterID, confirmationToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Subscriber), args.Error(1)
}

func (m *MockSubscriberRepository) ConfirmSubscriber(confirmationToken string) (*repository.Subscriber, error) {
	args := m.Called(confirmationToken)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Subscriber), args.Error(1)
}

func (m *MockSubscriberRepository) ListSubscribersByNewsletterID(newsletterID string, limit, offset int) ([]repository.Subscriber, int, error) {
	args := m.Called(newsletterID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]repository.Subscriber), args.Int(1), args.Error(2)
}

// MockEmailService is defined in publishing_test.go

// Test Scenarios
func TestSubscribe_Success(t *testing.T) {
	ctx := context.Background()
	mockSubscriberRepo := new(MockSubscriberRepository)
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockEmailService := new(MockEmailService)

	subscriberService := NewSubscriberService(mockSubscriberRepo, mockNewsletterRepo, mockEmailService)

	email := "test@example.com"
	newsletterID := "newsletter-123"
	newsletter := &repository.Newsletter{ID: newsletterID, Name: "Test Newsletter"}
	expectedSubscriber := &repository.Subscriber{ID: "sub-123", Email: email, NewsletterID: newsletterID}

	mockNewsletterRepo.On("GetNewsletterByID", newsletterID).Return(newsletter, nil)
	mockSubscriberRepo.On("GetSubscriberByEmailAndNewsletterID", email, newsletterID).Return(nil, nil)
	mockSubscriberRepo.On("CreateSubscriber", email, newsletterID, mock.AnythingOfType("string")).Return(expectedSubscriber, nil)
	mockEmailService.On("SendConfirmationEmail", email, email, mock.AnythingOfType("string")).Return(nil)

	subscriber, err := subscriberService.Subscribe(ctx, email, newsletterID)

	assert.NoError(t, err)
	assert.NotNil(t, subscriber)
	assert.Equal(t, email, subscriber.Email)
	mockSubscriberRepo.AssertExpectations(t)
	mockNewsletterRepo.AssertExpectations(t)
	mockEmailService.AssertExpectations(t)
}

func TestSubscribe_AlreadySubscribed(t *testing.T) {
	ctx := context.Background()
	mockSubscriberRepo := new(MockSubscriberRepository)
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockEmailService := new(MockEmailService)

	subscriberService := NewSubscriberService(mockSubscriberRepo, mockNewsletterRepo, mockEmailService)

	email := "test@example.com"
	newsletterID := "newsletter-123"
	newsletter := &repository.Newsletter{ID: newsletterID, Name: "Test Newsletter"}
	existingSubscriber := &repository.Subscriber{ID: "sub-123", Email: email, NewsletterID: newsletterID, IsConfirmed: true}

	mockNewsletterRepo.On("GetNewsletterByID", newsletterID).Return(newsletter, nil)
	mockSubscriberRepo.On("GetSubscriberByEmailAndNewsletterID", email, newsletterID).Return(existingSubscriber, nil)

	subscriber, err := subscriberService.Subscribe(ctx, email, newsletterID)

	assert.Error(t, err)
	assert.Nil(t, subscriber)
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

	confirmationToken := "token-123"
	confirmedSubscriber := &repository.Subscriber{
		ID:           "sub-123",
		Email:        "test@example.com",
		NewsletterID: "newsletter-123",
		IsConfirmed:  true,
	}

	mockSubscriberRepo.On("ConfirmSubscriber", confirmationToken).Return(confirmedSubscriber, nil)

	subscriber, err := subscriberService.ConfirmSubscription(ctx, confirmationToken)

	assert.NoError(t, err)
	assert.NotNil(t, subscriber)
	assert.True(t, subscriber.IsConfirmed)
	mockSubscriberRepo.AssertExpectations(t)
}

func TestListSubscribers_Success(t *testing.T) {
	ctx := context.Background()
	mockSubscriberRepo := new(MockSubscriberRepository)
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockEmailService := new(MockEmailService)

	subscriberService := NewSubscriberService(mockSubscriberRepo, mockNewsletterRepo, mockEmailService)

	newsletterID := "newsletter-123"
	limit := 10
	offset := 0
	expectedSubscribers := []repository.Subscriber{
		{ID: "sub-1", Email: "user1@example.com", NewsletterID: newsletterID, IsConfirmed: true},
		{ID: "sub-2", Email: "user2@example.com", NewsletterID: newsletterID, IsConfirmed: true},
	}
	totalCount := 2

	mockSubscriberRepo.On("ListSubscribersByNewsletterID", newsletterID, limit, offset).Return(expectedSubscribers, totalCount, nil)

	subscribers, count, err := subscriberService.ListSubscribers(ctx, newsletterID, limit, offset)

	assert.NoError(t, err)
	assert.Equal(t, expectedSubscribers, subscribers)
	assert.Equal(t, totalCount, count)
	mockSubscriberRepo.AssertExpectations(t)
}
