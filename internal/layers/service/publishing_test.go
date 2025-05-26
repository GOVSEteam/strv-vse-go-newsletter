package service

import (
	"context"
	"errors"
	"testing"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock implementations
type MockNewsletterService struct {
	mock.Mock
}

func (m *MockNewsletterService) GetPostForPublishing(ctx context.Context, postID uuid.UUID, editorFirebaseUID string) (*models.Post, error) {
	args := m.Called(ctx, postID, editorFirebaseUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockNewsletterService) MarkPostAsPublished(ctx context.Context, editorFirebaseUID string, postID uuid.UUID) error {
	args := m.Called(ctx, editorFirebaseUID, postID)
	return args.Error(0)
}

// Default behavior for unexpected calls
func (m *MockNewsletterService) DefaultMockBehavior(methodName string, args ...interface{}) {
	m.T.Fatalf("Unexpected call to %s with args: %v", methodName, args)
}

// Remove panic-based stubs; rely on default behavior for unexpected calls
func (m *MockNewsletterService) GetPostsByNewsletterID(ctx context.Context, editorFirebaseUID string, newsletterID uuid.UUID, limit, offset int, statusFilter string) ([]models.Post, int, error) {
	panic("GetPostsByNewsletterID should not be called in publishing tests")
}
func (m *MockNewsletterService) UpdatePost(ctx context.Context, editorFirebaseUID string, postID uuid.UUID, title *string, content *string) (*models.Post, error) {
	panic("UpdatePost should not be called in publishing tests")
}
func (m *MockNewsletterService) DeletePost(ctx context.Context, editorFirebaseUID string, postID uuid.UUID) error {
	args := m.Called(ctx, editorFirebaseUID, postID)
	return args.Error(0)
}
func (m *MockNewsletterService) ListPostsByNewsletter(ctx context.Context, newsletterID uuid.UUID, limit int, offset int) ([]*models.Post, int, error) {
	args := m.Called(ctx, newsletterID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Post), args.Int(1), args.Error(2)
}

type MockSubscriberService struct {
	mock.Mock
}

func (m *MockSubscriberService) GetActiveSubscribersForNewsletter(ctx context.Context, newsletterID string) ([]models.Subscriber, error) {
	args := m.Called(ctx, newsletterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]models.Subscriber), args.Error(1)
}

// Add other methods from SubscriberServiceInterface to satisfy the interface
func (m *MockSubscriberService) SubscribeToNewsletter(ctx context.Context, req SubscribeToNewsletterRequest) (*SubscribeToNewsletterResponse, error) {
	panic("SubscribeToNewsletter should not be called in publishing tests")
}
func (m *MockSubscriberService) ConfirmSubscription(ctx context.Context, req ConfirmSubscriptionRequest) error {
	panic("ConfirmSubscription should not be called in publishing tests")
}
func (m *MockSubscriberService) UnsubscribeByToken(ctx context.Context, token string) error {
	panic("UnsubscribeByToken should not be called in publishing tests")
}
func (m *MockSubscriberService) UnsubscribeFromNewsletter(ctx context.Context, req UnsubscribeFromNewsletterRequest) error {
	panic("UnsubscribeFromNewsletter should not be called in publishing tests")
}

type MockEmailService struct {
	mock.Mock
}

func (m *MockEmailService) SendNewsletterIssue(toEmail, recipientName, subject, htmlContent, unsubscribeLink string) error {
	args := m.Called(toEmail, recipientName, subject, htmlContent, unsubscribeLink)
	return args.Error(0)
}

// Add other methods from email.EmailService to satisfy the interface
func (m *MockEmailService) SendConfirmationEmail(toEmail, recipientName, confirmationLink string) error {
	args := m.Called(toEmail, recipientName, confirmationLink)
	return args.Error(0)
}

func TestPublishPostToSubscribers_Success(t *testing.T) {
	ctx := context.Background()
	mockNewsletter := new(MockNewsletterService)
	mockSubscriber := new(MockSubscriberService)
	mockEmail := new(MockEmailService)

	postUUID := uuid.New()
	newsletterUUID := uuid.New()
	editorFirebaseUID := "editor-firebase-uid-123"

	testPost := testutils.CreateTestPost(newsletterUUID, 1)
	testPost.ID = postUUID
	testPost.NewsletterID = newsletterUUID

	// Convert to slice of values instead of pointers
	testSubscriberPtrs := []*models.Subscriber{
		testutils.CreateTestSubscriber(newsletterUUID.String(), 1),
		testutils.CreateTestSubscriber(newsletterUUID.String(), 2),
	}
	testSubscriberPtrs[0].Email = "sub1@example.com"
	testSubscriberPtrs[0].UnsubscribeToken = "unsub1"
	testSubscriberPtrs[1].Email = "sub2@example.com"
	testSubscriberPtrs[1].UnsubscribeToken = "unsub2"
	
	// Convert to slice of values for the mock
	testSubscribers := []models.Subscriber{
		*testSubscriberPtrs[0],
		*testSubscriberPtrs[1],
	}

	mockNewsletter.On("GetPostForPublishing", ctx, postUUID, editorFirebaseUID).Return(testPost, nil)
	mockSubscriber.On("GetActiveSubscribersForNewsletter", ctx, newsletterUUID.String()).Return(testSubscribers, nil)
	mockEmail.On("SendNewsletterIssue", testSubscriberPtrs[0].Email, testSubscriberPtrs[0].Email, testPost.Title, testPost.Content, mock.AnythingOfType("string")).Return(nil)
	mockEmail.On("SendNewsletterIssue", testSubscriberPtrs[1].Email, testSubscriberPtrs[1].Email, testPost.Title, testPost.Content, mock.AnythingOfType("string")).Return(nil)
	mockNewsletter.On("MarkPostAsPublished", ctx, editorFirebaseUID, postUUID).Return(nil)

	service := NewPublishingService(mockNewsletter, mockSubscriber, mockEmail)
	err := service.PublishPostToSubscribers(ctx, postUUID.String(), editorFirebaseUID)

	assert.NoError(t, err)
	mockNewsletter.AssertExpectations(t)
	mockSubscriber.AssertExpectations(t)
	mockEmail.AssertExpectations(t)
}

func TestPublishPostToSubscribers_NoSubscribers(t *testing.T) {
	ctx := context.Background()
	mockNewsletter := new(MockNewsletterService)
	mockSubscriber := new(MockSubscriberService)
	mockEmail := new(MockEmailService)

	postUUID := uuid.New()
	newsletterUUID := uuid.New()
	editorFirebaseUID := "editor-firebase-uid-123"

	testPost := testutils.CreateTestPost(newsletterUUID, 1)
	testPost.ID = postUUID
	testPost.NewsletterID = newsletterUUID

	var noSubscribers []models.Subscriber

	mockNewsletter.On("GetPostForPublishing", ctx, postUUID, editorFirebaseUID).Return(testPost, nil)
	mockSubscriber.On("GetActiveSubscribersForNewsletter", ctx, newsletterUUID.String()).Return(noSubscribers, nil)
	mockNewsletter.On("MarkPostAsPublished", ctx, editorFirebaseUID, postUUID).Return(nil)

	service := NewPublishingService(mockNewsletter, mockSubscriber, mockEmail)
	err := service.PublishPostToSubscribers(ctx, postUUID.String(), editorFirebaseUID)

	assert.NoError(t, err)
	mockNewsletter.AssertExpectations(t)
	mockSubscriber.AssertExpectations(t)
	mockEmail.AssertNotCalled(t, "SendNewsletterIssue", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
}

func TestPublishPostToSubscribers_PostNotFound(t *testing.T) {
	ctx := context.Background()
	mockNewsletter := new(MockNewsletterService)
	mockSubscriber := new(MockSubscriberService)
	mockEmail := new(MockEmailService)

	nonExistentPostUUID := uuid.New()
	editorFirebaseUID := "editor-firebase-uid-123"
	expectedError := errors.New("post not found")

	mockNewsletter.On("GetPostForPublishing", ctx, nonExistentPostUUID, editorFirebaseUID).Return(nil, expectedError)

	service := NewPublishingService(mockNewsletter, mockSubscriber, mockEmail)
	err := service.PublishPostToSubscribers(ctx, nonExistentPostUUID.String(), editorFirebaseUID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError.Error())

	mockNewsletter.AssertExpectations(t)
	mockSubscriber.AssertNotCalled(t, "GetActiveSubscribersForNewsletter", mock.Anything, mock.Anything)
	mockEmail.AssertNotCalled(t, "SendNewsletterIssue", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockNewsletter.AssertNotCalled(t, "MarkPostAsPublished", mock.Anything, mock.Anything, mock.Anything)
}

func TestPublishPostToSubscribers_ErrorFetchingSubscribers(t *testing.T) {
	ctx := context.Background()
	mockNewsletter := new(MockNewsletterService)
	mockSubscriber := new(MockSubscriberService)
	mockEmail := new(MockEmailService)

	postUUID := uuid.New()
	newsletterUUID := uuid.New()
	editorFirebaseUID := "editor-firebase-uid-123"

	testPost := testutils.CreateTestPost(newsletterUUID, 1)
	testPost.ID = postUUID
	testPost.NewsletterID = newsletterUUID

	expectedError := errors.New("db error fetching subscribers")

	mockNewsletter.On("GetPostForPublishing", ctx, postUUID, editorFirebaseUID).Return(testPost, nil)
	mockSubscriber.On("GetActiveSubscribersForNewsletter", ctx, newsletterUUID.String()).Return(nil, expectedError)

	service := NewPublishingService(mockNewsletter, mockSubscriber, mockEmail)
	err := service.PublishPostToSubscribers(ctx, postUUID.String(), editorFirebaseUID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError.Error())

	mockNewsletter.AssertExpectations(t)
	mockSubscriber.AssertExpectations(t)
	mockEmail.AssertNotCalled(t, "SendNewsletterIssue", mock.Anything, mock.Anything, mock.Anything, mock.Anything, mock.Anything)
	mockNewsletter.AssertNotCalled(t, "MarkPostAsPublished", mock.Anything, mock.Anything, mock.Anything)
}

func TestPublishPostToSubscribers_ErrorMarkingPublished(t *testing.T) {
	ctx := context.Background()
	mockNewsletter := new(MockNewsletterService)
	mockSubscriber := new(MockSubscriberService)
	mockEmail := new(MockEmailService)

	postUUID := uuid.New()
	newsletterUUID := uuid.New()
	editorFirebaseUID := "editor-firebase-uid-123"

	testPost := testutils.CreateTestPost(newsletterUUID, 1)
	testPost.ID = postUUID
	testPost.NewsletterID = newsletterUUID

	testSubscriberPtr := testutils.CreateTestSubscriber(newsletterUUID.String(), 1)
	testSubscriberPtr.Email = "sub1@example.com"
	testSubscriberPtr.UnsubscribeToken = "unsub1"
	
	testSubscribers := []models.Subscriber{*testSubscriberPtr}

	expectedError := errors.New("update failed")

	mockNewsletter.On("GetPostForPublishing", ctx, postUUID, editorFirebaseUID).Return(testPost, nil)
	mockSubscriber.On("GetActiveSubscribersForNewsletter", ctx, newsletterUUID.String()).Return(testSubscribers, nil)
	mockEmail.On("SendNewsletterIssue", testSubscriberPtr.Email, testSubscriberPtr.Email, testPost.Title, testPost.Content, mock.AnythingOfType("string")).Return(nil)
	mockNewsletter.On("MarkPostAsPublished", ctx, editorFirebaseUID, postUUID).Return(expectedError)

	service := NewPublishingService(mockNewsletter, mockSubscriber, mockEmail)
	err := service.PublishPostToSubscribers(ctx, postUUID.String(), editorFirebaseUID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedError.Error())

	mockNewsletter.AssertExpectations(t)
	mockSubscriber.AssertExpectations(t)
	mockEmail.AssertExpectations(t)
}

func TestPublishPostToSubscribers_PartialEmailFailures(t *testing.T) {
	ctx := context.Background()
	mockNewsletter := new(MockNewsletterService)
	mockSubscriber := new(MockSubscriberService)
	mockEmail := new(MockEmailService)

	postUUID := uuid.New()
	newsletterUUID := uuid.New()
	editorFirebaseUID := "editor-firebase-uid-123"

	testPost := testutils.CreateTestPost(newsletterUUID, 1)
	testPost.ID = postUUID
	testPost.NewsletterID = newsletterUUID

	sub1 := testutils.CreateTestSubscriber(newsletterUUID.String(), 1)
	sub1.Email = "sub1@example.com"
	sub1.UnsubscribeToken = "unsub1"
	sub2 := testutils.CreateTestSubscriber(newsletterUUID.String(), 2)
	sub2.Email = "sub2@example.com"
	sub2.UnsubscribeToken = "unsub2"
	
	testSubscribers := []models.Subscriber{*sub1, *sub2}

	emailSendFailedError := errors.New("email send failed")

	mockNewsletter.On("GetPostForPublishing", ctx, postUUID, editorFirebaseUID).Return(testPost, nil)
	mockSubscriber.On("GetActiveSubscribersForNewsletter", ctx, newsletterUUID.String()).Return(testSubscribers, nil)
	mockEmail.On("SendNewsletterIssue", sub1.Email, sub1.Email, testPost.Title, testPost.Content, mock.AnythingOfType("string")).Return(nil)
	mockEmail.On("SendNewsletterIssue", sub2.Email, sub2.Email, testPost.Title, testPost.Content, mock.AnythingOfType("string")).Return(emailSendFailedError)
	mockNewsletter.On("MarkPostAsPublished", ctx, editorFirebaseUID, postUUID).Return(nil)

	service := NewPublishingService(mockNewsletter, mockSubscriber, mockEmail)
	err := service.PublishPostToSubscribers(ctx, postUUID.String(), editorFirebaseUID)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "email(s) failed to send")
	assert.Contains(t, err.Error(), "1 email(s) failed to send")

	mockNewsletter.AssertExpectations(t)
	mockSubscriber.AssertExpectations(t)
	mockEmail.AssertExpectations(t)
}

// TestPublishPostToSubscribers_ErrorCreatingPublishingRecord
// This test assumes there's a step to create a "publishing record" or similar,
// which is not explicitly in the RFC's provided PublishingService mock.
// If the actual service has such a step and it can fail, a test like this would be needed.
// For now, it's commented out as it might not apply directly to the provided structure.
/*
func TestPublishPostToSubscribers_ErrorCreatingPublishingRecord(t *testing.T) {
    // ... similar setup ...
    // mockPublishingRecordRepo.On("CreateRecord", mock.Anything).Return(errors.New("record creation failed"))
    // ...
    // assert.Error(t, err)
    // assert.EqualError(t, err, "record creation failed")
}
*/ 