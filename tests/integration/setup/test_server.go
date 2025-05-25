package setup

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/router"
	"github.com/joho/godotenv"
)

// TestServer wraps httptest.Server with additional test utilities
type TestServer struct {
	*httptest.Server
	DB     *sql.DB
	Client *http.Client
}

// NewTestServer creates a new test server with real dependencies
func NewTestServer(t *testing.T) *TestServer {
	// Load test environment
	if err := godotenv.Load("../../../.env"); err != nil {
		t.Logf("Warning: Could not load .env file: %v", err)
	}

	// Ensure we're using test database
	ensureTestEnvironment(t)

	// Create test database
	testDB := SetupTestDatabase(t)

	// Create router with real dependencies
	appRouter := router.Router()

	// Create test server
	server := httptest.NewServer(appRouter)

	// Create HTTP client for making requests
	client := &http.Client{}

	return &TestServer{
		Server: server,
		DB:     testDB,
		Client: client,
	}
}

// Close cleans up the test server and database
func (ts *TestServer) Close() {
	if ts.Server != nil {
		ts.Server.Close()
	}
	if ts.DB != nil {
		CleanupTestDatabase(ts.DB)
		ts.DB.Close()
	}
}

// URL returns the base URL for the test server
func (ts *TestServer) URL() string {
	return ts.Server.URL
}

// ensureTestEnvironment verifies we're running in a test environment
func ensureTestEnvironment(t *testing.T) {
	// Check if we're using a test database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		if publicURL := os.Getenv("DATABASE_PUBLIC_URL"); publicURL != "" {
			dbURL = publicURL
		}
	}

	if dbURL == "" {
		t.Skip("No database URL configured for integration tests")
	}

	// Warn if not using a test database
	if !containsTestIndicator(dbURL) {
		t.Logf("WARNING: Database URL doesn't appear to be a test database: %s", dbURL)
		t.Logf("Consider using a dedicated test database to avoid data loss")
	}
}

// containsTestIndicator checks if the database URL indicates it's for testing
func containsTestIndicator(dbURL string) bool {
	testIndicators := []string{"test", "testing", "integration", "ci"}
	for _, indicator := range testIndicators {
		if contains(dbURL, indicator) {
			return true
		}
	}
	return false
}

// contains checks if a string contains a substring (case-insensitive)
func contains(s, substr string) bool {
	return len(s) >= len(substr) && 
		   (s == substr || 
		    (len(s) > len(substr) && 
		     (s[:len(substr)] == substr || 
		      s[len(s)-len(substr):] == substr ||
		      containsSubstring(s, substr))))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
} 