package models

import (
	"strings"
	"time"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
)

// Newsletter represents the domain model for a newsletter
type Newsletter struct {
	ID          string    `json:"id"`
	EditorID    string    `json:"editor_id"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Validate performs business validation on the Newsletter fields
func (n *Newsletter) Validate() error {
	if strings.TrimSpace(n.ID) == "" {
		return apperrors.WrapValidation(nil, "newsletter ID is required")
	}
	
	if strings.TrimSpace(n.EditorID) == "" {
		return apperrors.WrapValidation(nil, "editor ID is required for newsletter")
	}
	
	trimmedName := strings.TrimSpace(n.Name)
	if trimmedName == "" {
		return apperrors.ErrNameEmpty
	}
	if len(trimmedName) < 3 {
		return apperrors.WrapValidation(nil, "newsletter name must be at least 3 characters")
	}
	if len(trimmedName) > 100 {
		return apperrors.WrapValidation(nil, "newsletter name exceeds maximum length of 100 characters")
	}
	
	if len(strings.TrimSpace(n.Description)) > 500 {
		return apperrors.WrapValidation(nil, "newsletter description exceeds maximum length of 500 characters")
	}
	
	return nil
} 