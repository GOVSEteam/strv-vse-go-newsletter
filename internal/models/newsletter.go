package models

import (
	"errors"
	"time"
)

// Newsletter represents the domain model for a newsletter.
// It contains no database or persistence-specific tags.
type Newsletter struct {
	ID          string    `json:"id"`
	EditorID    string    `json:"editorId"`
	Name        string    `json:"name"`
	Description string    `json:"description,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// Validate performs basic validation on the Newsletter fields.
// This can be extended with more specific business rules.
func (n *Newsletter) Validate() error {
	if n.ID == "" {
		return errors.New("newsletter ID is required")
	}
	if n.EditorID == "" {
		return errors.New("editor ID is required for newsletter")
	}
	if n.Name == "" {
		return errors.New("newsletter name is required")
	}
	// Add other validation rules as needed, e.g., max length for Name/Description
	return nil
} 