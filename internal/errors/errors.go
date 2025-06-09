// Package errors provides centralized error definitions and HTTP status mapping
// for the Newsletter Service application.
//
// Error Types:
//   - Standard HTTP errors (NotFound, Unauthorized, etc.)
//   - Domain-specific errors (NewsletterNotFound, EditorNotFound, etc.)
//   - Validation errors (NameEmpty, InvalidEmail, etc.)
//   - Business logic errors (AlreadySubscribed, TokenInvalid, etc.)
//
// Usage:
//   - Use Is* functions to check error types: errors.IsNotFound(err)
//   - Use ErrorToHTTPStatus() to map errors to HTTP status codes
//   - Use Wrap* functions for adding context to errors
package errors

import (
	"errors"
	"fmt"
	"net/http"
)

// Standard HTTP errors - these are the base error types
var (
	ErrNotFound      = errors.New("not found") // 404
	ErrUnauthorized  = errors.New("unauthorized") // 401
	ErrForbidden     = errors.New("forbidden") // 403
	ErrConflict      = errors.New("conflict") // 409
	ErrValidation    = errors.New("validation failed") // 400
	ErrInternal      = errors.New("internal server error") // 500
	ErrBadRequest    = errors.New("bad request") // 400
)

// Domain-specific not found errors - wrap ErrNotFound for specific resources
var (
	ErrNewsletterNotFound = fmt.Errorf("%w: newsletter not found", ErrNotFound) // 404
	ErrEditorNotFound     = fmt.Errorf("%w: editor not found", ErrNotFound) // 404
	ErrPostNotFound       = fmt.Errorf("%w: post not found", ErrNotFound) // 404
	ErrSubscriberNotFound = fmt.Errorf("%w: subscriber not found", ErrNotFound) // 404
)

// Validation errors - wrap ErrValidation for specific validation failures
var (
	ErrNameEmpty      = fmt.Errorf("%w: name cannot be empty", ErrValidation) // 400
	ErrInvalidEmail   = fmt.Errorf("%w: invalid email format", ErrValidation) // 400
	ErrContentTooLong = fmt.Errorf("%w: content is too long", ErrValidation) // 400
	ErrTokenInvalid   = fmt.Errorf("%w: token is invalid", ErrValidation) // 400
)

// Business logic errors - wrap appropriate base errors for specific business rules
var (
	ErrAlreadySubscribed     = fmt.Errorf("%w: already subscribed", ErrConflict) // 409
	ErrAlreadyConfirmed      = fmt.Errorf("%w: already confirmed", ErrConflict) // 409
	ErrSubscriptionNotFound  = fmt.Errorf("%w: subscription not found", ErrNotFound) // 404
	ErrInvalidOrExpiredToken = fmt.Errorf("%w: invalid or expired token", ErrUnauthorized) // 401
)

// Error wrapping functions provide consistent error context formatting

// WrapNotFound wraps an error with a resource-specific not found message.
// If err is nil, returns a simple not found error for the resource.
func WrapNotFound(err error, resource string) error {
	if err == nil {
		return fmt.Errorf("%s %w", resource, ErrNotFound)
	}
	return fmt.Errorf("%s %w: %w", resource, ErrNotFound, err)
}

// WrapConflict wraps an error with a resource-specific conflict message.
// If err is nil, returns a simple conflict error for the resource.
func WrapConflict(err error, resource string) error {
	if err == nil {
		return fmt.Errorf("%s %w", resource, ErrConflict)
	}
	return fmt.Errorf("%s %w: %w", resource, ErrConflict, err)
}

// WrapValidation wraps an error with a validation context message.
// If err is nil, returns a simple validation error with the message.
func WrapValidation(err error, message string) error {
	if err == nil {
		return fmt.Errorf("%w: %s", ErrValidation, message)
	}
	return fmt.Errorf("%w: %s: %w", ErrValidation, message, err)
}

// Error type checking functions - use these instead of direct error comparison

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

// HTTP status mapping - converts application errors to appropriate HTTP status codes

// ErrorToHTTPStatus maps an error to an appropriate HTTP status code.
// This is used by HTTP handlers to return consistent status codes for different error types.
//
// Status code mappings:
//   - NotFound -> 404 Not Found
//   - Unauthorized -> 401 Unauthorized
//   - Forbidden -> 403 Forbidden
//   - Conflict -> 409 Conflict
//   - Validation -> 400 Bad Request
//   - BadRequest -> 400 Bad Request
//   - All others -> 500 Internal Server Error
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
	case IsValidation(err), IsBadRequest(err):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
} 