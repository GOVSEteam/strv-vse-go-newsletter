package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Standard application errors
var (
	ErrNotFound         = errors.New("not found")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrForbidden        = errors.New("forbidden")
	ErrConflict         = errors.New("conflict")
	ErrValidation       = errors.New("validation failed")
	ErrInternal         = errors.New("internal server error")
	ErrBadRequest       = errors.New("bad request") // Added for common client errors
	ErrRateLimitExceeded = errors.New("rate limit exceeded") // Added for rate limiting
)

// Repository-specific errors
var (
	ErrNewsletterNotFound  = fmt.Errorf("%w: newsletter not found", ErrNotFound)
	ErrEditorNotFound      = fmt.Errorf("%w: editor not found", ErrNotFound)
	ErrPostNotFound        = fmt.Errorf("%w: post not found", ErrNotFound)
	ErrSubscriberNotFound  = fmt.Errorf("%w: subscriber not found", ErrNotFound)
)

// Service-specific validation errors
var (
	ErrNameEmpty           = fmt.Errorf("%w: name cannot be empty", ErrValidation)
	ErrInvalidEmail        = fmt.Errorf("%w: invalid email format", ErrValidation)
	ErrContentTooLong      = fmt.Errorf("%w: content is too long", ErrValidation)
	ErrAlreadySubscribed   = fmt.Errorf("%w: already subscribed", ErrConflict)
	ErrPasswordTooShort    = fmt.Errorf("%w: password is too short", ErrValidation) // Added for password validation
	ErrTokenInvalid        = fmt.Errorf("%w: token is invalid", ErrValidation)       // Added for token validation
	ErrResourceMismatch    = fmt.Errorf("%w: resource mismatch", ErrValidation)     // Added for ID mismatches etc.
)

// WrapNotFound wraps an error with a resource-specific not found message.
func WrapNotFound(err error, resource string) error {
	if err == nil {
		return fmt.Errorf("%s %w", resource, ErrNotFound)
	}
	return fmt.Errorf("%s %w: %w", resource, ErrNotFound, err)
}

// WrapConflict wraps an error with a resource-specific conflict message.
func WrapConflict(err error, resource string) error {
	if err == nil {
		return fmt.Errorf("%s %w", resource, ErrConflict)
	}
	return fmt.Errorf("%s %w: %w", resource, ErrConflict, err)
}

// Helper functions to check error types

// IsNotFound checks if an error is an ErrNotFound or wraps it.
func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsUnauthorized checks if an error is an ErrUnauthorized or wraps it.
func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized)
}

// IsForbidden checks if an error is an ErrForbidden or wraps it.
func IsForbidden(err error) bool {
	return errors.Is(err, ErrForbidden)
}

// IsConflict checks if an error is an ErrConflict or wraps it.
func IsConflict(err error) bool {
	return errors.Is(err, ErrConflict)
}

// IsValidation checks if an error is an ErrValidation or wraps it.
func IsValidation(err error) bool {
	return errors.Is(err, ErrValidation)
}

// IsInternal checks if an error is an ErrInternal or wraps it.
func IsInternal(err error) bool {
	return errors.Is(err, ErrInternal)
}

// IsBadRequest checks if an error is an ErrBadRequest or wraps it.
func IsBadRequest(err error) bool {
	return errors.Is(err, ErrBadRequest)
}


// ErrorToHTTPStatus maps an error to an HTTP status code.
func ErrorToHTTPStatus(err error) int {
	switch {
	case IsNotFound(err):
		return http.StatusNotFound
	case IsUnauthorized(err):
		return http.StatusUnauthorized
	case IsForbidden(err):
		return http.StatusForbidden
	case IsConflict(err):
		return http.StatusConflict
	case IsValidation(err):
		return http.StatusBadRequest // Or StatusUnprocessableEntity (422) if preferred
	case IsBadRequest(err):
		return http.StatusBadRequest
	case errors.Is(err, ErrRateLimitExceeded):
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
} 