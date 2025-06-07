package models

import (
	"strings"
	"time"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
)

// Post represents the domain model for a blog post within a newsletter
type Post struct {
	ID           string     `json:"id"`
	NewsletterID string     `json:"newsletterId"`
	Title        string     `json:"title"`
	Content      string     `json:"content"`
	PublishedAt  *time.Time `json:"publishedAt,omitempty"` // Pointer for nullability
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// Validate performs basic business validation on the Post fields
func (p *Post) Validate() error {
	if strings.TrimSpace(p.ID) == "" {
		return apperrors.WrapValidation(nil, "post ID is required")
	}
	
	if strings.TrimSpace(p.NewsletterID) == "" {
		return apperrors.WrapValidation(nil, "post newsletter ID is required")
	}
	
	if strings.TrimSpace(p.Title) == "" {
		return apperrors.WrapValidation(nil, "post title is required")
	}
	
	if strings.TrimSpace(p.Content) == "" {
		return apperrors.WrapValidation(nil, "post content is required")
	}
	
	return nil
}

// IsPublished checks if the post has a publication timestamp set
func (p *Post) IsPublished() bool {
	return p.PublishedAt != nil
}
