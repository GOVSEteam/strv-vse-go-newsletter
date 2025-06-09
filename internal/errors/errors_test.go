package errors

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrorToHTTPStatus(t *testing.T) {
	tests := []struct {
		name           string
		err            error
		expectedStatus int
	}{
		{
			name:           "not found error",
			err:            ErrNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "newsletter not found error",
			err:            ErrNewsletterNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "editor not found error",
			err:            ErrEditorNotFound,
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "unauthorized error",
			err:            ErrUnauthorized,
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "forbidden error",
			err:            ErrForbidden,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "conflict error",
			err:            ErrConflict,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "already subscribed error",
			err:            ErrAlreadySubscribed,
			expectedStatus: http.StatusConflict,
		},
		{
			name:           "validation error",
			err:            ErrValidation,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "name empty validation error",
			err:            ErrNameEmpty,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "bad request error",
			err:            ErrBadRequest,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "internal server error",
			err:            ErrInternal,
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "unknown error defaults to internal server error",
			err:            errors.New("some unknown error"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "wrapped not found error",
			err:            fmt.Errorf("operation failed: %w", ErrNotFound),
			expectedStatus: http.StatusNotFound,
		},
		{
			name:           "deeply wrapped validation error",
			err:            fmt.Errorf("handler error: %w", fmt.Errorf("service error: %w", ErrValidation)),
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status := ErrorToHTTPStatus(tt.err)
			assert.Equal(t, tt.expectedStatus, status)
		})
	}
}

func TestErrorTypeChecking(t *testing.T) {
	t.Run("IsNotFound", func(t *testing.T) {
		tests := []struct {
			name     string
			err      error
			expected bool
		}{
			{"base not found error", ErrNotFound, true},
			{"newsletter not found error", ErrNewsletterNotFound, true},
			{"wrapped not found error", fmt.Errorf("failed: %w", ErrNotFound), true},
			{"validation error", ErrValidation, false},
			{"nil error", nil, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := IsNotFound(tt.err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("IsValidation", func(t *testing.T) {
		tests := []struct {
			name     string
			err      error
			expected bool
		}{
			{"base validation error", ErrValidation, true},
			{"name empty validation error", ErrNameEmpty, true},
			{"wrapped validation error", fmt.Errorf("service error: %w", ErrValidation), true},
			{"not found error", ErrNotFound, false},
			{"nil error", nil, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := IsValidation(tt.err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("IsConflict", func(t *testing.T) {
		tests := []struct {
			name     string
			err      error
			expected bool
		}{
			{"base conflict error", ErrConflict, true},
			{"already subscribed error", ErrAlreadySubscribed, true},
			{"wrapped conflict error", fmt.Errorf("repo error: %w", ErrConflict), true},
			{"validation error", ErrValidation, false},
			{"nil error", nil, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := IsConflict(tt.err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("IsUnauthorized", func(t *testing.T) {
		tests := []struct {
			name     string
			err      error
			expected bool
		}{
			{"base unauthorized error", ErrUnauthorized, true},
			{"invalid token error", ErrInvalidOrExpiredToken, true},
			{"wrapped unauthorized error", fmt.Errorf("auth error: %w", ErrUnauthorized), true},
			{"forbidden error", ErrForbidden, false},
			{"nil error", nil, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := IsUnauthorized(tt.err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})

	t.Run("IsForbidden", func(t *testing.T) {
		tests := []struct {
			name     string
			err      error
			expected bool
		}{
			{"base forbidden error", ErrForbidden, true},
			{"wrapped forbidden error", fmt.Errorf("middleware error: %w", ErrForbidden), true},
			{"unauthorized error", ErrUnauthorized, false},
			{"nil error", nil, false},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := IsForbidden(tt.err)
				assert.Equal(t, tt.expected, result)
			})
		}
	})
}

func TestErrorWrapping(t *testing.T) {
	t.Run("WrapNotFound", func(t *testing.T) {
		tests := []struct {
			name         string
			err          error
			resource     string
			expectedText string
		}{
			{
				name:         "wrap nil error",
				err:          nil,
				resource:     "newsletter",
				expectedText: "newsletter not found",
			},
			{
				name:         "wrap existing error",
				err:          errors.New("database connection failed"),
				resource:     "user",
				expectedText: "user not found: database connection failed",
			},
			{
				name:         "wrap with empty resource",
				err:          nil,
				resource:     "",
				expectedText: " not found",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := WrapNotFound(tt.err, tt.resource)
				assert.Error(t, result)
				assert.Contains(t, result.Error(), tt.expectedText)
				assert.True(t, IsNotFound(result))
			})
		}
	})

	t.Run("WrapConflict", func(t *testing.T) {
		tests := []struct {
			name         string
			err          error
			resource     string
			expectedText string
		}{
			{
				name:         "wrap nil error",
				err:          nil,
				resource:     "email",
				expectedText: "email conflict",
			},
			{
				name:         "wrap existing error",
				err:          errors.New("unique constraint violation"),
				resource:     "username",
				expectedText: "username conflict: unique constraint violation",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := WrapConflict(tt.err, tt.resource)
				assert.Error(t, result)
				assert.Contains(t, result.Error(), tt.expectedText)
				assert.True(t, IsConflict(result))
			})
		}
	})

	t.Run("WrapValidation", func(t *testing.T) {
		tests := []struct {
			name         string
			err          error
			message      string
			expectedText string
		}{
			{
				name:         "wrap nil error with message",
				err:          nil,
				message:      "field is required",
				expectedText: "validation failed: field is required",
			},
			{
				name:         "wrap existing error with message",
				err:          errors.New("parsing failed"),
				message:      "invalid format",
				expectedText: "validation failed: invalid format: parsing failed",
			},
			{
				name:         "wrap with empty message",
				err:          nil,
				message:      "",
				expectedText: "validation failed: ",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				result := WrapValidation(tt.err, tt.message)
				assert.Error(t, result)
				assert.Contains(t, result.Error(), tt.expectedText)
				assert.True(t, IsValidation(result))
			})
		}
	})
}

func TestDomainSpecificErrors(t *testing.T) {
	t.Run("newsletter specific errors", func(t *testing.T) {
		// Test that domain-specific errors wrap the base errors correctly
		assert.True(t, IsNotFound(ErrNewsletterNotFound))
		assert.False(t, IsValidation(ErrNewsletterNotFound))
		assert.Contains(t, ErrNewsletterNotFound.Error(), "newsletter not found")
	})

	t.Run("validation specific errors", func(t *testing.T) {
		assert.True(t, IsValidation(ErrNameEmpty))
		assert.True(t, IsValidation(ErrInvalidEmail))
		assert.False(t, IsNotFound(ErrNameEmpty))
		assert.Contains(t, ErrNameEmpty.Error(), "name cannot be empty")
	})

	t.Run("business logic errors", func(t *testing.T) {
		assert.True(t, IsConflict(ErrAlreadySubscribed))
		assert.True(t, IsUnauthorized(ErrInvalidOrExpiredToken))
		assert.Contains(t, ErrAlreadySubscribed.Error(), "already subscribed")
	})
}

func TestErrorChaining(t *testing.T) {
	t.Run("complex error chain", func(t *testing.T) {
		// Simulate a complex error chain: repository -> service -> handler
		repoErr := fmt.Errorf("database query failed: %w", ErrNotFound)
		serviceErr := fmt.Errorf("newsletter service: failed to get newsletter: %w", repoErr)
		handlerErr := fmt.Errorf("handler: %w", serviceErr)

		// Should still be identified as not found
		assert.True(t, IsNotFound(handlerErr))
		assert.Equal(t, http.StatusNotFound, ErrorToHTTPStatus(handlerErr))
		
		// Should contain all error messages
		errMsg := handlerErr.Error()
		assert.Contains(t, errMsg, "handler")
		assert.Contains(t, errMsg, "newsletter service")
		assert.Contains(t, errMsg, "database query failed")
		assert.Contains(t, errMsg, "not found")
	})

	t.Run("error unwrapping works correctly", func(t *testing.T) {
		originalErr := errors.New("original error")
		wrappedErr := fmt.Errorf("wrapped: %w", originalErr)
		
		// errors.Is should work
		assert.True(t, errors.Is(wrappedErr, originalErr))
		
		// errors.Unwrap should work
		unwrapped := errors.Unwrap(wrappedErr)
		assert.Equal(t, originalErr, unwrapped)
	})
} 