package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetIDFromPath(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		prefix      string
		suffix      string
		expectedID  string
		expectError bool
	}{
		{
			name:        "Valid path with suffix",
			path:        "/api/newsletters/123/subscribe",
			prefix:      "/api/newsletters/",
			suffix:      "/subscribe",
			expectedID:  "123",
			expectError: false,
		},
		{
			name:        "Valid path without suffix",
			path:        "/api/newsletters/456",
			prefix:      "/api/newsletters/",
			suffix:      "",
			expectedID:  "456",
			expectError: false,
		},
		{
			name:        "Invalid prefix",
			path:        "/api/posts/123/subscribe",
			prefix:      "/api/newsletters/",
			suffix:      "/subscribe",
			expectedID:  "",
			expectError: true,
		},
		{
			name:        "Invalid suffix",
			path:        "/api/newsletters/123/publish",
			prefix:      "/api/newsletters/",
			suffix:      "/subscribe",
			expectedID:  "",
			expectError: true,
		},
		{
			name:        "Empty ID",
			path:        "/api/newsletters//subscribe",
			prefix:      "/api/newsletters/",
			suffix:      "/subscribe",
			expectedID:  "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			id, err := GetIDFromPath(tt.path, tt.prefix, tt.suffix)
			
			if tt.expectError {
				assert.Error(t, err)
				assert.Empty(t, id)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedID, id)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name        string
		email       string
		expectError bool
	}{
		{
			name:        "Valid email",
			email:       "test@example.com",
			expectError: false,
		},
		{
			name:        "Valid email with subdomain",
			email:       "user@mail.example.com",
			expectError: false,
		},
		{
			name:        "Valid email with numbers",
			email:       "user123@example123.com",
			expectError: false,
		},
		{
			name:        "Empty email",
			email:       "",
			expectError: true,
		},
		{
			name:        "Invalid email - no @",
			email:       "testexample.com",
			expectError: true,
		},
		{
			name:        "Invalid email - no domain",
			email:       "test@",
			expectError: true,
		},
		{
			name:        "Invalid email - no local part",
			email:       "@example.com",
			expectError: true,
		},
		{
			name:        "Invalid email - spaces",
			email:       "test @example.com",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateEmail(tt.email)
			
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
} 