package service

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockNewsletterRepository is a mock type for the NewsletterRepository type
type MockNewsletterRepository struct {
	mock.Mock
}

func (m *MockNewsletterRepository) GetNewsletterByNameAndEditorID(name, editorID string) (*repository.Newsletter, error) {
	args := m.Called(name, editorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Newsletter), args.Error(1)
}

func (m *MockNewsletterRepository) CreateNewsletter(editorID, name, description string) (*repository.Newsletter, error) {
	args := m.Called(editorID, name, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Newsletter), args.Error(1)
}

func (m *MockNewsletterRepository) UpdateNewsletter(newsletterID string, editorID string, name *string, description *string) (*repository.Newsletter, error) {
	args := m.Called(newsletterID, editorID, name, description)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Newsletter), args.Error(1)
}

func (m *MockNewsletterRepository) DeleteNewsletter(newsletterID string, editorID string) error {
	args := m.Called(newsletterID, editorID)
	return args.Error(0)
}

func (m *MockNewsletterRepository) GetNewsletterByID(newsletterID string) (*repository.Newsletter, error) {
	args := m.Called(newsletterID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Newsletter), args.Error(1)
}

func (m *MockNewsletterRepository) ListNewslettersByEditorID(editorID string, limit, offset int) ([]repository.Newsletter, int, error) {
	args := m.Called(editorID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]repository.Newsletter), args.Int(1), args.Error(2)
}

func (m *MockNewsletterRepository) GetNewsletterByIDAndEditorID(newsletterID string, editorID string) (*repository.Newsletter, error) {
	args := m.Called(newsletterID, editorID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Newsletter), args.Error(1)
}

// MockPostRepository is a mock type for the PostRepository type
type MockPostRepository struct {
	mock.Mock
}

func (m *MockPostRepository) CreatePost(ctx context.Context, post *models.Post) (uuid.UUID, error) {
	args := m.Called(ctx, post)
	return args.Get(0).(uuid.UUID), args.Error(1)
}

func (m *MockPostRepository) GetPostByID(ctx context.Context, postID uuid.UUID) (*models.Post, error) {
	args := m.Called(ctx, postID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.Post), args.Error(1)
}

func (m *MockPostRepository) UpdatePost(ctx context.Context, post *models.Post) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

func (m *MockPostRepository) DeletePost(ctx context.Context, postID uuid.UUID) error {
	args := m.Called(ctx, postID)
	return args.Error(0)
}

func (m *MockPostRepository) MarkPostAsPublished(ctx context.Context, postID uuid.UUID, publishedAt time.Time) error {
	args := m.Called(ctx, postID, publishedAt)
	return args.Error(0)
}

func (m *MockPostRepository) ListPostsByNewsletterID(ctx context.Context, newsletterID uuid.UUID, limit, offset int) ([]*models.Post, int, error) {
	args := m.Called(ctx, newsletterID, limit, offset)
	if args.Get(0) == nil {
		return nil, args.Int(1), args.Error(2)
	}
	return args.Get(0).([]*models.Post), args.Int(1), args.Error(2)
}

// MockEditorRepository is defined in editor_test.go

// Test Scenarios
func TestCreateNewsletter_Success(t *testing.T) {
	ctx := context.Background()
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockPostRepo := new(MockPostRepository)
	mockEditorRepo := new(MockEditorRepository)

	newsletterService := NewNewsletterService(mockNewsletterRepo, mockPostRepo, mockEditorRepo)

	editorID := "editor-id-123"
	name := "My Test Newsletter"
	description := "A test newsletter."
	expectedNewsletter := &repository.Newsletter{ID: "nl-id-456", EditorID: editorID, Name: name, Description: description}

	mockNewsletterRepo.On("GetNewsletterByNameAndEditorID", name, editorID).Return(nil, nil)
	mockNewsletterRepo.On("CreateNewsletter", editorID, name, description).Return(expectedNewsletter, nil)

	createdNewsletter, err := newsletterService.CreateNewsletter(ctx, editorID, name, description)

	assert.NoError(t, err)
	assert.NotNil(t, createdNewsletter)
	assert.Equal(t, expectedNewsletter.ID, createdNewsletter.ID)
	mockNewsletterRepo.AssertExpectations(t)
}

func TestCreateNewsletter_DuplicateName(t *testing.T) {
	ctx := context.Background()
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockPostRepo := new(MockPostRepository)
	mockEditorRepo := new(MockEditorRepository)

	newsletterService := NewNewsletterService(mockNewsletterRepo, mockPostRepo, mockEditorRepo)

	editorID := "editor-id-123"
	name := "My Test Newsletter"
	description := "A test newsletter."
	existingNewsletter := &repository.Newsletter{ID: "nl-id-000", EditorID: editorID, Name: name, Description: "Old one"}

	mockNewsletterRepo.On("GetNewsletterByNameAndEditorID", name, editorID).Return(existingNewsletter, nil)

	createdNewsletter, err := newsletterService.CreateNewsletter(ctx, editorID, name, description)

	assert.Error(t, err)
	assert.Nil(t, createdNewsletter)
	assert.EqualError(t, err, ErrNewsletterNameTaken.Error())
	mockNewsletterRepo.AssertExpectations(t)
}

func TestUpdateNewsletter_Success(t *testing.T) {
	ctx := context.Background()
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockPostRepo := new(MockPostRepository)
	mockEditorRepo := new(MockEditorRepository)

	newsletterService := NewNewsletterService(mockNewsletterRepo, mockPostRepo, mockEditorRepo)

	newsletterID := "nl-id-789"
	editorID := "editor-id-123"
	newName := "Updated Newsletter Name"
	newDescription := "Updated description."
	updatedNewsletter := &repository.Newsletter{ID: newsletterID, EditorID: editorID, Name: newName, Description: newDescription}

	mockNewsletterRepo.On("UpdateNewsletter", newsletterID, editorID, &newName, &newDescription).Return(updatedNewsletter, nil)

	result, err := newsletterService.UpdateNewsletter(ctx, newsletterID, editorID, &newName, &newDescription)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, newName, result.Name)
	assert.Equal(t, newDescription, result.Description)
	mockNewsletterRepo.AssertExpectations(t)
}

func TestDeleteNewsletter_Success(t *testing.T) {
	ctx := context.Background()
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockPostRepo := new(MockPostRepository)
	mockEditorRepo := new(MockEditorRepository)

	newsletterService := NewNewsletterService(mockNewsletterRepo, mockPostRepo, mockEditorRepo)

	newsletterID := "nl-id-789"
	editorID := "editor-id-123"

	mockNewsletterRepo.On("DeleteNewsletter", newsletterID, editorID).Return(nil)

	err := newsletterService.DeleteNewsletter(ctx, newsletterID, editorID)

	assert.NoError(t, err)
	mockNewsletterRepo.AssertExpectations(t)
}

func TestListNewsletters_Success(t *testing.T) {
	ctx := context.Background()
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockPostRepo := new(MockPostRepository)
	mockEditorRepo := new(MockEditorRepository)

	newsletterService := NewNewsletterService(mockNewsletterRepo, mockPostRepo, mockEditorRepo)

	editorID := "editor-id-123"
	limit := 10
	offset := 0
	expectedNewsletters := []repository.Newsletter{
		{ID: "nl-1", EditorID: editorID, Name: "Newsletter 1", Description: "Desc 1"},
		{ID: "nl-2", EditorID: editorID, Name: "Newsletter 2", Description: "Desc 2"},
	}
	totalCount := 2

	mockNewsletterRepo.On("ListNewslettersByEditorID", editorID, limit, offset).Return(expectedNewsletters, totalCount, nil)

	newsletters, count, err := newsletterService.ListNewsletters(ctx, editorID, limit, offset)

	assert.NoError(t, err)
	assert.Equal(t, expectedNewsletters, newsletters)
	assert.Equal(t, totalCount, count)
	mockNewsletterRepo.AssertExpectations(t)
}
