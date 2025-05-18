package models

import (
	"time"

	"github.com/google/uuid"
)

type Post struct {
	ID           uuid.UUID  `json:"id"`
	NewsletterID uuid.UUID  `json:"newsletter_id"`
	Title        string     `json:"title"`
	Content      string     `json:"content"`
	PublishedAt  *time.Time `json:"published_at,omitempty"` // Pointer for nullability
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
