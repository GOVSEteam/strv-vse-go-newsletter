package setup

import (
	"database/sql"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

// TestFixtures holds all test data fixtures
type TestFixtures struct {
	DB       *sql.DB
	EditorID string
	AuthConfig TestAuthConfig
}

// NewTestFixtures creates a new test fixtures instance
func NewTestFixtures(t *testing.T, db *sql.DB, authConfig TestAuthConfig) *TestFixtures {
	fixtures := &TestFixtures{
		DB:         db,
		AuthConfig: authConfig,
	}

	// Create test editor
	fixtures.EditorID = fixtures.CreateTestEditor(t)

	return fixtures
}

// NewTestFixturesWithEditor creates a new test fixtures instance with an existing editor ID
func NewTestFixturesWithEditor(t *testing.T, db *sql.DB, authConfig TestAuthConfig, editorID string) *TestFixtures {
	fixtures := &TestFixtures{
		DB:         db,
		AuthConfig: authConfig,
		EditorID:   editorID,
	}

	return fixtures
}

// CreateTestEditor creates a test editor and returns the ID
func (f *TestFixtures) CreateTestEditor(t *testing.T) string {
	query := `INSERT INTO editors (firebase_uid, email) VALUES ($1, $2) RETURNING id`
	var editorID string
	err := f.DB.QueryRow(query, f.AuthConfig.FirebaseUID, f.AuthConfig.Email).Scan(&editorID)
	require.NoError(t, err, "Failed to create test editor")
	return editorID
}

// CreateTestNewsletter creates a test newsletter and returns the ID
func (f *TestFixtures) CreateTestNewsletter(t *testing.T, name, description string) string {
	if name == "" {
		name = "Test Newsletter"
	}
	if description == "" {
		description = "A test newsletter for integration testing"
	}

	query := `INSERT INTO newsletters (name, description, editor_id, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`
	
	now := time.Now()
	var newsletterID string
	err := f.DB.QueryRow(query, name, description, f.EditorID, now, now).Scan(&newsletterID)
	require.NoError(t, err, "Failed to create test newsletter")
	
	return newsletterID
}

// CreateTestPost creates a test post and returns the ID
func (f *TestFixtures) CreateTestPost(t *testing.T, newsletterID, title, content string) string {
	if title == "" {
		title = "Test Post"
	}
	if content == "" {
		content = "This is test content for integration testing"
	}

	query := `INSERT INTO posts (newsletter_id, title, content, created_at, updated_at) 
			  VALUES ($1, $2, $3, $4, $5) RETURNING id`
	
	now := time.Now()
	var postID string
	err := f.DB.QueryRow(query, newsletterID, title, content, now, now).Scan(&postID)
	require.NoError(t, err, "Failed to create test post")
	
	return postID
}

// CreateCompleteTestData creates a complete set of test data (editor, newsletter, post)
// Note: Subscribers are now stored in Firestore and created via API calls
func (f *TestFixtures) CreateCompleteTestData(t *testing.T) TestDataSet {
	newsletterID := f.CreateTestNewsletter(t, "Integration Test Newsletter", "Newsletter for testing")
	postID := f.CreateTestPost(t, newsletterID, "Integration Test Post", "Content for testing")

	return TestDataSet{
		EditorID:     f.EditorID,
		NewsletterID: newsletterID,
		PostID:       postID,
	}
}

// TestDataSet holds a complete set of related test data
// Note: SubscriberID removed since subscribers are now in Firestore
type TestDataSet struct {
	EditorID     string
	NewsletterID string
	PostID       string
}

// Cleanup removes all test data created by this fixtures instance
func (f *TestFixtures) Cleanup(t *testing.T) {
	// Clean in reverse dependency order
	// Note: Subscribers are in Firestore and will be cleaned up separately if needed
	cleanupQueries := []string{
		"DELETE FROM posts WHERE newsletter_id IN (SELECT id FROM newsletters WHERE editor_id = $1)",
		"DELETE FROM newsletters WHERE editor_id = $1",
		"DELETE FROM editors WHERE id = $1",
	}

	for _, query := range cleanupQueries {
		_, err := f.DB.Exec(query, f.EditorID)
		if err != nil {
			t.Logf("Warning: Failed to cleanup with query %s: %v", query, err)
		}
	}
}

// NewsletterTestData holds test data for newsletter testing
type NewsletterTestData struct {
	Name        string
	Description string
}

// DefaultNewsletterTestData returns default newsletter test data
func DefaultNewsletterTestData() NewsletterTestData {
	return NewsletterTestData{
		Name:        "Integration Test Newsletter",
		Description: "A newsletter created for integration testing purposes",
	}
}

// PostTestData holds test data for post testing
type PostTestData struct {
	Title   string
	Content string
}

// DefaultPostTestData returns default post test data
// Note: Status field removed since posts use published_at instead
func DefaultPostTestData() PostTestData {
	return PostTestData{
		Title:   "Integration Test Post",
		Content: "This is the content of a test post created for integration testing. It contains enough text to be meaningful for testing purposes.",
	}
}

// SubscriberTestData holds test data for subscriber testing
// Note: Subscribers are now stored in Firestore with different status values
type SubscriberTestData struct {
	Email  string
	Status string
}

// DefaultSubscriberTestData returns default subscriber test data
// Note: Status values updated to match Firestore subscriber model
func DefaultSubscriberTestData() SubscriberTestData {
	return SubscriberTestData{
		Email:  "subscriber@integration.test",
		Status: "pending_confirmation", // Updated to match Firestore model
	}
} 