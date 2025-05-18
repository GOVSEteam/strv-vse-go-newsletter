package service

import (
	"errors"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
)

// ErrNewsletterNameTaken is returned when a newsletter name is already in use by an editor.
var ErrNewsletterNameTaken = errors.New("newsletter name already taken by this editor")

type NewsletterServiceInterface interface {
	// ListNewsletters() ([]repository.Newsletter, error) // Old signature
	ListNewslettersByEditorID(editorID string, limit int, offset int) ([]repository.Newsletter, int, error) // New signature
	CreateNewsletter(editorID, name, description string) (*repository.Newsletter, error)
	UpdateNewsletter(newsletterID string, editorID string, name *string, description *string) (*repository.Newsletter, error)
	DeleteNewsletter(newsletterID string, editorID string) error
}

type newsletterService struct {
	repo repository.NewsletterRepository
}

func NewsletterService(repo repository.NewsletterRepository) NewsletterServiceInterface {
	return &newsletterService{repo: repo}
}

// ListNewsletters is deprecated, use ListNewslettersByEditorID
// func (s *newsletterService) ListNewsletters() ([]repository.Newsletter, error) {
// 	return s.repo.ListNewsletters()
// }

// ListNewslettersByEditorID fetches a paginated list of newsletters for a specific editor.
func (s *newsletterService) ListNewslettersByEditorID(editorID string, limit int, offset int) ([]repository.Newsletter, int, error) {
	// Add any specific business logic for listing if needed.
	// For now, it's a direct pass-through to the repository.
	return s.repo.ListNewslettersByEditorID(editorID, limit, offset)
}

func (s *newsletterService) CreateNewsletter(editorID, name, description string) (*repository.Newsletter, error) {
	// Check for name uniqueness
	existing, err := s.repo.GetNewsletterByNameAndEditorID(name, editorID)
	if err != nil {
		return nil, err // DB error during check
	}
	if existing != nil {
		return nil, ErrNewsletterNameTaken
	}
	return s.repo.CreateNewsletter(editorID, name, description)
}

// UpdateNewsletter handles the business logic for updating a newsletter.
func (s *newsletterService) UpdateNewsletter(newsletterID string, editorID string, name *string, description *string) (*repository.Newsletter, error) {
	// If name is being updated, check for uniqueness
	if name != nil && *name != "" { // Also ensure name is not being set to empty if provided
		existingWithNewName, err := s.repo.GetNewsletterByNameAndEditorID(*name, editorID)
		if err != nil {
			return nil, err // DB error during check
		}
		// If a newsletter with the new name exists AND it's not the current newsletter being updated
		if existingWithNewName != nil && existingWithNewName.ID != newsletterID {
			return nil, ErrNewsletterNameTaken
		}
	}

	// Basic validation: if name is provided, it should not be empty (redundant if check above is primary).
	// The CreateHandler and UpdateHandler also check for empty name if provided.
	// if name != nil && *name == "" { 
	// 	 return nil, errors.New("newsletter name, if provided, cannot be empty")
	// }

	return s.repo.UpdateNewsletter(newsletterID, editorID, name, description)
}

// DeleteNewsletter handles the business logic for deleting a newsletter.
func (s *newsletterService) DeleteNewsletter(newsletterID string, editorID string) error {
	// The repository's DeleteNewsletter method checks for ownership
	// and returns sql.ErrNoRows if not found/not owned.

	// TODO: Implement deletion of related subscribers from Firebase (requires SubscriberRepository from RFC-004).
	// This should be done in a way that considers atomicity or handles potential partial failures.
	// For example:
	// err := s.subscriberRepo.DeleteSubscribersByNewsletterID(newsletterID)
	// if err != nil {
	//     // Log error, decide if newsletter deletion should still proceed or be rolled back if in a transaction.
	//     // For now, we proceed with newsletter deletion even if subscriber deletion fails.
	//     log.Printf("Warning: Failed to delete subscribers for newsletter %s: %v", newsletterID, err)
	// }

	// Note: Deletion of related posts should be handled by ON DELETE CASCADE in the database schema.

	return s.repo.DeleteNewsletter(newsletterID, editorID)
}
