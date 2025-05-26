package repository_test

import (
	"context"
	"database/sql"
	"math/rand"
	"testing"
	"time"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create an editor for newsletter tests, as newsletters are tied to editors.
func createTestEditorForPostTests(t *testing.T, ctx context.Context, repo repository.EditorRepository) *repository.Editor {
	// Use a unique suffix based on current time and random component to avoid conflicts
	uniqueSuffix := int(time.Now().UnixNano()%10000) + rand.Intn(1000)
	testEditor := testutils.CreateTestEditor(uniqueSuffix)
	createdEditor, err := repo.InsertEditor(testEditor.FirebaseUID, testEditor.Email)
	require.NoError(t, err)
	require.NotNil(t, createdEditor)
	return createdEditor
}

// Helper to create an editor and a newsletter for post tests
func createEditorAndNewsletterForPostTests(t *testing.T, ctx context.Context, suite *testutils.TestSuite) (editorID string, newsletterID uuid.UUID) {
	editorRepo := repository.EditorRepo(suite.DB)
	newsletterRepo := repository.NewsletterRepo(suite.DB)

	// Use a unique suffix based on current time and random component to avoid conflicts
	uniqueSuffix := int(time.Now().UnixNano()%10000) + rand.Intn(1000)
	testEditor := testutils.CreateTestEditor(uniqueSuffix)
	createdEditor, err := editorRepo.InsertEditor(testEditor.FirebaseUID, testEditor.Email)
	require.NoError(t, err)
	editorID = createdEditor.ID

	testNewsletter, err := newsletterRepo.CreateNewsletter(editorID, "Post Test NL "+uuid.New().String(), "NL for posts")
	require.NoError(t, err)
	newsletterID, err = uuid.Parse(testNewsletter.ID)
	require.NoError(t, err)
	return
}

func TestCreatePost_Success(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	postRepo := repository.NewPostRepository(suite.DB)

	_, newsletterUUID := createEditorAndNewsletterForPostTests(t, ctx, suite)

	postToCreate := &models.Post{
		NewsletterID: newsletterUUID,
		Title:        "My First Post " + uuid.New().String(),
		Content:      "This is the content of my first post.",
		// ID, CreatedAt, UpdatedAt will be set by repo or DB
	}

	createdPostID, err := postRepo.CreatePost(ctx, postToCreate)
	require.NoError(t, err)
	assert.NotEqual(t, uuid.Nil, createdPostID)

	// Verify by fetching
	fetchedPost, err := postRepo.GetPostByID(ctx, createdPostID)
	require.NoError(t, err)
	require.NotNil(t, fetchedPost)
	assert.Equal(t, createdPostID, fetchedPost.ID)
	assert.Equal(t, postToCreate.NewsletterID, fetchedPost.NewsletterID)
	assert.Equal(t, postToCreate.Title, fetchedPost.Title)
	assert.Equal(t, postToCreate.Content, fetchedPost.Content)
	assert.NotZero(t, fetchedPost.CreatedAt)
	assert.NotZero(t, fetchedPost.UpdatedAt)
	assert.Nil(t, fetchedPost.PublishedAt) // Should be nil initially
}

func TestGetPostByID_Success(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	postRepo := repository.NewPostRepository(suite.DB)

	_, newsletterUUID := createEditorAndNewsletterForPostTests(t, ctx, suite)
	postToCreate := &models.Post{NewsletterID: newsletterUUID, Title: "Gettable Post", Content: "Content"}
	createdPostID, err := postRepo.CreatePost(ctx, postToCreate)
	require.NoError(t, err)

	fetchedPost, err := postRepo.GetPostByID(ctx, createdPostID)
	require.NoError(t, err)
	require.NotNil(t, fetchedPost)
	assert.Equal(t, createdPostID, fetchedPost.ID)
}

func TestGetPostByID_NotFound(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	postRepo := repository.NewPostRepository(suite.DB)

	nonExistentID := uuid.New()
	fetchedPost, err := postRepo.GetPostByID(ctx, nonExistentID)

	// Repository returns nil, nil for not found cases
	assert.NoError(t, err)
	assert.Nil(t, fetchedPost)
}

func TestListPostsByNewsletterID_Success(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	postRepo := repository.NewPostRepository(suite.DB)

	_, nl1UUID := createEditorAndNewsletterForPostTests(t, ctx, suite)
	_, nl2UUID := createEditorAndNewsletterForPostTests(t, ctx, suite)

	// Create 2 posts for nl1UUID
	_, err := postRepo.CreatePost(ctx, &models.Post{NewsletterID: nl1UUID, Title: "NL1 Post1", Content: "C1"})
	require.NoError(t, err)
	_, err = postRepo.CreatePost(ctx, &models.Post{NewsletterID: nl1UUID, Title: "NL1 Post2", Content: "C2"})
	require.NoError(t, err)

	// Create 1 post for nl2UUID
	_, err = postRepo.CreatePost(ctx, &models.Post{NewsletterID: nl2UUID, Title: "NL2 Post1", Content: "C3"})
	require.NoError(t, err)

	// List for nl1UUID
	postsNL1, totalNL1, err := postRepo.ListPostsByNewsletterID(ctx, nl1UUID, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, totalNL1)
	assert.Len(t, postsNL1, 2)
	for _, p := range postsNL1 {
		assert.Equal(t, nl1UUID, p.NewsletterID)
	}

	// List for nl2UUID
	postsNL2, totalNL2, err := postRepo.ListPostsByNewsletterID(ctx, nl2UUID, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, totalNL2)
	assert.Len(t, postsNL2, 1)
	assert.Equal(t, nl2UUID, postsNL2[0].NewsletterID)

	// List for nl1UUID with pagination
	paginatedPostsNL1, totalPaginatedNL1, err := postRepo.ListPostsByNewsletterID(ctx, nl1UUID, 1, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, totalPaginatedNL1)
	assert.Len(t, paginatedPostsNL1, 1)
}

func TestListPostsByNewsletterID_NoPosts(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	postRepo := repository.NewPostRepository(suite.DB)

	_, newsletterWithoutPostsUUID := createEditorAndNewsletterForPostTests(t, ctx, suite)

	posts, total, err := postRepo.ListPostsByNewsletterID(ctx, newsletterWithoutPostsUUID, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, posts)
}

func TestUpdatePost_Success(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	postRepo := repository.NewPostRepository(suite.DB)

	_, newsletterUUID := createEditorAndNewsletterForPostTests(t, ctx, suite)
	postToCreate := &models.Post{NewsletterID: newsletterUUID, Title: "Original Title", Content: "Original Content"}
	createdPostID, err := postRepo.CreatePost(ctx, postToCreate)
	require.NoError(t, err)

	newTitle := "Updated Post Title " + uuid.New().String()
	newContent := "Updated post content."

	postToUpdate := &models.Post{
		ID:           createdPostID,
		NewsletterID: newsletterUUID, // Usually not changed during update but good to have
		Title:        newTitle,
		Content:      newContent,
		// UpdatedAt will be set by the repo
	}

	err = postRepo.UpdatePost(ctx, postToUpdate)
	require.NoError(t, err)

	fetchedPost, err := postRepo.GetPostByID(ctx, createdPostID)
	require.NoError(t, err)
	assert.Equal(t, newTitle, fetchedPost.Title)
	assert.Equal(t, newContent, fetchedPost.Content)
	assert.True(t, fetchedPost.UpdatedAt.After(fetchedPost.CreatedAt)) // UpdatedAt should be more recent
}

func TestUpdatePost_NotFound(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	postRepo := repository.NewPostRepository(suite.DB)

	_, newsletterUUID := createEditorAndNewsletterForPostTests(t, ctx, suite) // Need a valid newsletter ID
	nonExistentPost := &models.Post{
		ID:           uuid.New(),
		NewsletterID: newsletterUUID,
		Title:        "Try to update non-existent",
		Content:      "Content",
	}
	err := postRepo.UpdatePost(ctx, nonExistentPost)
	// Repository returns sql.ErrNoRows for not found cases
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDeletePost_Success(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	postRepo := repository.NewPostRepository(suite.DB)

	_, newsletterUUID := createEditorAndNewsletterForPostTests(t, ctx, suite)
	postToCreate := &models.Post{NewsletterID: newsletterUUID, Title: "Post to Delete", Content: "Content"}
	createdPostID, err := postRepo.CreatePost(ctx, postToCreate)
	require.NoError(t, err)

	err = postRepo.DeletePost(ctx, createdPostID)
	require.NoError(t, err)

	// Verify it's gone - repository returns nil, nil for not found
	fetched, err := postRepo.GetPostByID(ctx, createdPostID)
	assert.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestDeletePost_NotFound(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	postRepo := repository.NewPostRepository(suite.DB)

	nonExistentID := uuid.New()
	err := postRepo.DeletePost(ctx, nonExistentID)

	// Repository returns sql.ErrNoRows for not found cases
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestMarkPostAsPublished_Success(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	postRepo := repository.NewPostRepository(suite.DB)

	_, newsletterUUID := createEditorAndNewsletterForPostTests(t, ctx, suite)
	postToCreate := &models.Post{NewsletterID: newsletterUUID, Title: "Post to Publish", Content: "Content"}
	createdPostID, err := postRepo.CreatePost(ctx, postToCreate)
	require.NoError(t, err)

	// Check it's not published yet
	initialPost, _ := postRepo.GetPostByID(ctx, createdPostID)
	require.Nil(t, initialPost.PublishedAt)

	publishTime := time.Now().UTC().Truncate(time.Second) // Truncate for easier comparison
	err = postRepo.MarkPostAsPublished(ctx, createdPostID, publishTime)
	require.NoError(t, err)

	publishedPost, err := postRepo.GetPostByID(ctx, createdPostID)
	require.NoError(t, err)
	require.NotNil(t, publishedPost.PublishedAt)
	assert.Equal(t, publishTime, publishedPost.PublishedAt.UTC().Truncate(time.Second))
}

func TestMarkPostAsPublished_NotFound(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	postRepo := repository.NewPostRepository(suite.DB)

	nonExistentID := uuid.New()
	err := postRepo.MarkPostAsPublished(ctx, nonExistentID, time.Now())
	// Repository returns sql.ErrNoRows for not found cases
	assert.Equal(t, sql.ErrNoRows, err)
} 