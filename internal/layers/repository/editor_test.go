package repository_test

import (
	"testing"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateEditor_Success(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)

	repo := repository.EditorRepo(suite.DB)

	testEditor := testutils.CreateTestEditor(1)

	createdEditor, err := repo.InsertEditor(testEditor.FirebaseUID, testEditor.Email)

	require.NoError(t, err)
	assert.NotEmpty(t, createdEditor.ID)
	assert.Equal(t, testEditor.Email, createdEditor.Email)
	assert.Equal(t, testEditor.FirebaseUID, createdEditor.FirebaseUID)

	retrievedEditor, err := repo.GetEditorByFirebaseUID(createdEditor.FirebaseUID)
	require.NoError(t, err)
	require.NotNil(t, retrievedEditor)
	assert.Equal(t, createdEditor.Email, retrievedEditor.Email)
	assert.Equal(t, createdEditor.FirebaseUID, retrievedEditor.FirebaseUID)
}

func TestCreateEditor_DuplicateEmail(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)

	repo := repository.EditorRepo(suite.DB)

	testEditor1 := testutils.CreateTestEditor(1)
	_, err := repo.InsertEditor(testEditor1.FirebaseUID, testEditor1.Email)
	require.NoError(t, err)

	testEditor2FirebaseUID := "firebase-uid-" + uuid.New().String()
	_, err = repo.InsertEditor(testEditor2FirebaseUID, testEditor1.Email)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "violates unique constraint")
	assert.Contains(t, err.Error(), "editors_email_key")
}

func TestCreateEditor_DuplicateFirebaseUID(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)

	repo := repository.EditorRepo(suite.DB)

	testEditor1 := testutils.CreateTestEditor(1)
	_, err := repo.InsertEditor(testEditor1.FirebaseUID, testEditor1.Email)
	require.NoError(t, err)

	differentEmail := "another-" + uuid.New().String() + "@example.com"
	_, err = repo.InsertEditor(testEditor1.FirebaseUID, differentEmail)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "violates unique constraint")
	assert.Contains(t, err.Error(), "editors_firebase_uid_key")
}

func TestGetEditorByID_Success(t *testing.T) {
	t.Skip("Skipping GetEditorByID tests as method is not on EditorRepository interface")
}

func TestGetEditorByID_NotFound(t *testing.T) {
	t.Skip("Skipping GetEditorByID tests as method is not on EditorRepository interface")
}

func TestGetEditorByEmail_Success(t *testing.T) {
	t.Skip("Skipping TestGetEditorByEmail_Success as GetEditorByEmail is not on EditorRepository interface")
}

func TestGetEditorByEmail_NotFound(t *testing.T) {
	t.Skip("Skipping TestGetEditorByEmail_NotFound as GetEditorByEmail is not on EditorRepository interface")
}

func TestGetEditorByFirebaseUID_Success(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	repo := repository.EditorRepo(suite.DB)

	originalEditor := testutils.CreateTestEditor(1)
	createdEditor, err := repo.InsertEditor(originalEditor.FirebaseUID, originalEditor.Email)
	require.NoError(t, err)

	retrievedEditor, err := repo.GetEditorByFirebaseUID(originalEditor.FirebaseUID)
	require.NoError(t, err)
	require.NotNil(t, retrievedEditor)
	assert.Equal(t, createdEditor.ID, retrievedEditor.ID)
	assert.Equal(t, originalEditor.FirebaseUID, retrievedEditor.FirebaseUID)
}

func TestGetEditorByFirebaseUID_NotFound(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	repo := repository.EditorRepo(suite.DB)

	nonExistentUID := "firebase-uid-" + uuid.New().String()
	retrievedEditor, err := repo.GetEditorByFirebaseUID(nonExistentUID)

	// Repository returns nil, nil for not found cases
	assert.NoError(t, err)
	assert.Nil(t, retrievedEditor)
}

// UpdateEditor and DeleteEditor tests are not explicitly in the RFC for EditorRepository.
// The RFC shows Create, GetByID, GetByEmail, GetByFirebaseUID.
// If UpdateEditor and DeleteEditor methods exist on EditorRepository and need testing,
// they should be added here following a similar pattern.
// For now, sticking to the RFC's listed scenarios for EditorRepository.

// Example for TestUpdateEditor_Success (if UpdateEditor method existed and was to be tested):
/*
func TestUpdateEditor_Success(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	repo := repository.NewEditorRepository(suite.DB)

	// 1. Create an editor
	initialEditor := testutils.CreateTestEditor(1)
	created, err := repo.InsertEditor(ctx, initialEditor.FirebaseUID, initialEditor.Email)
	require.NoError(t, err)

	// 2. Update some fields (assuming UpdateEditor takes the fields to update)
	updatedEditorModel := &models.Editor{
		ID:          created.ID, // Must match the ID of the editor to update
		FirebaseUID: created.FirebaseUID, // Usually FirebaseUID doesn't change
		Email:       "updated-" + uuid.New().String() + "@example.com",
		// Name, etc., if applicable
	}

	// Assume an UpdateEditor method like: err := repo.UpdateEditor(ctx, updatedEditorModel)
	// For this to work, EditorRepository needs an UpdateEditor method.
	// The current repository.Editor struct used in InsertEditor/Get returns *repository.Editor not *models.Editor
	// This needs alignment.

	// err = repo.UpdateEditor(ctx, updatedEditorModel) // Hypothetical call
	// require.NoError(t, err)

	// 3. Retrieve and verify
	// fetched, err := repo.GetEditorByID(ctx, created.ID)
	// require.NoError(t, err)
	// require.NotNil(t, fetched)
	// assert.Equal(t, updatedEditorModel.Email, fetched.Email)
	t.Skip("TestUpdateEditor_Success skipped: UpdateEditor method not specified in RFC for EditorRepository or not yet implemented.")
}
*/

// Example for TestDeleteEditor_Success (if DeleteEditor method existed and was to be tested):
/*
func TestDeleteEditor_Success(t *testing.T) {
	suite := testutils.NewTestSuite(t)
	defer suite.Cleanup(t)
	ctx := context.Background()
	repo := repository.NewEditorRepository(suite.DB)

	// 1. Create an editor
	editorToDelete := testutils.CreateTestEditor(1)
	created, err := repo.InsertEditor(ctx, editorToDelete.FirebaseUID, editorToDelete.Email)
	require.NoError(t, err)

	// 2. Delete it (assuming DeleteEditor takes editor ID)
	// err = repo.DeleteEditor(ctx, created.ID) // Hypothetical call
	// require.NoError(t, err)

	// 3. Try to retrieve it, should be not found
	// _, err = repo.GetEditorByID(ctx, created.ID)
	// assert.Error(t, err)
	// assert.EqualError(t, err, repository.ErrEditorNotFound.Error())
	t.Skip("TestDeleteEditor_Success skipped: DeleteEditor method not specified in RFC for EditorRepository or not yet implemented.")
}
*/ 