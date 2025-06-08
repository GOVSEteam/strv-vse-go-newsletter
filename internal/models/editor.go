package models

import (
	"net/mail"
	"strings"
	"time"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
)

// Editor represents the domain model for an editor/user
type Editor struct {
	ID          string    `json:"id"`
	FirebaseUID string    `json:"firebase_uid"`
	Email       string    `json:"email"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Validate performs business validation on the Editor fields
func (e *Editor) Validate() error {
	if strings.TrimSpace(e.ID) == "" {
		return apperrors.WrapValidation(nil, "editor ID is required")
	}
	
	if strings.TrimSpace(e.FirebaseUID) == "" {
		return apperrors.WrapValidation(nil, "editor Firebase UID is required")
	}
	
	trimmedEmail := strings.TrimSpace(e.Email)
	if trimmedEmail == "" {
		return apperrors.WrapValidation(nil, "editor email is required")
	}
	if _, err := mail.ParseAddress(trimmedEmail); err != nil {
		return apperrors.ErrInvalidEmail
	}
	
	return nil
} 