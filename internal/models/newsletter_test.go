package models

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
)

func TestNewsletter_Validate(t *testing.T) {
	tests := []struct {
		name          string
		newsletter    Newsletter
		expectedError string
		shouldFail    bool
	}{
		{
			name: "valid newsletter",
			newsletter: Newsletter{
				ID:          "newsletter_123",
				EditorID:    "editor_456",
				Name:        "Tech Weekly",
				Description: "A weekly tech newsletter",
			},
			shouldFail: false,
		},
		{
			name: "valid newsletter with minimal name",
			newsletter: Newsletter{
				ID:       "newsletter_123",
				EditorID: "editor_456",
				Name:     "ABC", // 3 characters - minimum
			},
			shouldFail: false,
		},
		{
			name: "valid newsletter with maximum length name",
			newsletter: Newsletter{
				ID:       "newsletter_123",
				EditorID: "editor_456",
				Name:     strings.Repeat("A", 100), // 100 characters - maximum
			},
			shouldFail: false,
		},
		{
			name: "valid newsletter with maximum length description",
			newsletter: Newsletter{
				ID:          "newsletter_123",
				EditorID:    "editor_456",
				Name:        "Tech Newsletter",
				Description: strings.Repeat("A", 500), // 500 characters - maximum
			},
			shouldFail: false,
		},
		{
			name: "valid newsletter with empty description",
			newsletter: Newsletter{
				ID:          "newsletter_123",
				EditorID:    "editor_456",
				Name:        "Tech Newsletter",
				Description: "", // Empty description is allowed
			},
			shouldFail: false,
		},
		{
			name: "invalid newsletter - empty ID",
			newsletter: Newsletter{
				ID:       "",
				EditorID: "editor_456",
				Name:     "Tech Newsletter",
			},
			expectedError: "newsletter ID is required",
			shouldFail:    true,
		},
		{
			name: "invalid newsletter - whitespace-only ID",
			newsletter: Newsletter{
				ID:       "   ",
				EditorID: "editor_456",
				Name:     "Tech Newsletter",
			},
			expectedError: "newsletter ID is required",
			shouldFail:    true,
		},
		{
			name: "invalid newsletter - empty editor ID",
			newsletter: Newsletter{
				ID:       "newsletter_123",
				EditorID: "",
				Name:     "Tech Newsletter",
			},
			expectedError: "editor ID is required for newsletter",
			shouldFail:    true,
		},
		{
			name: "invalid newsletter - whitespace-only editor ID",
			newsletter: Newsletter{
				ID:       "newsletter_123",
				EditorID: "   ",
				Name:     "Tech Newsletter",
			},
			expectedError: "editor ID is required for newsletter",
			shouldFail:    true,
		},
		{
			name: "invalid newsletter - empty name",
			newsletter: Newsletter{
				ID:       "newsletter_123",
				EditorID: "editor_456",
				Name:     "",
			},
			expectedError: "name cannot be empty",
			shouldFail:    true,
		},
		{
			name: "invalid newsletter - whitespace-only name",
			newsletter: Newsletter{
				ID:       "newsletter_123",
				EditorID: "editor_456",
				Name:     "   ",
			},
			expectedError: "name cannot be empty",
			shouldFail:    true,
		},
		{
			name: "invalid newsletter - name too short",
			newsletter: Newsletter{
				ID:       "newsletter_123",
				EditorID: "editor_456",
				Name:     "AB", // 2 characters - below minimum of 3
			},
			expectedError: "newsletter name must be at least 3 characters",
			shouldFail:    true,
		},
		{
			name: "invalid newsletter - name too long",
			newsletter: Newsletter{
				ID:       "newsletter_123",
				EditorID: "editor_456",
				Name:     strings.Repeat("A", 101), // 101 characters - exceeds maximum of 100
			},
			expectedError: "newsletter name exceeds maximum length of 100 characters",
			shouldFail:    true,
		},
		{
			name: "invalid newsletter - description too long",
			newsletter: Newsletter{
				ID:          "newsletter_123",
				EditorID:    "editor_456",
				Name:        "Tech Newsletter",
				Description: strings.Repeat("A", 501), // 501 characters - exceeds maximum of 500
			},
			expectedError: "newsletter description exceeds maximum length of 500 characters",
			shouldFail:    true,
		},
		{
			name: "invalid newsletter - name with only whitespace padding",
			newsletter: Newsletter{
				ID:       "newsletter_123",
				EditorID: "editor_456",
				Name:     "  A  ", // Should be trimmed to "A" which is too short
			},
			expectedError: "newsletter name must be at least 3 characters",
			shouldFail:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.newsletter.Validate()

			if tt.shouldFail {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
				
				// Verify error types for known validation errors
				switch {
				case strings.Contains(tt.expectedError, "name cannot be empty"):
					assert.True(t, apperrors.IsValidation(err))
				case strings.Contains(tt.expectedError, "ID is required"):
					assert.True(t, apperrors.IsValidation(err))
				case strings.Contains(tt.expectedError, "editor ID is required"):
					assert.True(t, apperrors.IsValidation(err))
				case strings.Contains(tt.expectedError, "exceeds maximum length"):
					assert.True(t, apperrors.IsValidation(err))
				case strings.Contains(tt.expectedError, "must be at least"):
					assert.True(t, apperrors.IsValidation(err))
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestNewsletter_Validate_EdgeCases(t *testing.T) {
	t.Run("newsletter with unicode characters in name", func(t *testing.T) {
		newsletter := Newsletter{
			ID:       "newsletter_123",
			EditorID: "editor_456",
			Name:     "📧 Tech Newsletter 🚀", // Unicode characters
		}
		err := newsletter.Validate()
		assert.NoError(t, err)
	})

	t.Run("newsletter with newlines in description", func(t *testing.T) {
		newsletter := Newsletter{
			ID:          "newsletter_123",
			EditorID:    "editor_456",
			Name:        "Tech Newsletter",
			Description: "Line 1\nLine 2\nLine 3", // Newlines in description
		}
		err := newsletter.Validate()
		assert.NoError(t, err)
	})

	t.Run("newsletter with special characters in name", func(t *testing.T) {
		newsletter := Newsletter{
			ID:       "newsletter_123",
			EditorID: "editor_456",
			Name:     "Tech & AI Newsletter!", // Special characters
		}
		err := newsletter.Validate()
		assert.NoError(t, err)
	})

	t.Run("newsletter name length exactly at boundaries", func(t *testing.T) {
		// Test exactly 3 characters (minimum)
		newsletterMin := Newsletter{
			ID:       "newsletter_123",
			EditorID: "editor_456",
			Name:     "xyz", // Exactly 3 characters
		}
		assert.NoError(t, newsletterMin.Validate())

		// Test exactly 100 characters (maximum)
		newsletterMax := Newsletter{
			ID:       "newsletter_123",
			EditorID: "editor_456",
			Name:     strings.Repeat("A", 100), // Exactly 100 characters
		}
		assert.NoError(t, newsletterMax.Validate())
	})

	t.Run("newsletter description length exactly at boundary", func(t *testing.T) {
		// Test exactly 500 characters (maximum)
		newsletter := Newsletter{
			ID:          "newsletter_123",
			EditorID:    "editor_456",
			Name:        "Tech Newsletter",
			Description: strings.Repeat("A", 500), // Exactly 500 characters
		}
		assert.NoError(t, newsletter.Validate())
	})
} 