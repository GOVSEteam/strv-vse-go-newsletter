package newsletter_test

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/joho/godotenv"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	h "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/handler/newsletter"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	rtr "github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/router" // Alias for router package
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/setup-postgresql"
	_ "github.com/lib/pq" // PostgreSQL driver
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	testServer *httptest.Server
	testDB     *sql.DB
	testEditorDBID string
)

const (
	// Define these in your test environment or a .env.test file
	// For simplicity, using a default. Ensure your local PG is running and this DB exists.
	// Or use a Dockerized PG for tests.
	// testDatabaseURL = "postgres://user:password@localhost:5432/newsletter_test_db?sslmode=disable" // REMOVED

	testEditorFirebaseUID = "test-editor-firebase-uid-integration"
	testEditorEmail       = "editor-integration@example.com"
	// testEditorPassword    = "password123" // Not used if we mock VerifyFirebaseJWT
)

// --- Test Schemas (corrected and with updated_at) ---
const schemaEditors = `
CREATE EXTENSION IF NOT EXISTS "pgcrypto";
DROP TABLE IF EXISTS editors CASCADE;
CREATE TABLE IF NOT EXISTS editors (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    firebase_uid TEXT UNIQUE NOT NULL,
    email TEXT UNIQUE NOT NULL
);
`

const schemaNewsletters = `
DROP TABLE IF EXISTS newsletters CASCADE;
CREATE TABLE IF NOT EXISTS newsletters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    editor_id UUID NOT NULL REFERENCES editors(id) ON DELETE CASCADE,
    name TEXT NOT NULL,
    description TEXT,
    created_at TIMESTAMPTZ DEFAULT now(),
    updated_at TIMESTAMPTZ DEFAULT now()
);
ALTER TABLE newsletters ADD CONSTRAINT newsletter_name_per_editor_unique UNIQUE (editor_id, name);
`

func TestMain(m *testing.M) {
	// Get the directory of the current test file
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		panic("Failed to get current test file path")
	}
	basepath := filepath.Dir(currentFile)
	envPath := filepath.Join(basepath, "../../../../.env") // Path to .env in project root (corrected to 4 levels up)

	println("DEBUG: Attempting to load .env from calculated path:", envPath)

	// Attempt to load .env file from the calculated project root path.
	err := godotenv.Load(envPath)
	if err != nil {
		println("Note: .env file not loaded or error during load (this might be fine in CI):", err.Error())
	} else {
		println("Successfully loaded .env file from:", envPath)
	}

	// Log the URLs being used by setup_postgresql.ConnectDB internally
	println("DEBUG: DATABASE_URL from env in TestMain: ", os.Getenv("DATABASE_URL"))
	println("DEBUG: DATABASE_PUBLIC_URL from env in TestMain: ", os.Getenv("DATABASE_PUBLIC_URL"))
	println("DEBUG: RAILWAY_ENVIRONMENT from env in TestMain: ", os.Getenv("RAILWAY_ENVIRONMENT"))

	// Setup: Database connection
	// originalDbURL := os.Getenv("DATABASE_URL") // REMOVED
	// os.Setenv("DATABASE_URL", testDatabaseURL) // REMOVED
	// testDB = setup_postgresql.ConnectDB() // Uses our test DSN // MODIFIED BELOW
	// os.Setenv("DATABASE_URL", originalDbURL) // Restore original DSN // REMOVED

	// testDB will be initialized using the DATABASE_URL from the environment,
	// which should be loaded from the .env file by setup_postgresql.ConnectDB() or by the execution environment.
	testDB = setup_postgresql.ConnectDB()
	if testDB == nil {
		// The setup_postgresql.ConnectDB() should ideally handle errors,
		// but we add a check here for robustness in the test context.
		// If using Go 1.21+, you might use log.Fatalf directly.
		// For now, ensuring it's not nil before proceeding.
		// A proper error check/panic from ConnectDB is better.
		panic("Failed to connect to testDB. Ensure DATABASE_URL is set and the DB is accessible.")
	}

	defer testDB.Close()

	// Setup: Initialize router and test server
	router := rtr.Router() // This uses the real DB connection logic
	testServer = httptest.NewServer(router)
	defer testServer.Close()

	// Run tests
	code := m.Run()

	// Teardown (if any)
	os.Exit(code)
}

// setupTestEditor ensures a test editor exists and sets testEditorDBID
func setupTestEditor(t *testing.T) {
	_, err := testDB.Exec(schemaEditors) // Ensure table exists
	require.NoError(t, err, "Failed to apply editors schema")

	// Clear existing test editor if any (based on firebase_uid)
	_, err = testDB.Exec(`DELETE FROM editors WHERE firebase_uid = $1`, testEditorFirebaseUID)
	require.NoError(t, err)

	// Insert test editor
	var editorID string
	err = testDB.QueryRow(`INSERT INTO editors (firebase_uid, email) VALUES ($1, $2) RETURNING id`, 
		testEditorFirebaseUID, testEditorEmail).Scan(&editorID)
	require.NoError(t, err, "Failed to insert test editor")
	testEditorDBID = editorID

	// Mock JWT verification to use this editor
	auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
		// Check for a specific test token if we were generating one
		// For now, just return the test Firebase UID directly for authenticated routes
		return testEditorFirebaseUID, nil
	}
}

func cleanupTables(t *testing.T) {
	_, err := testDB.Exec(`DELETE FROM newsletters; DELETE FROM editors;`)
	require.NoError(t, err, "Failed to clean tables")
	// Re-apply schemas to ensure they are pristine (or use TRUNCATE ... RESTART IDENTITY CASCADE)
	_, err = testDB.Exec(schemaEditors)
	require.NoError(t, err)
	_, err = testDB.Exec(schemaNewsletters)
	require.NoError(t, err)
}

// --- Actual Tests ---

func TestNewsletterAPI_CreateAndList(t *testing.T) {
	cleanupTables(t) // Ensure clean state
	setupTestEditor(t) // Ensure our test editor exists and JWT mock is active

	// 1. Create Newsletter (Success)
	createPayload := map[string]string{
		"name":        "Integration Test Newsletter",
		"description": "Created via API test",
	}
	bodyBytes, _ := json.Marshal(createPayload)
	req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/api/newsletters", bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token") // Token content doesn't matter due to VerifyFirebaseJWT mock

	client := &http.Client{}
	resp, err := client.Do(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode, "Create should succeed")

	var createdNewsletter repository.Newsletter
	json.NewDecoder(resp.Body).Decode(&createdNewsletter)
	resp.Body.Close()

	assert.NotEmpty(t, createdNewsletter.ID)
	assert.Equal(t, createPayload["name"], createdNewsletter.Name)
	assert.Equal(t, testEditorDBID, createdNewsletter.EditorID)

	// 2. Create Newsletter (Name Conflict)
	reqConflict, _ := http.NewRequest(http.MethodPost, testServer.URL+"/api/newsletters", bytes.NewReader(bodyBytes)) // Same payload
	reqConflict.Header.Set("Content-Type", "application/json")
	reqConflict.Header.Set("Authorization", "Bearer test-token")
	respConflict, err := client.Do(reqConflict)
	require.NoError(t, err)
	assert.Equal(t, http.StatusConflict, respConflict.StatusCode, "Create with conflicting name should fail")
	respConflict.Body.Close()

	// 3. List Newsletters (Success, should see the one created)
	reqList, _ := http.NewRequest(http.MethodGet, testServer.URL+"/api/newsletters?limit=5&offset=0", nil)
	reqList.Header.Set("Authorization", "Bearer test-token")
	respList, err := client.Do(reqList)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, respList.StatusCode, "List should succeed")

	var listResponse h.PaginatedNewslettersResponse
	json.NewDecoder(respList.Body).Decode(&listResponse)
	respList.Body.Close()

	assert.Equal(t, 1, listResponse.Total, "Should have 1 newsletter for this editor")
	require.Len(t, listResponse.Data, 1, "List data should contain 1 newsletter")
	assert.Equal(t, createdNewsletter.ID, listResponse.Data[0].ID)
	assert.Equal(t, createPayload["name"], listResponse.Data[0].Name)
}

func TestNewsletterAPI_UpdateAndDelete(t *testing.T) {
	cleanupTables(t)
	setupTestEditor(t)
	client := &http.Client{}

	// --- Setup: Create a newsletter first ---
	initialName := "Newsletter To Update & Delete"
	initialDesc := "Initial description"
	createP := map[string]string{"name": initialName, "description": initialDesc}
	bodyBytes, _ := json.Marshal(createP)
	reqCreate, _ := http.NewRequest(http.MethodPost, testServer.URL+"/api/newsletters", bytes.NewReader(bodyBytes))
	reqCreate.Header.Set("Content-Type", "application/json")
	reqCreate.Header.Set("Authorization", "Bearer test-token")
	respCreate, err := client.Do(reqCreate)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, respCreate.StatusCode)
	var nl repository.Newsletter
	json.NewDecoder(respCreate.Body).Decode(&nl)
	respCreate.Body.Close()
	require.NotEmpty(t, nl.ID)

	// --- Test Update ---
	t.Run("UpdateNewsletter_Success", func(t *testing.T) {
		updateName := "Updated Newsletter Name via API"
		updatePayload := map[string]string{"name": updateName} // Only updating name
		updateBody, _ := json.Marshal(updatePayload)
		reqUpdate, _ := http.NewRequest(http.MethodPatch, testServer.URL+"/api/newsletters/"+nl.ID, bytes.NewReader(updateBody))
		reqUpdate.Header.Set("Content-Type", "application/json")
		reqUpdate.Header.Set("Authorization", "Bearer test-token")
		respUpdate, err := client.Do(reqUpdate)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, respUpdate.StatusCode)
		var updatedNL repository.Newsletter
		json.NewDecoder(respUpdate.Body).Decode(&updatedNL)
		respUpdate.Body.Close()
		assert.Equal(t, updateName, updatedNL.Name)
		assert.Equal(t, initialDesc, updatedNL.Description) // Description should be unchanged
		assert.NotEqual(t, nl.UpdatedAt, updatedNL.UpdatedAt) // UpdatedAt should have changed
	})

	t.Run("UpdateNewsletter_NameConflict", func(t *testing.T) {
		// Create another newsletter to cause a name conflict
		conflictNLName := "Existing Other Newsletter"
		cnp := map[string]string{"name": conflictNLName, "description": "desc"}
		cb, _ := json.Marshal(cnp)
		rc, _ := http.NewRequest(http.MethodPost, testServer.URL+"/api/newsletters", bytes.NewReader(cb))
		rc.Header.Set("Content-Type", "application/json")
		rc.Header.Set("Authorization", "Bearer test-token")
		respc, err := client.Do(rc)
		require.NoError(t, err); require.Equal(t, http.StatusCreated, respc.StatusCode); respc.Body.Close()

		// Try to update the original newsletter (nl.ID) to this conflicting name
		updatePayload := map[string]string{"name": conflictNLName}
		updateBody, _ := json.Marshal(updatePayload)
		reqUpdate, _ := http.NewRequest(http.MethodPatch, testServer.URL+"/api/newsletters/"+nl.ID, bytes.NewReader(updateBody))
		reqUpdate.Header.Set("Content-Type", "application/json")
		reqUpdate.Header.Set("Authorization", "Bearer test-token")
		respUpdate, err := client.Do(reqUpdate)
		require.NoError(t, err)
		assert.Equal(t, http.StatusConflict, respUpdate.StatusCode)
		respUpdate.Body.Close()
	})

	t.Run("UpdateNewsletter_NotFound", func(t *testing.T) {
		updateName := "Doesn't Matter"
		updatePayload := map[string]string{"name": updateName}
		updateBody, _ := json.Marshal(updatePayload)
		// Use a valid UUID format for the non-existent ID
		nonExistentUUID := "00000000-0000-0000-0000-000000000000"
		reqUpdate, _ := http.NewRequest(http.MethodPatch, testServer.URL+"/api/newsletters/"+nonExistentUUID, bytes.NewReader(updateBody))
		reqUpdate.Header.Set("Content-Type", "application/json")
		reqUpdate.Header.Set("Authorization", "Bearer test-token")
		respUpdate, err := client.Do(reqUpdate)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, respUpdate.StatusCode)
		respUpdate.Body.Close()
	})

	// --- Test Delete ---
	t.Run("DeleteNewsletter_Success", func(t *testing.T) {
		reqDelete, _ := http.NewRequest(http.MethodDelete, testServer.URL+"/api/newsletters/"+nl.ID, nil)
		reqDelete.Header.Set("Authorization", "Bearer test-token")
		respDelete, err := client.Do(reqDelete)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNoContent, respDelete.StatusCode)
		respDelete.Body.Close()

		// Verify it's actually deleted (e.g., try to get it - should be 404, or list should not contain it)
		reqGet, _ := http.NewRequest(http.MethodGet, testServer.URL+"/api/newsletters?limit=10&offset=0", nil) // Assuming ListHandler for a specific ID is not implemented, check full list
		reqGet.Header.Set("Authorization", "Bearer test-token")
		respGet, err := client.Do(reqGet)
		require.NoError(t, err)
		require.Equal(t, http.StatusOK, respGet.StatusCode) // List itself should be OK
		var listResponse h.PaginatedNewslettersResponse
		json.NewDecoder(respGet.Body).Decode(&listResponse)
		respGet.Body.Close()

		found := false
		for _, item := range listResponse.Data {
			if item.ID == nl.ID {
				found = true
				break
			}
		}
		assert.False(t, found, "Deleted newsletter should not be in the list")
	})

	t.Run("DeleteNewsletter_NotFound", func(t *testing.T) {
		// nl.ID should already be deleted from the previous sub-test
		reqDelete, _ := http.NewRequest(http.MethodDelete, testServer.URL+"/api/newsletters/"+nl.ID, nil)
		reqDelete.Header.Set("Authorization", "Bearer test-token")
		respDelete, err := client.Do(reqDelete)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, respDelete.StatusCode)
		respDelete.Body.Close()

		// Also test with a completely random non-existent ID (but valid UUID format)
		completelyNonExistentUUID := "11111111-1111-1111-1111-111111111111"
		reqDeleteNonExistent, _ := http.NewRequest(http.MethodDelete, testServer.URL+"/api/newsletters/"+completelyNonExistentUUID, nil)
		reqDeleteNonExistent.Header.Set("Authorization", "Bearer test-token")
		respDeleteNonExistent, err := client.Do(reqDeleteNonExistent)
		require.NoError(t, err)
		assert.Equal(t, http.StatusNotFound, respDeleteNonExistent.StatusCode)
		respDeleteNonExistent.Body.Close()
	})
} 