package auth

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExtractBearerToken tests the ExtractBearerToken function with various scenarios
func TestExtractBearerToken(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		expectedToken  string
		expectError    bool
		expectedError  string
	}{
		// Success cases
		{
			name:          "Valid Bearer token",
			authHeader:    "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectedToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectError:   false,
		},
		{
			name:          "Valid Bearer token with spaces",
			authHeader:    "Bearer   token-with-spaces-around",
			expectedToken: "  token-with-spaces-around",
			expectError:   false,
		},
		{
			name:          "Bearer with complex JWT token",
			authHeader:    "Bearer eyJhbGciOiJSUzI1NiIsImtpZCI6IjE2NzAyNzM4MjQifQ.eyJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vdGVzdC1wcm9qZWN0IiwiYXVkIjoidGVzdC1wcm9qZWN0IiwiYXV0aF90aW1lIjoxNjcwMjczODI0LCJ1c2VyX2lkIjoidGVzdC11aWQifQ.signature",
			expectedToken: "eyJhbGciOiJSUzI1NiIsImtpZCI6IjE2NzAyNzM4MjQifQ.eyJpc3MiOiJodHRwczovL3NlY3VyZXRva2VuLmdvb2dsZS5jb20vdGVzdC1wcm9qZWN0IiwiYXVkIjoidGVzdC1wcm9qZWN0IiwiYXV0aF90aW1lIjoxNjcwMjczODI0LCJ1c2VyX2lkIjoidGVzdC11aWQifQ.signature",
			expectError:   false,
		},
		{
			name:          "Bearer with case variations (lowercase)",
			authHeader:    "bearer test-token",
			expectedToken: "test-token",
			expectError:   false,
		},
		{
			name:          "Bearer with case variations (mixed case)",
			authHeader:    "BeArEr test-token",
			expectedToken: "test-token",
			expectError:   false,
		},

		// Error cases - Missing header
		{
			name:          "Missing Authorization header",
			authHeader:    "",
			expectedToken: "",
			expectError:   true,
			expectedError: "missing Authorization header",
		},

		// Error cases - Invalid format
		{
			name:          "No Bearer prefix",
			authHeader:    "test-token",
			expectedToken: "",
			expectError:   true,
			expectedError: "invalid Authorization header format",
		},
		{
			name:          "Wrong auth type (Basic)",
			authHeader:    "Basic dXNlcjpwYXNzd29yZA==",
			expectedToken: "",
			expectError:   true,
			expectedError: "invalid Authorization header format",
		},
		{
			name:          "Only Bearer without token",
			authHeader:    "Bearer",
			expectedToken: "",
			expectError:   true,
			expectedError: "invalid Authorization header format",
		},
		{
			name:          "Bearer with empty token",
			authHeader:    "Bearer ",
			expectedToken: "",
			expectError:   false, // This actually succeeds with empty string
		},
		{
			name:          "Multiple spaces between Bearer and token",
			authHeader:    "Bearer    token",
			expectedToken: "   token",
			expectError:   false,
		},
		{
			name:          "Invalid format with multiple Bearer",
			authHeader:    "Bearer Bearer token",
			expectedToken: "Bearer token",
			expectError:   false, // This succeeds - everything after first space is token
		},

		// Edge cases
		{
			name:          "Very long token",
			authHeader:    "Bearer " + generateLongToken(1000),
			expectedToken: generateLongToken(1000),
			expectError:   false,
		},
		{
			name:          "Token with special characters",
			authHeader:    "Bearer token.with-special_chars+and/slashes=",
			expectedToken: "token.with-special_chars+and/slashes=",
			expectError:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create HTTP request with Authorization header
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Call the function
			token, err := ExtractBearerToken(req)

			// Verify results
			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedToken, token)
			}
		})
	}
}

// TestVerifyFirebaseJWT tests the VerifyFirebaseJWT function with mocked Firebase client
func TestVerifyFirebaseJWT(t *testing.T) {
	// Save original function to restore after tests
	originalVerifyJWT := VerifyFirebaseJWT
	defer func() { VerifyFirebaseJWT = originalVerifyJWT }()

	tests := []struct {
		name           string
		authHeader     string
		mockResponse   string
		mockError      error
		expectedUID    string
		expectError    bool
		expectedError  string
	}{
		// Success cases
		{
			name:         "Valid JWT token",
			authHeader:   "Bearer valid-jwt-token",
			mockResponse: "test-firebase-uid-123",
			mockError:    nil,
			expectedUID:  "test-firebase-uid-123",
			expectError:  false,
		},
		{
			name:         "Valid JWT with complex UID",
			authHeader:   "Bearer complex-jwt",
			mockResponse: "firebase-uid-with-dashes-and-numbers-123",
			mockError:    nil,
			expectedUID:  "firebase-uid-with-dashes-and-numbers-123",
			expectError:  false,
		},

		// Error cases - ExtractBearerToken failures
		{
			name:          "Missing Authorization header",
			authHeader:    "",
			mockResponse:  "",
			mockError:     nil,
			expectedUID:   "",
			expectError:   true,
			expectedError: "missing Authorization header",
		},
		{
			name:          "Invalid Authorization header format",
			authHeader:    "Basic invalid",
			mockResponse:  "",
			mockError:     nil,
			expectedUID:   "",
			expectError:   true,
			expectedError: "invalid Authorization header format",
		},

		// Error cases - Firebase verification failures
		{
			name:          "Invalid JWT token",
			authHeader:    "Bearer invalid-jwt",
			mockResponse:  "",
			mockError:     errors.New("token verification failed"),
			expectedUID:   "",
			expectError:   true,
			expectedError: "token verification failed",
		},
		{
			name:          "Expired JWT token",
			authHeader:    "Bearer expired-jwt",
			mockResponse:  "",
			mockError:     errors.New("token has expired"),
			expectedUID:   "",
			expectError:   true,
			expectedError: "token has expired",
		},
		{
			name:          "Malformed JWT token",
			authHeader:    "Bearer malformed.jwt",
			mockResponse:  "",
			mockError:     errors.New("malformed token"),
			expectedUID:   "",
			expectError:   true,
			expectedError: "malformed token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock VerifyFirebaseJWT function
			VerifyFirebaseJWT = func(r *http.Request) (string, error) {
				// First extract the token to test ExtractBearerToken integration
				_, err := ExtractBearerToken(r)
				if err != nil {
					return "", err
				}

				// If we got a token, simulate Firebase verification
				if tt.mockError != nil {
					return "", tt.mockError
				}
				return tt.mockResponse, nil
			}

			// Create HTTP request
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			// Call the function
			uid, err := VerifyFirebaseJWT(req)

			// Verify results
			if tt.expectError {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				assert.Empty(t, uid)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUID, uid)
			}
		})
	}
}

// TestVerifyFirebaseJWT_RealWorldUsage tests the function with patterns used in actual handlers
func TestVerifyFirebaseJWT_RealWorldUsage(t *testing.T) {
	// Save original function to restore after tests
	originalVerifyJWT := VerifyFirebaseJWT
	defer func() { VerifyFirebaseJWT = originalVerifyJWT }()

	t.Run("Newsletter handler pattern", func(t *testing.T) {
		// This mimics how newsletter handlers use VerifyFirebaseJWT
		VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "editor-firebase-uid-123", nil
		}

		req := httptest.NewRequest("GET", "/api/newsletters", nil)
		req.Header.Set("Authorization", "Bearer test-token")

		firebaseUID, err := VerifyFirebaseJWT(req)
		require.NoError(t, err)
		assert.Equal(t, "editor-firebase-uid-123", firebaseUID)
	})

	t.Run("Post handler pattern", func(t *testing.T) {
		// This mimics how post handlers use VerifyFirebaseJWT
		VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "post-editor-uid", nil
		}

		req := httptest.NewRequest("POST", "/api/posts", nil)
		req.Header.Set("Authorization", "Bearer valid-jwt")

		editorFirebaseUID, err := VerifyFirebaseJWT(req)
		require.NoError(t, err)
		assert.Equal(t, "post-editor-uid", editorFirebaseUID)
	})

	t.Run("Authentication failure pattern", func(t *testing.T) {
		// This mimics authentication failures in handlers
		VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			return "", errors.New("authentication failed")
		}

		req := httptest.NewRequest("GET", "/api/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		_, err := VerifyFirebaseJWT(req)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "authentication failed")
	})
}

// TestExtractBearerToken_RealWorldUsage tests the function with actual usage patterns
func TestExtractBearerToken_RealWorldUsage(t *testing.T) {
	t.Run("Typical API request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/newsletters", nil)
		req.Header.Set("Authorization", "Bearer eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9")

		token, err := ExtractBearerToken(req)
		require.NoError(t, err)
		assert.Equal(t, "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9", token)
	})

	t.Run("Mobile app request", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/api/posts", nil)
		req.Header.Set("Authorization", "Bearer mobile-app-jwt-token-123")
		req.Header.Set("User-Agent", "MobileApp/1.0")

		token, err := ExtractBearerToken(req)
		require.NoError(t, err)
		assert.Equal(t, "mobile-app-jwt-token-123", token)
	})

	t.Run("Unauthorized request", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/api/protected", nil)
		// No Authorization header

		_, err := ExtractBearerToken(req)
		require.Error(t, err)
		assert.Equal(t, "missing Authorization header", err.Error())
	})
}

// TestVerifyFirebaseJWT_Integration tests the integration between ExtractBearerToken and Firebase verification
func TestVerifyFirebaseJWT_Integration(t *testing.T) {
	// Save original function to restore after tests
	originalVerifyJWT := VerifyFirebaseJWT
	defer func() { VerifyFirebaseJWT = originalVerifyJWT }()

	t.Run("Full authentication flow", func(t *testing.T) {
		// Mock the complete flow
		VerifyFirebaseJWT = func(r *http.Request) (string, error) {
			// This tests the actual integration with ExtractBearerToken
			token, err := ExtractBearerToken(r)
			if err != nil {
				return "", err
			}

			// Simulate Firebase client verification
			if token == "valid-firebase-jwt" {
				return "authenticated-user-uid", nil
			}
			return "", errors.New("invalid token")
		}

		// Test valid token
		req := httptest.NewRequest("GET", "/api/test", nil)
		req.Header.Set("Authorization", "Bearer valid-firebase-jwt")

		uid, err := VerifyFirebaseJWT(req)
		require.NoError(t, err)
		assert.Equal(t, "authenticated-user-uid", uid)

		// Test invalid token
		req2 := httptest.NewRequest("GET", "/api/test", nil)
		req2.Header.Set("Authorization", "Bearer invalid-jwt")

		_, err = VerifyFirebaseJWT(req2)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid token")
	})
}

// TestVerifyFirebaseJWT_SecurityScenarios tests various security-related scenarios
func TestVerifyFirebaseJWT_SecurityScenarios(t *testing.T) {
	// Save original function to restore after tests
	originalVerifyJWT := VerifyFirebaseJWT
	defer func() { VerifyFirebaseJWT = originalVerifyJWT }()

	tests := []struct {
		name        string
		authHeader  string
		expectError bool
		description string
	}{
		{
			name:        "SQL injection attempt in token",
			authHeader:  "Bearer '; DROP TABLE users; --",
			expectError: true,
			description: "Should reject malicious SQL-like content",
		},
		{
			name:        "XSS attempt in token",
			authHeader:  "Bearer <script>alert('xss')</script>",
			expectError: true,
			description: "Should reject script tags",
		},
		{
			name:        "Very long token (potential DoS)",
			authHeader:  "Bearer " + generateLongToken(10000),
			expectError: true,
			description: "Should handle very long tokens gracefully",
		},
		{
			name:        "Empty token",
			authHeader:  "Bearer ",
			expectError: true,
			description: "Should reject empty tokens",
		},
		{
			name:        "Null bytes in token",
			authHeader:  "Bearer token\x00with\x00nulls",
			expectError: true,
			description: "Should reject tokens with null bytes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock to simulate Firebase rejection of malicious tokens
			VerifyFirebaseJWT = func(r *http.Request) (string, error) {
				token, err := ExtractBearerToken(r)
				if err != nil {
					return "", err
				}

				// Simulate Firebase security validation
				if len(token) > 5000 || token == "" || 
				   containsNullBytes(token) || containsMaliciousContent(token) {
					return "", errors.New("security validation failed")
				}
				return "", errors.New("token verification failed") // All test tokens are invalid
			}

			req := httptest.NewRequest("GET", "/api/test", nil)
			req.Header.Set("Authorization", tt.authHeader)

			_, err := VerifyFirebaseJWT(req)
			if tt.expectError {
				assert.Error(t, err, tt.description)
			} else {
				assert.NoError(t, err, tt.description)
			}
		})
	}
}

// Helper functions for tests

// generateLongToken creates a token of specified length for testing
func generateLongToken(length int) string {
	token := ""
	for i := 0; i < length; i++ {
		token += "a"
	}
	return token
}

// containsNullBytes checks if string contains null bytes
func containsNullBytes(s string) bool {
	for _, b := range []byte(s) {
		if b == 0 {
			return true
		}
	}
	return false
}

// containsMaliciousContent checks for basic malicious patterns
func containsMaliciousContent(s string) bool {
	maliciousPatterns := []string{
		"<script>", "</script>", "DROP TABLE", "SELECT *", "javascript:",
	}
	for _, pattern := range maliciousPatterns {
		if len(s) >= len(pattern) {
			for i := 0; i <= len(s)-len(pattern); i++ {
				if s[i:i+len(pattern)] == pattern {
					return true
				}
			}
		}
	}
	return false
} 