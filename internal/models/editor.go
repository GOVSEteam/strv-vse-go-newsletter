package models

import (
	"errors"
	"regexp"
	"time"
)

// EmailRegex defines a basic regex for email validation.
var EmailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Editor represents the domain model for an editor.
// It contains no database or persistence-specific tags.
type Editor struct {
	ID          string    `json:"id"`
	FirebaseUID string    `json:"firebaseUid"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Validate performs basic validation on the Editor fields.
// This can be extended with more specific business rules.
func (e *Editor) Validate() error {
	if e.ID == "" {
		return errors.New("editor ID is required")
	}
	if e.FirebaseUID == "" {
		return errors.New("editor FirebaseUID is required")
	}
	if e.Email == "" {
		return errors.New("editor email is required")
	}
	if !EmailRegex.MatchString(e.Email) {
		return errors.New("invalid email format")
	}
	// Add other validation rules as needed
	return nil
} 