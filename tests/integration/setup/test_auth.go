package setup

import (
	"database/sql"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/auth"
	"github.com/stretchr/testify/require"
)

// TestAuthConfig holds authentication configuration for tests
type TestAuthConfig struct {
	FirebaseUID string
	Email       string
	EditorID    string
}

// DefaultTestAuthConfig returns default test authentication configuration
func DefaultTestAuthConfig() TestAuthConfig {
	// Generate unique Firebase UID using timestamp to avoid conflicts
	timestamp := time.Now().UnixNano()
	return TestAuthConfig{
		FirebaseUID: fmt.Sprintf("test-firebase-uid-%d", timestamp),
		Email:       fmt.Sprintf("test-editor-%d@integration.test", timestamp),
		EditorID:    "", // Will be set after editor creation
	}
}

// MockFirebaseAuth sets up Firebase authentication mocking for tests
func MockFirebaseAuth(t *testing.T, config TestAuthConfig) func() {
	// Store original function
	originalVerifyJWT := auth.VerifyFirebaseJWT

	// Mock the JWT verification function
	auth.VerifyFirebaseJWT = func(r *http.Request) (string, error) {
		// Check for test authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			return "", fmt.Errorf("missing Authorization header")
		}

		// For integration tests, we accept a specific test token
		expectedToken := "Bearer test-jwt-token-" + config.FirebaseUID
		if authHeader != expectedToken {
			return "", fmt.Errorf("invalid test token")
		}

		return config.FirebaseUID, nil
	}

	// Return cleanup function
	return func() {
		auth.VerifyFirebaseJWT = originalVerifyJWT
	}
}

// CreateAuthenticatedRequest creates an HTTP request with authentication headers
func CreateAuthenticatedRequest(method, url string, body interface{}, config TestAuthConfig) (*http.Request, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}

	// Add test authentication header
	testToken := "test-jwt-token-" + config.FirebaseUID
	req.Header.Set("Authorization", "Bearer "+testToken)
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

// AuthenticatedTestClient wraps http.Client with automatic authentication
type AuthenticatedTestClient struct {
	*http.Client
	Config TestAuthConfig
}

// NewAuthenticatedTestClient creates a new authenticated test client
func NewAuthenticatedTestClient(config TestAuthConfig) *AuthenticatedTestClient {
	return &AuthenticatedTestClient{
		Client: &http.Client{},
		Config: config,
	}
}

// Do performs an HTTP request with automatic authentication
func (c *AuthenticatedTestClient) Do(req *http.Request) (*http.Response, error) {
	// Add authentication header if not already present
	if req.Header.Get("Authorization") == "" {
		testToken := "test-jwt-token-" + c.Config.FirebaseUID
		req.Header.Set("Authorization", "Bearer "+testToken)
	}

	// Ensure content type is set for JSON requests
	if req.Header.Get("Content-Type") == "" && req.Body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	return c.Client.Do(req)
}

// TestAuthHelper provides utilities for authentication in tests
type TestAuthHelper struct {
	Config      TestAuthConfig
	CleanupFunc func()
}

// NewTestAuthHelper creates a new test authentication helper
func NewTestAuthHelper(t *testing.T, config TestAuthConfig) *TestAuthHelper {
	cleanupFunc := MockFirebaseAuth(t, config)
	
	return &TestAuthHelper{
		Config:      config,
		CleanupFunc: cleanupFunc,
	}
}

// Cleanup restores original authentication functions
func (h *TestAuthHelper) Cleanup() {
	if h.CleanupFunc != nil {
		h.CleanupFunc()
	}
}

// CreateTestEditor creates a test editor in the database and returns the editor ID
func (h *TestAuthHelper) CreateTestEditor(t *testing.T, db *sql.DB) string {
	// Insert test editor
	var editorID string
	query := `INSERT INTO editors (firebase_uid, email) VALUES ($1, $2) RETURNING id`
	err := db.QueryRow(query, h.Config.FirebaseUID, h.Config.Email).Scan(&editorID)
	require.NoError(t, err, "Failed to create test editor")

	// Update config with editor ID
	h.Config.EditorID = editorID

	return editorID
}

// GetAuthHeaders returns authentication headers for HTTP requests
func (h *TestAuthHelper) GetAuthHeaders() map[string]string {
	return map[string]string{
		"Authorization": "Bearer test-jwt-token-" + h.Config.FirebaseUID,
		"Content-Type":  "application/json",
	}
}

// AddAuthHeaders adds authentication headers to an HTTP request
func (h *TestAuthHelper) AddAuthHeaders(req *http.Request) {
	headers := h.GetAuthHeaders()
	for key, value := range headers {
		req.Header.Set(key, value)
	}
} 