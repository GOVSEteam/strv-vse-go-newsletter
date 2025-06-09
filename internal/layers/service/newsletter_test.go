package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
)

// MockNewsletterRepository mocks the newsletter repository
type MockNewsletterRepository struct {
	mock.Mock
}

func (m *MockNewsletterRepository) CreateNewsletter(ctx context.Context, editorID, name, description string) (*models.Newsletter, error) {
	args := m.Called(ctx, editorID, name, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Newsletter), args.Error(1)
}

func (m *MockNewsletterRepository) GetNewsletterByID(ctx context.Context, newsletterID string) (*models.Newsletter, error) {
	args := m.Called(ctx, newsletterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Newsletter), args.Error(1)
}

func (m *MockNewsletterRepository) ListNewslettersByEditorID(ctx context.Context, editorID string, limit int, offset int) ([]models.Newsletter, int, error) {
	args := m.Called(ctx, editorID, limit, offset)
	return args.Get(0).([]models.Newsletter), args.Get(1).(int), args.Error(2)
}

func (m *MockNewsletterRepository) GetNewsletterByIDAndEditorID(ctx context.Context, newsletterID string, editorID string) (*models.Newsletter, error) {
	args := m.Called(ctx, newsletterID, editorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Newsletter), args.Error(1)
}

func (m *MockNewsletterRepository) UpdateNewsletter(ctx context.Context, newsletterID string, editorID string, name *string, description *string) (*models.Newsletter, error) {
	args := m.Called(ctx, newsletterID, editorID, name, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Newsletter), args.Error(1)
}

func (m *MockNewsletterRepository) DeleteNewsletter(ctx context.Context, newsletterID string, editorID string) error {
	args := m.Called(ctx, newsletterID, editorID)
	return args.Error(0)
}

func (m *MockNewsletterRepository) GetNewsletterByNameAndEditorID(ctx context.Context, name string, editorID string) (*models.Newsletter, error) {
	args := m.Called(ctx, name, editorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Newsletter), args.Error(1)
}

// MockPostRepository mocks the post repository
type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) CreatePost(ctx context.Context, post *models.Post) (*models.Post, error) {
	args := m.Called(ctx, post)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostRepository) GetPostByID(ctx context.Context, postID string) (*models.Post, error) {
	args := m.Called(ctx, postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostRepository) ListPostsByNewsletterID(ctx context.Context, newsletterID string, limit int, offset int) ([]models.Post, int, error) {
	args := m.Called(ctx, newsletterID, limit, offset)
	return args.Get(0).([]models.Post), args.Get(1).(int), args.Error(2)
}

func (m *MockPostRepository) UpdatePost(ctx context.Context, postID string, updates repository.PostUpdate) (*models.Post, error) {
	args := m.Called(ctx, postID, updates)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostRepository) SetPostPublished(ctx context.Context, postID string, publishedAt time.Time) (*models.Post, error) {
	args := m.Called(ctx, postID, publishedAt)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostRepository) SetPostUnpublished(ctx context.Context, postID string) (*models.Post, error) {
	args := m.Called(ctx, postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostRepository) DeletePost(ctx context.Context, postID string) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

// MockSubscriberService mocks the subscriber service
type MockSubscriberService struct {
	mock.Mock
}

func (m *MockSubscriberService) SubscribeToNewsletter(ctx context.Context, email, newsletterID string) (*models.Subscriber, error) {
	args := m.Called(ctx, email, newsletterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Subscriber), args.Error(1)
}

func (m *MockSubscriberService) UnsubscribeByToken(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *MockSubscriberService) ListActiveSubscribersByNewsletterID(ctx context.Context, editorAuthID string, newsletterID string, limit, offset int) ([]models.Subscriber, int, error) {
	args := m.Called(ctx, editorAuthID, newsletterID, limit, offset)
	return args.Get(0).([]models.Subscriber), args.Get(1).(int), args.Error(2)
}

func (m *MockSubscriberService) GetActiveSubscribersForNewsletter(ctx context.Context, newsletterID string) ([]models.Subscriber, error) {
	args := m.Called(ctx, newsletterID)
	return args.Get(0).([]models.Subscriber), args.Error(1)
}

func (m *MockSubscriberService) DeleteAllSubscribersByNewsletterID(ctx context.Context, newsletterID string) error {
	args := m.Called(ctx, newsletterID)
	return args.Error(0)
}

func TestNewsletterService_CreateNewsletter(t *testing.T) {
	tests := []struct {
		name           string
		editorID       string
		newsletterName string
		description    string
		setupMocks     func(*MockNewsletterRepository, *MockPostRepository, *MockSubscriberService)
		setupContext   func() context.Context
		expectedError  string
		expectSuccess  bool
	}{
		{
			name:           "successful newsletter creation",
			editorID:       "editor_123",
			newsletterName: "Tech Weekly",
			description:    "Weekly tech updates",
			setupMocks: func(mockRepo *MockNewsletterRepository, mockPostRepo *MockPostRepository, mockSubService *MockSubscriberService) {
				expectedNewsletter := &models.Newsletter{
					ID:          "newsletter_456",
					EditorID:    "editor_123",
					Name:        "Tech Weekly",
					Description: "Weekly tech updates",
					CreatedAt:   time.Now(),
					UpdatedAt:   time.Now(),
				}
				mockRepo.On("CreateNewsletter", mock.Anything, "editor_123", "Tech Weekly", "Weekly tech updates").
					Return(expectedNewsletter, nil)
			},
			setupContext: func() context.Context {
				editor := &models.Editor{
					ID:          "editor_123",
					FirebaseUID: "firebase_456",
					Email:       "test@example.com",
				}
				ctx := context.WithValue(context.Background(), middleware.EditorContextKey, editor)
				return ctx
			},
			expectSuccess: true,
		},
		{
			name:           "empty newsletter name",
			editorID:       "editor_123",
			newsletterName: "",
			description:    "Some description",
			setupMocks: func(mockRepo *MockNewsletterRepository, mockPostRepo *MockPostRepository, mockSubService *MockSubscriberService) {
				// No repository calls expected due to validation failure
			},
			setupContext: func() context.Context {
				editor := &models.Editor{ID: "editor_123"}
				return context.WithValue(context.Background(), middleware.EditorContextKey, editor)
			},
			expectedError: "name cannot be empty",
			expectSuccess: false,
		},
		{
			name:           "newsletter name too long",
			editorID:       "editor_123",
			newsletterName: "This is a very long newsletter name that exceeds the maximum allowed length for a newsletter name in our system",
			description:    "Description",
			setupMocks: func(mockRepo *MockNewsletterRepository, mockPostRepo *MockPostRepository, mockSubService *MockSubscriberService) {
				// No repository calls expected due to validation failure
			},
			setupContext: func() context.Context {
				editor := &models.Editor{ID: "editor_123"}
				return context.WithValue(context.Background(), middleware.EditorContextKey, editor)
			},
			expectedError: "name exceeds max length",
			expectSuccess: false,
		},
		{
			name:           "no editor in context",
			editorID:       "editor_123",
			newsletterName: "Tech Weekly",
			description:    "Description",
			setupMocks: func(mockRepo *MockNewsletterRepository, mockPostRepo *MockPostRepository, mockSubService *MockSubscriberService) {
				// No repository calls expected due to authorization failure
			},
			setupContext: func() context.Context {
				return context.Background() // No editor in context
			},
			expectedError: "forbidden",
			expectSuccess: false,
		},
		{
			name:           "repository conflict error",
			editorID:       "editor_123",
			newsletterName: "Existing Newsletter",
			description:    "Description",
			setupMocks: func(mockRepo *MockNewsletterRepository, mockPostRepo *MockPostRepository, mockSubService *MockSubscriberService) {
				mockRepo.On("CreateNewsletter", mock.Anything, "editor_123", "Existing Newsletter", "Description").
					Return(nil, apperrors.ErrConflict)
			},
			setupContext: func() context.Context {
				editor := &models.Editor{ID: "editor_123"}
				return context.WithValue(context.Background(), middleware.EditorContextKey, editor)
			},
			expectedError: "already taken",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockNewsletterRepo := &MockNewsletterRepository{}
			mockPostRepo := &MockPostRepository{}
			mockSubService := &MockSubscriberService{}
			tt.setupMocks(mockNewsletterRepo, mockPostRepo, mockSubService)

			// Create service
			service := NewNewsletterService(mockNewsletterRepo, mockPostRepo, mockSubService)

			// Setup context
			ctx := tt.setupContext()

			// Execute
			result, err := service.CreateNewsletter(ctx, tt.editorID, tt.newsletterName, tt.description)

			// Verify
			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.newsletterName, result.Name)
				assert.Equal(t, tt.description, result.Description)
			} else {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedError)
			}

			// Verify mock expectations
			mockNewsletterRepo.AssertExpectations(t)
			mockPostRepo.AssertExpectations(t)
			mockSubService.AssertExpectations(t)
		})
	}
}

func TestNewsletterService_GetNewsletterForEditor(t *testing.T) {
	tests := []struct {
		name          string
		newsletterID  string
		setupMocks    func(*MockNewsletterRepository, *MockPostRepository, *MockSubscriberService)
		setupContext  func() context.Context
		expectedError string
		expectSuccess bool
	}{
		{
			name:         "successful get newsletter",
			newsletterID: "newsletter_123",
			setupMocks: func(mockRepo *MockNewsletterRepository, mockPostRepo *MockPostRepository, mockSubService *MockSubscriberService) {
				expectedNewsletter := &models.Newsletter{
					ID:       "newsletter_123",
					EditorID: "editor_456",
					Name:     "Tech Newsletter",
				}
				mockRepo.On("GetNewsletterByID", mock.Anything, "newsletter_123").
					Return(expectedNewsletter, nil)
			},
			setupContext: func() context.Context {
				editor := &models.Editor{ID: "editor_456"}
				return context.WithValue(context.Background(), middleware.EditorContextKey, editor)
			},
			expectSuccess: true,
		},
		{
			name:         "newsletter not found",
			newsletterID: "nonexistent_newsletter",
			setupMocks: func(mockRepo *MockNewsletterRepository, mockPostRepo *MockPostRepository, mockSubService *MockSubscriberService) {
				mockRepo.On("GetNewsletterByID", mock.Anything, "nonexistent_newsletter").
					Return(nil, apperrors.ErrNewsletterNotFound)
			},
			setupContext: func() context.Context {
				editor := &models.Editor{ID: "editor_456"}
				return context.WithValue(context.Background(), middleware.EditorContextKey, editor)
			},
			expectedError: "newsletter not found",
			expectSuccess: false,
		},
		{
			name:         "editor doesn't own newsletter",
			newsletterID: "newsletter_123",
			setupMocks: func(mockRepo *MockNewsletterRepository, mockPostRepo *MockPostRepository, mockSubService *MockSubscriberService) {
				newsletter := &models.Newsletter{
					ID:       "newsletter_123",
					EditorID: "different_editor", // Different owner
					Name:     "Someone Else's Newsletter",
				}
				mockRepo.On("GetNewsletterByID", mock.Anything, "newsletter_123").
					Return(newsletter, nil)
			},
			setupContext: func() context.Context {
				editor := &models.Editor{ID: "editor_456"}
				return context.WithValue(context.Background(), middleware.EditorContextKey, editor)
			},
			expectedError: "forbidden",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockNewsletterRepo := &MockNewsletterRepository{}
			mockPostRepo := &MockPostRepository{}
			mockSubService := &MockSubscriberService{}
			tt.setupMocks(mockNewsletterRepo, mockPostRepo, mockSubService)

			// Create service
			service := NewNewsletterService(mockNewsletterRepo, mockPostRepo, mockSubService)

			// Setup context
			ctx := tt.setupContext()

			// Execute
			result, err := service.GetNewsletterForEditor(ctx, "editor_456", tt.newsletterID)

			// Verify
			if tt.expectSuccess {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				assert.Equal(t, tt.newsletterID, result.ID)
			} else {
				assert.Error(t, err)
				assert.Nil(t, result)
				assert.Contains(t, err.Error(), tt.expectedError)
			}

			// Verify mock expectations
			mockNewsletterRepo.AssertExpectations(t)
			mockPostRepo.AssertExpectations(t)
			mockSubService.AssertExpectations(t)
		})
	}
}

func TestNewsletterService_DeleteNewsletter(t *testing.T) {
	tests := []struct {
		name          string
		newsletterID  string
		setupMocks    func(*MockNewsletterRepository, *MockPostRepository, *MockSubscriberService)
		setupContext  func() context.Context
		expectedError string
		expectSuccess bool
	}{
		{
			name:         "successful newsletter deletion with subscriber cleanup",
			newsletterID: "newsletter_123",
			setupMocks: func(mockRepo *MockNewsletterRepository, mockPostRepo *MockPostRepository, mockSubService *MockSubscriberService) {
				newsletter := &models.Newsletter{
					ID:       "newsletter_123",
					EditorID: "editor_456",
					Name:     "Tech Newsletter",
				}
				// First call to check existence and ownership
				mockRepo.On("GetNewsletterByID", mock.Anything, "newsletter_123").
					Return(newsletter, nil)
				// Subscriber cleanup should be called before deletion
				mockSubService.On("DeleteAllSubscribersByNewsletterID", mock.Anything, "newsletter_123").
					Return(nil)
				// Finally delete the newsletter
				mockRepo.On("DeleteNewsletter", mock.Anything, "newsletter_123", "editor_456").
					Return(nil)
			},
			setupContext: func() context.Context {
				editor := &models.Editor{ID: "editor_456"}
				return context.WithValue(context.Background(), middleware.EditorContextKey, editor)
			},
			expectSuccess: true,
		},
		{
			name:         "newsletter not found",
			newsletterID: "nonexistent_newsletter",
			setupMocks: func(mockRepo *MockNewsletterRepository, mockPostRepo *MockPostRepository, mockSubService *MockSubscriberService) {
				mockRepo.On("GetNewsletterByID", mock.Anything, "nonexistent_newsletter").
					Return(nil, apperrors.ErrNewsletterNotFound)
			},
			setupContext: func() context.Context {
				editor := &models.Editor{ID: "editor_456"}
				return context.WithValue(context.Background(), middleware.EditorContextKey, editor)
			},
			expectedError: "newsletter not found",
			expectSuccess: false,
		},
		{
			name:         "editor doesn't own newsletter",
			newsletterID: "newsletter_123",
			setupMocks: func(mockRepo *MockNewsletterRepository, mockPostRepo *MockPostRepository, mockSubService *MockSubscriberService) {
				newsletter := &models.Newsletter{
					ID:       "newsletter_123",
					EditorID: "different_editor", // Different owner
					Name:     "Someone Else's Newsletter",
				}
				mockRepo.On("GetNewsletterByID", mock.Anything, "newsletter_123").
					Return(newsletter, nil)
			},
			setupContext: func() context.Context {
				editor := &models.Editor{ID: "editor_456"}
				return context.WithValue(context.Background(), middleware.EditorContextKey, editor)
			},
			expectedError: "forbidden",
			expectSuccess: false,
		},
		{
			name:         "subscriber cleanup fails",
			newsletterID: "newsletter_123",
			setupMocks: func(mockRepo *MockNewsletterRepository, mockPostRepo *MockPostRepository, mockSubService *MockSubscriberService) {
				newsletter := &models.Newsletter{
					ID:       "newsletter_123",
					EditorID: "editor_456",
					Name:     "Tech Newsletter",
				}
				mockRepo.On("GetNewsletterByID", mock.Anything, "newsletter_123").
					Return(newsletter, nil)
				// Subscriber cleanup fails
				mockSubService.On("DeleteAllSubscribersByNewsletterID", mock.Anything, "newsletter_123").
					Return(apperrors.ErrInternal)
				// Newsletter deletion should not be called if subscriber cleanup fails
			},
			setupContext: func() context.Context {
				editor := &models.Editor{ID: "editor_456"}
				return context.WithValue(context.Background(), middleware.EditorContextKey, editor)
			},
			expectedError: "failed to cleanup subscribers",
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mocks
			mockNewsletterRepo := &MockNewsletterRepository{}
			mockPostRepo := &MockPostRepository{}
			mockSubService := &MockSubscriberService{}
			tt.setupMocks(mockNewsletterRepo, mockPostRepo, mockSubService)

			// Create service
			service := NewNewsletterService(mockNewsletterRepo, mockPostRepo, mockSubService)

			// Setup context
			ctx := tt.setupContext()

			// Execute
			err := service.DeleteNewsletter(ctx, "editor_456", tt.newsletterID)

			// Verify
			if tt.expectSuccess {
				assert.NoError(t, err)
			} else {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			}

			// Verify mock expectations
			mockNewsletterRepo.AssertExpectations(t)
			mockPostRepo.AssertExpectations(t)
			mockSubService.AssertExpectations(t)
		})
	}
} 