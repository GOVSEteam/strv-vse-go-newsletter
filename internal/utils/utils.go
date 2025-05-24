package utils

import (
	"errors"
	"regexp"
	"strings"
)

// GetIDFromPath extracts an ID from a URL path based on a prefix and suffix.
// Example: path="/api/newsletters/123/subscribe", prefix="/api/newsletters/", suffix="/subscribe" -> returns "123"
func GetIDFromPath(path, prefix, suffix string) (string, error) {
	if !strings.HasPrefix(path, prefix) {
		return "", errors.New("path does not have expected prefix: " + prefix)
	}
	trimmedPath := strings.TrimPrefix(path, prefix)

	if suffix != "" {
		if !strings.HasSuffix(trimmedPath, suffix) {
			return "", errors.New("path does not have expected suffix: " + suffix)
		}
		trimmedPath = strings.TrimSuffix(trimmedPath, suffix)
	}
	
	if trimmedPath == "" {
		return "", errors.New("extracted ID is empty")
	}
	return trimmedPath, nil
}

// ValidateEmail checks if the email format is valid.
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email cannot be empty")
	}
	// A very basic email validation regex. For production, consider a more robust library.
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}
	return nil
}
