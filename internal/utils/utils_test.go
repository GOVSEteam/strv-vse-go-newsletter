package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetIDFromPath tests the GetIDFromPath function with various scenarios
func TestGetIDFromPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		prefix      string
		suffix      string
		expectedID  string
		expectError bool
		errorMsg    string
	}{
		// Success cases - Real usage patterns from the codebase
		{
			name:        "Newsletter subscription path",
			path:        "/api/newsletters/123/subscribe",
			prefix:      "/api/newsletters/",
			suffix:      "/subscribe",
			expectedID:  "123",
			expectError: false,
		},
		{
			name:        "Newsletter subscribers list path",
			path:        "/api/newsletters/456/subscribers",
			prefix:      "/api/newsletters/",
			suffix:      "/subscribers",
			expectedID:  "456",
			expectError: false,
		},
		{
			name:        "Post publish path",
			path:        "/api/posts/789/publish",
			prefix:      "/api/posts/",
			suffix:      "/publish",
			expectedID:  "789",
			expectError: false,
		},
		{
			name:        "UUID format ID",
			path:        "/api/newsletters/550e8400-e29b-41d4-a716-446655440000/subscribe",
			prefix:      "/api/newsletters/",
			suffix:      "/subscribe",
			expectedID:  "550e8400-e29b-41d4-a716-446655440000",
			expectError: false,
		},
		{
			name:        "No suffix (empty suffix)",
			path:        "/api/items/abc123",
			prefix:      "/api/items/",
			suffix:      "",
			expectedID:  "abc123",
			expectError: false,
		},
		{
			name:        "Complex ID with dashes and numbers",
			path:        "/api/newsletters/newsletter-123-test/subscribe",
			prefix:      "/api/newsletters/",
			suffix:      "/subscribe",
			expectedID:  "newsletter-123-test",
			expectError: false,
		},

		// Error cases - Missing prefix
		{
			name:        "Wrong prefix",
			path:        "/api/posts/123/subscribe",
			prefix:      "/api/newsletters/",
			suffix:      "/subscribe",
			expectedID:  "",
			expectError: true,
			errorMsg:    "path does not have expected prefix: /api/newsletters/",
		},
		{
			name:        "No prefix match",
			path:        "/different/path/123/subscribe",
			prefix:      "/api/newsletters/",
			suffix:      "/subscribe",
			expectedID:  "",
			expectError: true,
			errorMsg:    "path does not have expected prefix: /api/newsletters/",
		},

		// Error cases - Missing suffix
		{
			name:        "Wrong suffix",
			path:        "/api/newsletters/123/publish",
			prefix:      "/api/newsletters/",
			suffix:      "/subscribe",
			expectedID:  "",
			expectError: true,
			errorMsg:    "path does not have expected suffix: /subscribe",
		},
		{
			name:        "No suffix match",
			path:        "/api/newsletters/123/different",
			prefix:      "/api/newsletters/",
			suffix:      "/subscribe",
			expectedID:  "",
			expectError: true,
			errorMsg:    "path does not have expected suffix: /subscribe",
		},

		// Error cases - Empty ID
		{
			name:        "Empty ID between prefix and suffix",
			path:        "/api/newsletters//subscribe",
			prefix:      "/api/newsletters/",
			suffix:      "/subscribe",
			expectedID:  "",
			expectError: true,
			errorMsg:    "extracted ID is empty",
		},
		{
			name:        "Only prefix and suffix, no ID",
			path:        "/api/newsletters/subscribe",
			prefix:      "/api/newsletters/",
			suffix:      "subscribe",
			expectedID:  "",
			expectError: true,
			errorMsg:    "extracted ID is empty",
		},

		// Edge cases
		{
			name:        "Empty path",
			path:        "",
			prefix:      "/api/newsletters/",
			suffix:      "/subscribe",
			expectedID:  "",
			expectError: true,
			errorMsg:    "path does not have expected prefix: /api/newsletters/",
		},
		{
			name:        "Path equals prefix only",
			path:        "/api/newsletters/",
			prefix:      "/api/newsletters/",
			suffix:      "/subscribe",
			expectedID:  "",
			expectError: true,
			errorMsg:    "path does not have expected suffix: /subscribe",
		},
		{
			name:        "Path with query parameters (should fail)",
			path:        "/api/newsletters/123/subscribe?param=value",
			prefix:      "/api/newsletters/",
			suffix:      "/subscribe",
			expectedID:  "",
			expectError: true,
			errorMsg:    "path does not have expected suffix: /subscribe",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := GetIDFromPath(tt.path, tt.prefix, tt.suffix)

			if tt.expectError {
				assert.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
				assert.Empty(t, result)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, result)
			}
		})
	}
}

// TestValidateEmail tests the ValidateEmail function with various email formats
func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		expectError bool
		errorMsg    string
	}{
		// Valid email formats
		{
			name:        "Standard email",
			email:       "user@example.com",
			expectError: false,
		},
		{
			name:        "Email with subdomain",
			email:       "user@mail.example.com",
			expectError: false,
		},
		{
			name:        "Email with numbers",
			email:       "user123@example123.com",
			expectError: false,
		},
		{
			name:        "Email with dots in local part",
			email:       "user.name@example.com",
			expectError: false,
		},
		{
			name:        "Email with plus sign",
			email:       "user+tag@example.com",
			expectError: false,
		},
		{
			name:        "Email with dash in local part",
			email:       "user-name@example.com",
			expectError: false,
		},
		{
			name:        "Email with underscore",
			email:       "user_name@example.com",
			expectError: false,
		},
		{
			name:        "Email with percentage",
			email:       "user%test@example.com",
			expectError: false,
		},
		{
			name:        "Short domain extension",
			email:       "user@example.co",
			expectError: false,
		},
		{
			name:        "Long domain extension",
			email:       "user@example.info",
			expectError: false,
		},
		{
			name:        "Email with dash in domain",
			email:       "user@my-domain.com",
			expectError: false,
		},

		// Invalid email formats
		{
			name:        "Empty email",
			email:       "",
			expectError: true,
			errorMsg:    "email cannot be empty",
		},
		{
			name:        "Missing @ symbol",
			email:       "userexample.com",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "Missing local part",
			email:       "@example.com",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "Missing domain",
			email:       "user@",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "Missing domain extension",
			email:       "user@example",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "Multiple @ symbols",
			email:       "user@@example.com",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "Space in email",
			email:       "user @example.com",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "Invalid characters",
			email:       "user#@example.com",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "Domain extension too short",
			email:       "user@example.c",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "Domain extension too long",
			email:       "user@example.commmm",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "Uppercase letters (should fail with current regex)",
			email:       "User@Example.Com",
			expectError: true,
			errorMsg:    "invalid email format",
		},
		{
			name:        "Starting with dot",
			email:       ".user@example.com",
			expectError: false,
		},
		{
			name:        "Ending with dot in local part",
			email:       "user.@example.com",
			expectError: false,
		},
		{
			name:        "Double dots",
			email:       "user..name@example.com",
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)

			if tt.expectError {
				require.Error(t, err)
				assert.Equal(t, tt.errorMsg, err.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestGetIDFromPath_RealWorldUsage tests the function with actual usage patterns from the codebase
func TestGetIDFromPath_RealWorldUsage(t *testing.T) {
	t.Run("Newsletter subscription endpoint", func(t *testing.T) {
		// This is how it's used in subscribe.go
		newsletterID, err := GetIDFromPath("/api/newsletters/123/subscribe", "/api/newsletters/", "/subscribe")
		require.NoError(t, err)
		assert.Equal(t, "123", newsletterID)
	})

	t.Run("Newsletter subscribers list endpoint", func(t *testing.T) {
		// This is how it's used in list.go
		newsletterID, err := GetIDFromPath("/api/newsletters/456/subscribers", "/api/newsletters/", "/subscribers")
		require.NoError(t, err)
		assert.Equal(t, "456", newsletterID)
	})

	t.Run("Post publish endpoint", func(t *testing.T) {
		// This is how it's used in publish.go
		postID, err := GetIDFromPath("/api/posts/789/publish", "/api/posts/", "/publish")
		require.NoError(t, err)
		assert.Equal(t, "789", postID)
	})

	t.Run("Error case that handlers expect", func(t *testing.T) {
		// This should produce the error message that subscribe_test.go expects
		_, err := GetIDFromPath("/api/newsletters//subscribe", "/api/newsletters/", "/subscribe")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "extracted ID is empty")
	})
}

// TestValidateEmail_RealWorldUsage tests the function with actual usage patterns from the codebase
func TestValidateEmail_RealWorldUsage(t *testing.T) {
	t.Run("Valid email from subscription", func(t *testing.T) {
		// This is how it's used in subscribe.go
		err := ValidateEmail("user@example.com")
		assert.NoError(t, err)
	})

	t.Run("Empty email error that handlers expect", func(t *testing.T) {
		// This should produce the error message that subscribe_test.go expects
		err := ValidateEmail("")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "email cannot be empty")
	})

	t.Run("Invalid email format error that handlers expect", func(t *testing.T) {
		// This should produce the error message that subscribe_test.go expects
		err := ValidateEmail("invalid-email")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "invalid email format")
	})
} 