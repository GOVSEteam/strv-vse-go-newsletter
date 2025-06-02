package models

import (
	"errors"
	"time"
)

// Post represents the domain model for a blog post within a newsletter.
// It contains no database or persistence-specific tags.
type Post struct {
	ID           string     `json:"id"`
	NewsletterID string     `json:"newsletterId"`
	Title        string     `json:"title"`
	Content      string     `json:"content"`
	PublishedAt  *time.Time `json:"publishedAt,omitempty"` // Pointer for nullability
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

const (
	MaxPostTitleLength   = 255
	MaxPostContentLength = 65535 // Roughly a TEXT type limit
)

// Validate performs basic validation on the Post fields.
func (p *Post) Validate() error {
	if p.ID == "" {
		return errors.New("post ID is required")
	}
	if p.NewsletterID == "" {
		return errors.New("post newsletter ID is required")
	}
	if p.Title == "" {
		return errors.New("post title is required")
	}
	if len(p.Title) > MaxPostTitleLength {
		return errors.New("post title exceeds maximum length")
	}
	if p.Content == "" {
		return errors.New("post content is required")
	}
	if len(p.Content) > MaxPostContentLength {
		return errors.New("post content exceeds maximum length")
	}
	// Add other validation rules as needed
	return nil
}

// IsPublished checks if the post is considered published.
// A post is published if PublishedAt is not nil and the time is not in the future.
// For simplicity here, we just check if PublishedAt is not nil.
// Business logic could choose to check if PublishedAt.After(time.Now()) for scheduled posts.
func (p *Post) IsPublished() bool {
	return p.PublishedAt != nil
}
