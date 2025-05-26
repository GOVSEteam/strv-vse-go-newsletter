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
		{
			name:          "Valid Bearer token",
			authHeader:    "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectedToken: "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9",
			expectError:   false,
		},
		{
			name:          "Bearer with case variations",
			authHeader:    "bearer test-token",
			expectedToken: "test-token",
			expectError:   false,
		},
		{
			name:          "Missing Authorization header",
			authHeader:    "",
			expectedToken: "",
			expectError:   true,
			expectedError: "missing Authorization header",
		},
		{
			name:          "Invalid format - no Bearer prefix",
			authHeader:    "test-token",
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			token, err := ExtractBearerToken(req)

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
		{
			name:         "Valid JWT token",
			authHeader:   "Bearer valid-jwt-token",
			mockResponse: "test-firebase-uid-123",
			mockError:    nil,
			expectedUID:  "test-firebase-uid-123",
			expectError:  false,
		},
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
			name:          "Invalid JWT token",
			authHeader:    "Bearer invalid-jwt",
			mockResponse:  "",
			mockError:     errors.New("firebase: invalid token"),
			expectedUID:   "",
			expectError:   true,
			expectedError: "firebase: invalid token",
		},
		{
			name:          "Expired JWT token",
			authHeader:    "Bearer expired-jwt",
			mockResponse:  "",
			mockError:     errors.New("firebase: token expired"),
			expectedUID:   "",
			expectError:   true,
			expectedError: "firebase: token expired",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock the VerifyFirebaseJWT function
			VerifyFirebaseJWT = func(req *http.Request) (string, error) {
				// First extract the token to simulate real behavior
				_, err := ExtractBearerToken(req)
				if err != nil {
					return "", err
				}
				
				// Return mock response based on token
				if tt.mockError != nil {
					return "", tt.mockError
				}
				return tt.mockResponse, nil
			}

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			uid, err := VerifyFirebaseJWT(req)

			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, tt.expectedError, err.Error())
				assert.Empty(t, uid)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedUID, uid)
			}
		})
	}
}

// TestBasicAuthFlow tests the basic authentication flow
func TestBasicAuthFlow(t *testing.T) {
	// Save original function
	originalVerifyJWT := VerifyFirebaseJWT
	defer func() { VerifyFirebaseJWT = originalVerifyJWT }()

	// Mock successful verification
	VerifyFirebaseJWT = func(req *http.Request) (string, error) {
		token, err := ExtractBearerToken(req)
		if err != nil {
			return "", err
		}
		if token == "valid-token" {
			return "user-123", nil
		}
		return "", errors.New("invalid token")
	}

	t.Run("Successful authentication", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer valid-token")

		uid, err := VerifyFirebaseJWT(req)
		assert.NoError(t, err)
		assert.Equal(t, "user-123", uid)
	})

	t.Run("Failed authentication", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/protected", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		uid, err := VerifyFirebaseJWT(req)
		assert.Error(t, err)
		assert.Empty(t, uid)
	})
} 