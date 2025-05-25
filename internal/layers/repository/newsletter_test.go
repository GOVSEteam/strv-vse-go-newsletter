package repository_test

import (
	"context"
	"database/sql"
	"testing"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to create an editor for newsletter tests, as newsletters are tied to editors.
func createTestEditorForNewsletterTests(t *testing.T, ctx context.Context, repo repository.EditorRepository) *repository.Editor {
	testEditor := testutils.CreateTestEditor(0) // Index 0 for generic editor in this context
	createdEditor, err := repo.InsertEditor(testEditor.FirebaseUID, testEditor.Email)
	require.NoError(t, err)
	require.NotNil(t, createdEditor)
	return createdEditor
}

func TestCreateNewsletter_Success(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()

	editorRepo := repository.EditorRepo(suite.DB)
	newsletterRepo := repository.NewsletterRepo(suite.DB)

	// Newsletters require an editor, so create one first.
	ownerEditor := createTestEditorForNewsletterTests(t, ctx, editorRepo)

	name := "Test Newsletter " + uuid.New().String()
	description := "This is a test newsletter description."

	createdNewsletter, err := newsletterRepo.CreateNewsletter(ownerEditor.ID, name, description)

	require.NoError(t, err)
	require.NotNil(t, createdNewsletter)
	assert.NotEmpty(t, createdNewsletter.ID)
	assert.Equal(t, ownerEditor.ID, createdNewsletter.EditorID)
	assert.Equal(t, name, createdNewsletter.Name)
	assert.Equal(t, description, createdNewsletter.Description)
	assert.NotZero(t, createdNewsletter.CreatedAt)
	assert.NotZero(t, createdNewsletter.UpdatedAt)

	// Verify by fetching
	fetched, err := newsletterRepo.GetNewsletterByID(createdNewsletter.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, createdNewsletter.Name, fetched.Name)
}

func TestCreateNewsletter_DuplicateNameForSameEditor(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	editorRepo := repository.EditorRepo(suite.DB)
	newsletterRepo := repository.NewsletterRepo(suite.DB)
	ownerEditor := createTestEditorForNewsletterTests(t, ctx, editorRepo)

	name := "Duplicate Name Test Newsletter " + uuid.New().String()
	description := "First description."

	// Create first newsletter
	_, err := newsletterRepo.CreateNewsletter(ownerEditor.ID, name, description)
	require.NoError(t, err)

	// Try to create another with the same name and editor
	_, err = newsletterRepo.CreateNewsletter(ownerEditor.ID, name, "Second description")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "violates unique constraint")
	assert.Contains(t, err.Error(), "newsletters_editor_id_name_key") // Assuming this is the constraint name
}

func TestCreateNewsletter_SameNameDifferentEditor_Success(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	editorRepo := repository.EditorRepo(suite.DB)
	newsletterRepo := repository.NewsletterRepo(suite.DB)

	editor1 := createTestEditorForNewsletterTests(t, ctx, editorRepo)
	editor2 := createTestEditorForNewsletterTests(t, ctx, editorRepo)

	name := "Shared Name Test Newsletter " + uuid.New().String()

	// Editor 1 creates newsletter
	_, err := newsletterRepo.CreateNewsletter(editor1.ID, name, "Editor 1 version")
	require.NoError(t, err)

	// Editor 2 creates newsletter with the same name
	createdForEditor2, err := newsletterRepo.CreateNewsletter(editor2.ID, name, "Editor 2 version")

	require.NoError(t, err)
	assert.NotNil(t, createdForEditor2)
	assert.Equal(t, editor2.ID, createdForEditor2.EditorID)
	assert.Equal(t, name, createdForEditor2.Name)
}

func TestGetNewsletterByID_Success(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	editorRepo := repository.EditorRepo(suite.DB)
	newsletterRepo := repository.NewsletterRepo(suite.DB)
	ownerEditor := createTestEditorForNewsletterTests(t, ctx, editorRepo)

	name := "Newsletter to Get " + uuid.New().String()
	created, err := newsletterRepo.CreateNewsletter(ownerEditor.ID, name, "Desc")
	require.NoError(t, err)

	fetched, err := newsletterRepo.GetNewsletterByID(created.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, name, fetched.Name)
}

func TestGetNewsletterByID_NotFound(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	newsletterRepo := repository.NewsletterRepo(suite.DB)

	nonExistentID := uuid.New().String()
	fetched, err := newsletterRepo.GetNewsletterByID(nonExistentID)

	// Repository returns nil, nil for not found cases
	assert.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestListNewslettersByEditorID_Success(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	editorRepo := repository.EditorRepo(suite.DB)
	newsletterRepo := repository.NewsletterRepo(suite.DB)

	editor1 := createTestEditorForNewsletterTests(t, ctx, editorRepo)
	editor2 := createTestEditorForNewsletterTests(t, ctx, editorRepo)

	// Create 2 newsletters for editor1
	_, err := newsletterRepo.CreateNewsletter(editor1.ID, "E1 NL1 "+uuid.New().String(), "Desc1")
	require.NoError(t, err)
	_, err = newsletterRepo.CreateNewsletter(editor1.ID, "E1 NL2 "+uuid.New().String(), "Desc2")
	require.NoError(t, err)

	// Create 1 newsletter for editor2
	_, err = newsletterRepo.CreateNewsletter(editor2.ID, "E2 NL1 "+uuid.New().String(), "Desc3")
	require.NoError(t, err)

	// List for editor1
	newslettersE1, totalE1, err := newsletterRepo.ListNewslettersByEditorID(editor1.ID, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, totalE1)
	assert.Len(t, newslettersE1, 2)
	for _, nl := range newslettersE1 {
		assert.Equal(t, editor1.ID, nl.EditorID)
	}

	// List for editor2
	newslettersE2, totalE2, err := newsletterRepo.ListNewslettersByEditorID(editor2.ID, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 1, totalE2)
	assert.Len(t, newslettersE2, 1)
	assert.Equal(t, editor2.ID, newslettersE2[0].EditorID)

	// List for editor1 with pagination
	paginatedNLE1, totalPaginatedE1, err := newsletterRepo.ListNewslettersByEditorID(editor1.ID, 1, 0)
	require.NoError(t, err)
	assert.Equal(t, 2, totalPaginatedE1)
	assert.Len(t, paginatedNLE1, 1)
}

func TestListNewslettersByEditorID_NoNewsletters(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	editorRepo := repository.EditorRepo(suite.DB)
	newsletterRepo := repository.NewsletterRepo(suite.DB)

	editorWithNoNewsletters := createTestEditorForNewsletterTests(t, ctx, editorRepo)

	newsletters, total, err := newsletterRepo.ListNewslettersByEditorID(editorWithNoNewsletters.ID, 10, 0)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
	assert.Empty(t, newsletters)
}

func TestUpdateNewsletter_Success(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	editorRepo := repository.EditorRepo(suite.DB)
	newsletterRepo := repository.NewsletterRepo(suite.DB)
	ownerEditor := createTestEditorForNewsletterTests(t, ctx, editorRepo)

	originalName := "Original Name " + uuid.New().String()
	created, err := newsletterRepo.CreateNewsletter(ownerEditor.ID, originalName, "Original Desc")
	require.NoError(t, err)

	newName := "Updated Name " + uuid.New().String()
	newDesc := "Updated Description"

	updated, err := newsletterRepo.UpdateNewsletter(created.ID, ownerEditor.ID, &newName, &newDesc)
	require.NoError(t, err)
	require.NotNil(t, updated)
	assert.Equal(t, created.ID, updated.ID)
	assert.Equal(t, newName, updated.Name)
	assert.Equal(t, newDesc, updated.Description)
	assert.True(t, updated.UpdatedAt.After(created.UpdatedAt))

	// Verify by fetching
	fetched, err := newsletterRepo.GetNewsletterByID(created.ID)
	require.NoError(t, err)
	assert.Equal(t, newName, fetched.Name)
	assert.Equal(t, newDesc, fetched.Description)
}

func TestUpdateNewsletter_OnlyName(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	editorRepo := repository.EditorRepo(suite.DB)
	newsletterRepo := repository.NewsletterRepo(suite.DB)
	ownerEditor := createTestEditorForNewsletterTests(t, ctx, editorRepo)

	originalDesc := "Original Desc Only Name Update"
	created, err := newsletterRepo.CreateNewsletter(ownerEditor.ID, "Original Name ONNU "+uuid.New().String(), originalDesc)
	require.NoError(t, err)

	newName := "Updated Name ONNU " + uuid.New().String()
	updated, err := newsletterRepo.UpdateNewsletter(created.ID, ownerEditor.ID, &newName, nil) // No description update
	require.NoError(t, err)
	assert.Equal(t, newName, updated.Name)
	assert.Equal(t, originalDesc, updated.Description) // Description should remain unchanged
}

func TestUpdateNewsletter_OnlyDescription(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	editorRepo := repository.EditorRepo(suite.DB)
	newsletterRepo := repository.NewsletterRepo(suite.DB)
	ownerEditor := createTestEditorForNewsletterTests(t, ctx, editorRepo)

	originalName := "Original Name Only Desc Update " + uuid.New().String()
	created, err := newsletterRepo.CreateNewsletter(ownerEditor.ID, originalName, "Original Desc")
	require.NoError(t, err)

	newDesc := "Updated Description Only"
	updated, err := newsletterRepo.UpdateNewsletter(created.ID, ownerEditor.ID, nil, &newDesc) // No name update
	require.NoError(t, err)
	assert.Equal(t, originalName, updated.Name) // Name should remain unchanged
	assert.Equal(t, newDesc, updated.Description)
}

func TestUpdateNewsletter_NotFound(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	editorRepo := repository.EditorRepo(suite.DB)
	newsletterRepo := repository.NewsletterRepo(suite.DB)
	ownerEditor := createTestEditorForNewsletterTests(t, ctx, editorRepo)

	nonExistentID := uuid.New().String()
	nameUpdate := "Updated Name"
	_, err := newsletterRepo.UpdateNewsletter(nonExistentID, ownerEditor.ID, &nameUpdate, nil)

	// Repository returns sql.ErrNoRows for not found cases
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestUpdateNewsletter_NotOwner(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	editorRepo := repository.EditorRepo(suite.DB)
	newsletterRepo := repository.NewsletterRepo(suite.DB)

	actualOwner := createTestEditorForNewsletterTests(t, ctx, editorRepo)
	attackerEditor := createTestEditorForNewsletterTests(t, ctx, editorRepo)

	created, err := newsletterRepo.CreateNewsletter(actualOwner.ID, "Owned NL "+uuid.New().String(), "Desc")
	require.NoError(t, err)

	nameUpdate := "Hacked Name"
	_, err = newsletterRepo.UpdateNewsletter(created.ID, attackerEditor.ID, &nameUpdate, nil)

	// Repository returns nil, nil for not found/forbidden cases
	assert.NoError(t, err)
}

func TestDeleteNewsletter_Success(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	editorRepo := repository.EditorRepo(suite.DB)
	newsletterRepo := repository.NewsletterRepo(suite.DB)
	ownerEditor := createTestEditorForNewsletterTests(t, ctx, editorRepo)

	created, err := newsletterRepo.CreateNewsletter(ownerEditor.ID, "NL to Delete "+uuid.New().String(), "Desc")
	require.NoError(t, err)

	err = newsletterRepo.DeleteNewsletter(created.ID, ownerEditor.ID)
	require.NoError(t, err)

	// Verify it's gone - repository returns nil, nil for not found
	fetched, err := newsletterRepo.GetNewsletterByID(created.ID)
	assert.NoError(t, err)
	assert.Nil(t, fetched)
}

func TestDeleteNewsletter_NotFound(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	editorRepo := repository.EditorRepo(suite.DB)
	newsletterRepo := repository.NewsletterRepo(suite.DB)
	ownerEditor := createTestEditorForNewsletterTests(t, ctx, editorRepo)

	nonExistentID := uuid.New().String()
	err := newsletterRepo.DeleteNewsletter(nonExistentID, ownerEditor.ID)

	// Repository returns sql.ErrNoRows for not found cases
	assert.Equal(t, sql.ErrNoRows, err)
}

func TestDeleteNewsletter_NotOwner(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	editorRepo := repository.EditorRepo(suite.DB)
	newsletterRepo := repository.NewsletterRepo(suite.DB)

	actualOwner := createTestEditorForNewsletterTests(t, ctx, editorRepo)
	attackerEditor := createTestEditorForNewsletterTests(t, ctx, editorRepo)

	created, err := newsletterRepo.CreateNewsletter(actualOwner.ID, "Owned NL for Delete Test "+uuid.New().String(), "Desc")
	require.NoError(t, err)

	err = newsletterRepo.DeleteNewsletter(created.ID, attackerEditor.ID)
	// Repository returns sql.ErrNoRows for not found/forbidden cases
	assert.Equal(t, sql.ErrNoRows, err)

	// Verify it's still there (owned by actualOwner)
	fetched, err := newsletterRepo.GetNewsletterByID(created.ID)
	require.NoError(t, err)
	assert.NotNil(t, fetched)
}

func TestGetNewsletterByNameAndEditorID_Success(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	editorRepo := repository.EditorRepo(suite.DB)
	newsletterRepo := repository.NewsletterRepo(suite.DB)
	ownerEditor := createTestEditorForNewsletterTests(t, ctx, editorRepo)

	name := "Specific Name For Lookup " + uuid.New().String()
	created, err := newsletterRepo.CreateNewsletter(ownerEditor.ID, name, "Desc")
	require.NoError(t, err)

	fetched, err := newsletterRepo.GetNewsletterByNameAndEditorID(name, ownerEditor.ID)
	require.NoError(t, err)
	require.NotNil(t, fetched)
	assert.Equal(t, created.ID, fetched.ID)
	assert.Equal(t, name, fetched.Name)
}

func TestGetNewsletterByNameAndEditorID_NotFound(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	editorRepo := repository.EditorRepo(suite.DB)
	newsletterRepo := repository.NewsletterRepo(suite.DB)
	ownerEditor := createTestEditorForNewsletterTests(t, ctx, editorRepo)

	name := "NonExistent Name For Lookup " + uuid.New().String()
	fetched, err := newsletterRepo.GetNewsletterByNameAndEditorID(name, ownerEditor.ID)
	// Repository returns nil, nil for not found cases
	assert.NoError(t, err)
	assert.Nil(t, fetched)
} 