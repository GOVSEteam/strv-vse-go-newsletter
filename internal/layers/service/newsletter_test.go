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

// MockEditorRepositoryForNewsletterService is a mock for EditorRepository used by NewsletterService
// Re-using MockEditorRepository from editor_test.go might be possible if it's in the same package
// or making it public. For clarity, defining a specific one or ensuring access.
// Assuming MockEditorRepository from editor_test.go is accessible or we define a similar one here.
type MockEditorRepositoryForNWS struct { // Renamed to avoid conflict if in same package and unexported
	mock.Mock
}

func (m *MockEditorRepositoryForNWS) GetEditorByFirebaseUID(firebaseUID string) (*repository.Editor, error) {
	args := m.Called(firebaseUID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Editor), args.Error(1)
}

func (m *MockEditorRepositoryForNWS) InsertEditor(firebaseUID, email string) (*repository.Editor, error) {
	args := m.Called(firebaseUID, email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Editor), args.Error(1)
}

// Test Scenarios from RFC
func TestCreateNewsletter_Success(t *testing.T) {
	ctx := context.Background()
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockPostRepo := new(MockPostRepository)         // Added, as NewNewsletterService requires it
	mockEditorRepo := new(MockEditorRepositoryForNWS) // Added, as NewNewsletterService requires it

	newsletterService := NewNewsletterService(mockNewsletterRepo, mockPostRepo, mockEditorRepo)

	editorID := "editor-id-123"
	name := "My Test Newsletter"
	description := "A test newsletter."
	expectedNewsletter := &repository.Newsletter{ID: "nl-id-456", EditorID: editorID, Name: name, Description: description}

	mockNewsletterRepo.On("GetNewsletterByNameAndEditorID", name, editorID).Return(nil, nil) // No existing newsletter
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
	mockEditorRepo := new(MockEditorRepositoryForNWS)

	newsletterService := NewNewsletterService(mockNewsletterRepo, mockPostRepo, mockEditorRepo)

	editorID := "editor-id-123"
	name := "My Test Newsletter"
	description := "A test newsletter."
	existingNewsletter := &repository.Newsletter{ID: "nl-id-000", EditorID: editorID, Name: name, Description: "Old one"}

	mockNewsletterRepo.On("GetNewsletterByNameAndEditorID", name, editorID).Return(existingNewsletter, nil)

	createdNewsletter, err := newsletterService.CreateNewsletter(ctx, editorID, name, description)

	assert.Error(t, err)
	assert.Nil(t, createdNewsletter)
	assert.EqualError(t, err, ErrNewsletterNameTaken.Error()) // Using the exported error
	mockNewsletterRepo.AssertExpectations(t)
	mockNewsletterRepo.AssertNotCalled(t, "CreateNewsletter", mock.Anything, mock.Anything, mock.Anything)
}

func TestUpdateNewsletter_Success(t *testing.T) {
	ctx := context.Background()
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockPostRepo := new(MockPostRepository)
	mockEditorRepo := new(MockEditorRepositoryForNWS)

	newsletterService := NewNewsletterService(mockNewsletterRepo, mockPostRepo, mockEditorRepo)

	newsletterID := "nl-id-789"
	editorID := "editor-id-123"
	newName := "Updated Newsletter Name"
	newDescription := "Updated description."

	expectedUpdatedNewsletter := &repository.Newsletter{ID: newsletterID, EditorID: editorID, Name: newName, Description: newDescription}

	// Assume GetNewsletterByNameAndEditorID is called if name is updated, to check for duplicates.
	// For this success case, assume no duplicate with the new name (or it's the same newsletter ID).
	mockNewsletterRepo.On("GetNewsletterByNameAndEditorID", newName, editorID).Return(nil, nil) // No other newsletter with this new name
	mockNewsletterRepo.On("UpdateNewsletter", newsletterID, editorID, &newName, &newDescription).Return(expectedUpdatedNewsletter, nil)

	updatedNewsletter, err := newsletterService.UpdateNewsletter(ctx, newsletterID, editorID, &newName, &newDescription)

	assert.NoError(t, err)
	assert.NotNil(t, updatedNewsletter)
	assert.Equal(t, expectedUpdatedNewsletter.ID, updatedNewsletter.ID)
	assert.Equal(t, newName, updatedNewsletter.Name)
	mockNewsletterRepo.AssertExpectations(t)
}

func TestUpdateNewsletter_NotFound(t *testing.T) {
	ctx := context.Background()
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockPostRepo := new(MockPostRepository)
	mockEditorRepo := new(MockEditorRepositoryForNWS)
	newsletterService := NewNewsletterService(mockNewsletterRepo, mockPostRepo, mockEditorRepo)

	newsletterID := "nl-id-nonexistent"
	editorID := "editor-id-123"
	newName := "Updated Name"

	// If name is updated, GetNewsletterByNameAndEditorID is called first.
	mockNewsletterRepo.On("GetNewsletterByNameAndEditorID", newName, editorID).Return(nil, nil) // No duplicate for new name.
	// Then UpdateNewsletter is called, which would fail if the newsletterID doesn't exist or isn't owned by editorID.
	// The repository's UpdateNewsletter should handle this (e.g., return sql.ErrNoRows or a custom error).
	// The service layer might wrap this.
	// Assuming the repository UpdateNewsletter returns an error like sql.ErrNoRows which the service might propagate or wrap.
	mockNewsletterRepo.On("UpdateNewsletter", newsletterID, editorID, &newName, (*string)(nil)).Return(nil, sql.ErrNoRows) 

	updatedNewsletter, err := newsletterService.UpdateNewsletter(ctx, newsletterID, editorID, &newName, nil)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows) || err.Error() == sql.ErrNoRows.Error()) // Check if it's sql.ErrNoRows or a direct match
	assert.Nil(t, updatedNewsletter)
	mockNewsletterRepo.AssertExpectations(t)
}


func TestUpdateNewsletter_NotOwner(t *testing.T) {
	ctx := context.Background()
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockPostRepo := new(MockPostRepository)
	mockEditorRepo := new(MockEditorRepositoryForNWS)
	newsletterService := NewNewsletterService(mockNewsletterRepo, mockPostRepo, mockEditorRepo)

	newsletterID := "nl-id-belongs-to-other"
	editorID := "editor-id-attacker"
	newName := "Updated Name"

	mockNewsletterRepo.On("GetNewsletterByNameAndEditorID", newName, editorID).Return(nil, nil) // No duplicate for new name.
	// UpdateNewsletter in repo should check ownership and fail.
	// Assuming repo.UpdateNewsletter returns a specific error for forbidden access or sql.ErrNoRows if it filters by editorID internally.
	// The current NewsletterService directly calls repo.UpdateNewsletter which includes editorID for ownership check.
	// So, if editorID doesn't match, repo.UpdateNewsletter should ideally return 0 rows affected / sql.ErrNoRows.
	mockNewsletterRepo.On("UpdateNewsletter", newsletterID, editorID, &newName, (*string)(nil)).Return(nil, sql.ErrNoRows) // Simulating repo finds no matching row for newsletterID+editorID

	updatedNewsletter, err := newsletterService.UpdateNewsletter(ctx, newsletterID, editorID, &newName, nil)

	assert.Error(t, err)
	// Depending on how the repo/service handles this, it might be ErrForbidden or simply sql.ErrNoRows if not found for that user.
	// Given the current service structure, it relies on the repository's behavior.
	// If the repo's UpdateNewsletter is `UPDATE ... WHERE id = ? AND editor_id = ?`, then sql.ErrNoRows is appropriate.
	assert.True(t, errors.Is(err, sql.ErrNoRows) || err.Error() == sql.ErrNoRows.Error())
	assert.Nil(t, updatedNewsletter)
	mockNewsletterRepo.AssertExpectations(t)
}

func TestDeleteNewsletter_Success(t *testing.T) {
	ctx := context.Background()
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockPostRepo := new(MockPostRepository)
	mockEditorRepo := new(MockEditorRepositoryForNWS)
	newsletterService := NewNewsletterService(mockNewsletterRepo, mockPostRepo, mockEditorRepo)

	newsletterID := "nl-id-todelete"
	editorID := "editor-id-owner"

	mockNewsletterRepo.On("DeleteNewsletter", newsletterID, editorID).Return(nil)

	err := newsletterService.DeleteNewsletter(ctx, newsletterID, editorID)

	assert.NoError(t, err)
	mockNewsletterRepo.AssertExpectations(t)
}

func TestDeleteNewsletter_NotFound(t *testing.T) {
	ctx := context.Background()
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockPostRepo := new(MockPostRepository)
	mockEditorRepo := new(MockEditorRepositoryForNWS)
	newsletterService := NewNewsletterService(mockNewsletterRepo, mockPostRepo, mockEditorRepo)

	newsletterID := "nl-id-nonexistent"
	editorID := "editor-id-owner"

	// DeleteNewsletter in repo should check ownership and fail if not found/not owned.
	mockNewsletterRepo.On("DeleteNewsletter", newsletterID, editorID).Return(sql.ErrNoRows)

	err := newsletterService.DeleteNewsletter(ctx, newsletterID, editorID)

	assert.Error(t, err)
	assert.True(t, errors.Is(err, sql.ErrNoRows) || err.Error() == sql.ErrNoRows.Error())
	mockNewsletterRepo.AssertExpectations(t)
}

// --- Post Method Tests ---

func TestCreatePost_Success(t *testing.T) {
	ctx := context.Background()
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockPostRepo := new(MockPostRepository)
	mockEditorRepo := new(MockEditorRepositoryForNWS)
	newsletterService := NewNewsletterService(mockNewsletterRepo, mockPostRepo, mockEditorRepo)

	editorFirebaseUID := "firebase-uid-editor-owns-newsletter"
	newsletterUUID := uuid.New()
	title := "My Awesome Post"
	content := "This is the content of the post."

	// Mock editor lookup for ownership check
	expectedEditor := &repository.Editor{ID: "db-editor-id-1", FirebaseUID: editorFirebaseUID}
	mockEditorRepo.On("GetEditorByFirebaseUID", editorFirebaseUID).Return(expectedEditor, nil)

	// Mock newsletter lookup for ownership check
	expectedNewsletter := &repository.Newsletter{ID: newsletterUUID.String(), EditorID: expectedEditor.ID} // Editor owns this newsletter
	mockNewsletterRepo.On("GetNewsletterByID", newsletterUUID.String()).Return(expectedNewsletter, nil)

	// Mock post creation
	createdPostID := uuid.New()
	mockPostRepo.On("CreatePost", ctx, mock.AnythingOfType("*models.Post")).Run(func(args mock.Arguments) {
		postArg := args.Get(1).(*models.Post)
		assert.Equal(t, newsletterUUID, postArg.NewsletterID)
		assert.Equal(t, title, postArg.Title)
		assert.Equal(t, content, postArg.Content)
	}).Return(createdPostID, nil)

	createdPost, err := newsletterService.CreatePost(ctx, editorFirebaseUID, newsletterUUID, title, content)

	assert.NoError(t, err)
	assert.NotNil(t, createdPost)
	assert.Equal(t, createdPostID, createdPost.ID)
	assert.Equal(t, title, createdPost.Title)
	mockEditorRepo.AssertExpectations(t)
	mockNewsletterRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
}


func TestCreatePost_Forbidden_EditorNotOwner(t *testing.T) {
	ctx := context.Background()
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockPostRepo := new(MockPostRepository)
	mockEditorRepo := new(MockEditorRepositoryForNWS)
	newsletterService := NewNewsletterService(mockNewsletterRepo, mockPostRepo, mockEditorRepo)

	editorFirebaseUID := "firebase-uid-editor-does-not-own"
	newsletterUUID := uuid.New()
	title := "Attempted Post"
	content := "Content"

	// Mock editor lookup
	requestingEditor := &repository.Editor{ID: "db-editor-id-2", FirebaseUID: editorFirebaseUID}
	mockEditorRepo.On("GetEditorByFirebaseUID", editorFirebaseUID).Return(requestingEditor, nil)

	// Mock newsletter lookup - newsletter exists but owned by someone else
	actualOwnerEditorID := "db-editor-id-OWNER"
	existingNewsletter := &repository.Newsletter{ID: newsletterUUID.String(), EditorID: actualOwnerEditorID}
	mockNewsletterRepo.On("GetNewsletterByID", newsletterUUID.String()).Return(existingNewsletter, nil)

	createdPost, err := newsletterService.CreatePost(ctx, editorFirebaseUID, newsletterUUID, title, content)

	assert.Error(t, err)
	assert.Nil(t, createdPost)
	assert.EqualError(t, err, ErrForbidden.Error())
	mockEditorRepo.AssertExpectations(t)
	mockNewsletterRepo.AssertExpectations(t)
	mockPostRepo.AssertNotCalled(t, "CreatePost", mock.Anything, mock.Anything)
}

func TestUpdatePost_Success(t *testing.T) {
	ctx := context.Background()
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockPostRepo := new(MockPostRepository)
	mockEditorRepo := new(MockEditorRepositoryForNWS)
	newsletterService := NewNewsletterService(mockNewsletterRepo, mockPostRepo, mockEditorRepo)

	editorFirebaseUID := "firebase-uid-editor-owns-post"
	postUUID := uuid.New()
	newsletterUUID := uuid.New() // Newsletter to which the post belongs
	newTitle := "Updated Post Title"
	newContent := "Updated post content."

	// Mock editor lookup for ownership check
	ownerEditor := &repository.Editor{ID: "db-editor-id-owner", FirebaseUID: editorFirebaseUID}
	mockEditorRepo.On("GetEditorByFirebaseUID", editorFirebaseUID).Return(ownerEditor, nil)

	// Mock GetPostByID to return the existing post
	existingPost := &models.Post{
		ID:           postUUID,
		NewsletterID: newsletterUUID,
		Title:        "Old Title",
		Content:      "Old Content",
	}
	mockPostRepo.On("GetPostByID", ctx, postUUID).Return(existingPost, nil)

	// Mock GetNewsletterByID for ownership check (ensure editor owns the newsletter of the post)
	ownedNewsletter := &repository.Newsletter{ID: newsletterUUID.String(), EditorID: ownerEditor.ID}
	mockNewsletterRepo.On("GetNewsletterByID", newsletterUUID.String()).Return(ownedNewsletter, nil)

	// Mock UpdatePost
	mockPostRepo.On("UpdatePost", ctx, mock.MatchedBy(func(post *models.Post) bool {
		return post.ID == postUUID && post.Title == newTitle && post.Content == newContent
	})).Return(nil)

	updatedPost, err := newsletterService.UpdatePost(ctx, editorFirebaseUID, postUUID, &newTitle, &newContent)

	assert.NoError(t, err)
	assert.NotNil(t, updatedPost)
	assert.Equal(t, postUUID, updatedPost.ID)
	assert.Equal(t, newTitle, updatedPost.Title)
	assert.Equal(t, newContent, updatedPost.Content)
	mockEditorRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
	mockNewsletterRepo.AssertExpectations(t)
}

func TestDeletePost_Success(t *testing.T) {
	ctx := context.Background()
	mockNewsletterRepo := new(MockNewsletterRepository)
	mockPostRepo := new(MockPostRepository)
	mockEditorRepo := new(MockEditorRepositoryForNWS)
	newsletterService := NewNewsletterService(mockNewsletterRepo, mockPostRepo, mockEditorRepo)

	editorFirebaseUID := "firebase-uid-editor-owns-post-for-delete"
	postUUID := uuid.New()
	newsletterUUID := uuid.New()

	ownerEditor := &repository.Editor{ID: "db-editor-id-owner-for-delete", FirebaseUID: editorFirebaseUID}
	mockEditorRepo.On("GetEditorByFirebaseUID", editorFirebaseUID).Return(ownerEditor, nil)

	existingPost := &models.Post{ID: postUUID, NewsletterID: newsletterUUID}
	mockPostRepo.On("GetPostByID", ctx, postUUID).Return(existingPost, nil)

	ownedNewsletter := &repository.Newsletter{ID: newsletterUUID.String(), EditorID: ownerEditor.ID}
	mockNewsletterRepo.On("GetNewsletterByID", newsletterUUID.String()).Return(ownedNewsletter, nil)

	mockPostRepo.On("DeletePost", ctx, postUUID).Return(nil)

	err := newsletterService.DeletePost(ctx, editorFirebaseUID, postUUID)

	assert.NoError(t, err)
	mockEditorRepo.AssertExpectations(t)
	mockPostRepo.AssertExpectations(t)
	mockNewsletterRepo.AssertExpectations(t)
}

// Placeholder for other test scenarios from RFC if any for NewsletterService
// func TestUpdateNewsletter_NotFound(t *testing.T) - Already added
// func TestUpdateNewsletter_NotOwner(t *testing.T) - Already added
// func TestDeleteNewsletter_NotFound(t *testing.T) - Already added
