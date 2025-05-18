package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"github.com/google/uuid"
)

// ErrNewsletterNameTaken is returned when a newsletter name is already in use by an editor.
var ErrNewsletterNameTaken = errors.New("newsletter name already taken by this editor")
var ErrPostNotFound = errors.New("post not found")
var ErrServiceNewsletterNotFound = errors.New("newsletter not found in newsletter service") // Renamed
var ErrForbidden = errors.New("forbidden: editor does not own this resource")

type NewsletterServiceInterface interface {
	// Newsletter methods
	ListNewslettersByEditorID(ctx context.Context, editorID string, limit int, offset int) ([]repository.Newsletter, int, error)
	CreateNewsletter(ctx context.Context, editorID, name, description string) (*repository.Newsletter, error)
	UpdateNewsletter(ctx context.Context, newsletterID string, editorID string, name *string, description *string) (*repository.Newsletter, error)
	DeleteNewsletter(ctx context.Context, newsletterID string, editorID string) error
	GetNewsletterByID(ctx context.Context, newsletterID string) (*repository.Newsletter, error) // Added for ownership checks

	// Post methods
	CreatePost(ctx context.Context, editorFirebaseUID string, newsletterID uuid.UUID, title string, content string) (*models.Post, error)
	GetPostByID(ctx context.Context, postID uuid.UUID) (*models.Post, error)
	ListPostsByNewsletter(ctx context.Context, newsletterID uuid.UUID, limit int, offset int) ([]*models.Post, int, error)
	UpdatePost(ctx context.Context, editorFirebaseUID string, postID uuid.UUID, title *string, content *string) (*models.Post, error)
	DeletePost(ctx context.Context, editorFirebaseUID string, postID uuid.UUID) error
	MarkPostAsPublished(ctx context.Context, editorFirebaseUID string, postID uuid.UUID) error
}

type newsletterService struct {
	newsletterRepo repository.NewsletterRepository
	postRepo       repository.PostRepository
	editorRepo     repository.EditorRepository // Added for ownership checks
}

func NewNewsletterService(
	newsletterRepo repository.NewsletterRepository,
	postRepo repository.PostRepository,
	editorRepo repository.EditorRepository,
) NewsletterServiceInterface {
	return &newsletterService{
		newsletterRepo: newsletterRepo,
		postRepo:       postRepo,
		editorRepo:     editorRepo,
	}
}

// ListNewsletters is deprecated, use ListNewslettersByEditorID
// func (s *newsletterService) ListNewsletters() ([]repository.Newsletter, error) {
// 	return s.newsletterRepo.ListNewsletters()
// }

// ListNewslettersByEditorID fetches a paginated list of newsletters for a specific editor.
func (s *newsletterService) ListNewslettersByEditorID(ctx context.Context, editorID string, limit int, offset int) ([]repository.Newsletter, int, error) {
	// Add any specific business logic for listing if needed.
	// For now, it's a direct pass-through to the repository.
	return s.newsletterRepo.ListNewslettersByEditorID(editorID, limit, offset)
}

func (s *newsletterService) CreateNewsletter(ctx context.Context, editorID, name, description string) (*repository.Newsletter, error) {
	// Check for name uniqueness
	existing, err := s.newsletterRepo.GetNewsletterByNameAndEditorID(name, editorID)
	if err != nil {
		return nil, err // DB error during check
	}
	if existing != nil {
		return nil, ErrNewsletterNameTaken
	}
	return s.newsletterRepo.CreateNewsletter(editorID, name, description)
}

// UpdateNewsletter handles the business logic for updating a newsletter.
func (s *newsletterService) UpdateNewsletter(ctx context.Context, newsletterID string, editorID string, name *string, description *string) (*repository.Newsletter, error) {
	// If name is being updated, check for uniqueness
	if name != nil && *name != "" { // Also ensure name is not being set to empty if provided
		existingWithNewName, err := s.newsletterRepo.GetNewsletterByNameAndEditorID(*name, editorID)
		if err != nil {
			return nil, err // DB error during check
		}
		// If a newsletter with the new name exists AND it's not the current newsletter being updated
		if existingWithNewName != nil && existingWithNewName.ID != newsletterID {
			return nil, ErrNewsletterNameTaken
		}
	}
	return s.newsletterRepo.UpdateNewsletter(newsletterID, editorID, name, description)
}

// DeleteNewsletter handles the business logic for deleting a newsletter.
func (s *newsletterService) DeleteNewsletter(ctx context.Context, newsletterID string, editorID string) error {
	// The repository's DeleteNewsletter method checks for ownership
	// and returns sql.ErrNoRows if not found/not owned.
	// Note: Deletion of related posts should be handled by ON DELETE CASCADE in the database schema.
	return s.newsletterRepo.DeleteNewsletter(newsletterID, editorID)
}

func (s *newsletterService) GetNewsletterByID(ctx context.Context, newsletterID string) (*repository.Newsletter, error) {
	nl, err := s.newsletterRepo.GetNewsletterByID(newsletterID)
	if err != nil {
		return nil, err
	}
	if nl == nil {
		return nil, ErrServiceNewsletterNotFound
	}
	return nl, nil
}


// --- Post Methods ---

// checkOwnership verifies if the editor (identified by Firebase UID) owns the newsletter.
// It returns the editor's database ID (UUID string) if ownership is confirmed, or an error.
func (s *newsletterService) checkOwnership(ctx context.Context, editorFirebaseUID string, newsletterUUID uuid.UUID) (string, error) {
	editor, err := s.editorRepo.GetEditorByFirebaseUID(editorFirebaseUID)
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrForbidden // Editor not found by Firebase UID
		}
		return "", fmt.Errorf("failed to get editor by firebase UID: %w", err)
	}

	newsletter, err := s.newsletterRepo.GetNewsletterByID(newsletterUUID.String())
	if err != nil {
		if err == sql.ErrNoRows {
			return "", ErrServiceNewsletterNotFound
		}
		return "", fmt.Errorf("failed to get newsletter: %w", err)
	}
	if newsletter == nil { // Should be caught by sql.ErrNoRows above, but as a safeguard
	    return "", ErrServiceNewsletterNotFound
	}


	if newsletter.EditorID != editor.ID {
		return "", ErrForbidden // Editor does not own this newsletter
	}
	return editor.ID, nil // Ownership confirmed, return editor's DB ID
}


func (s *newsletterService) CreatePost(ctx context.Context, editorFirebaseUID string, newsletterID uuid.UUID, title string, content string) (*models.Post, error) {
	_, err := s.checkOwnership(ctx, editorFirebaseUID, newsletterID)
	if err != nil {
		return nil, err
	}

	if title == "" {
		return nil, errors.New("post title cannot be empty")
	}
	if content == "" {
		return nil, errors.New("post content cannot be empty")
	}

	post := &models.Post{
		ID:           uuid.New(), // Repository will use this or DB default
		NewsletterID: newsletterID,
		Title:        title,
		Content:      content,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	createdID, err := s.postRepo.CreatePost(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to create post: %w", err)
	}
	post.ID = createdID // Ensure the post object has the ID returned by the repo

	// Optionally, re-fetch the post to get all DB-generated fields accurately, though CreatePost should return enough.
	// For now, we assume the returned ID is sufficient and other fields are as set.
	return post, nil
}

func (s *newsletterService) GetPostByID(ctx context.Context, postID uuid.UUID) (*models.Post, error) {
	post, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post by ID: %w", err)
	}
	if post == nil {
		return nil, ErrPostNotFound
	}
	return post, nil
}

func (s *newsletterService) ListPostsByNewsletter(ctx context.Context, newsletterID uuid.UUID, limit int, offset int) ([]*models.Post, int, error) {
	// Future: Add ownership check if only owners can list posts of their newsletters.
	// For now, assuming public listing or auth handled at API layer.
	posts, total, err := s.postRepo.ListPostsByNewsletterID(ctx, newsletterID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list posts by newsletter: %w", err)
	}
	return posts, total, nil
}

func (s *newsletterService) UpdatePost(ctx context.Context, editorFirebaseUID string, postID uuid.UUID, title *string, content *string) (*models.Post, error) {
	post, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post for update: %w", err)
	}
	if post == nil {
		return nil, ErrPostNotFound
	}

	_, err = s.checkOwnership(ctx, editorFirebaseUID, post.NewsletterID)
	if err != nil {
		return nil, err
	}

	if title != nil {
		if *title == "" {
			return nil, errors.New("post title, if provided, cannot be empty")
		}
		post.Title = *title
	}
	if content != nil {
		if *content == "" {
			return nil, errors.New("post content, if provided, cannot be empty")
		}
		post.Content = *content
	}
	post.UpdatedAt = time.Now()

	err = s.postRepo.UpdatePost(ctx, post)
	if err != nil {
		return nil, fmt.Errorf("failed to update post: %w", err)
	}
	return post, nil
}

func (s *newsletterService) DeletePost(ctx context.Context, editorFirebaseUID string, postID uuid.UUID) error {
	post, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to get post for deletion: %w", err)
	}
	if post == nil {
		return ErrPostNotFound // Or just let DeletePost in repo return sql.ErrNoRows
	}

	_, err = s.checkOwnership(ctx, editorFirebaseUID, post.NewsletterID)
	if err != nil {
		return err
	}

	err = s.postRepo.DeletePost(ctx, postID)
	if err != nil {
		if err == sql.ErrNoRows { // Should be caught by GetPostByID above, but as safeguard
			return ErrPostNotFound
		}
		return fmt.Errorf("failed to delete post: %w", err)
	}
	return nil
}

func (s *newsletterService) MarkPostAsPublished(ctx context.Context, editorFirebaseUID string, postID uuid.UUID) error {
	post, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		return fmt.Errorf("failed to get post for publishing: %w", err)
	}
	if post == nil {
		return ErrPostNotFound
	}

	_, err = s.checkOwnership(ctx, editorFirebaseUID, post.NewsletterID)
	if err != nil {
		return err
	}

	if post.PublishedAt != nil && !post.PublishedAt.IsZero() {
		return errors.New("post is already published")
	}

	publishedAtTime := time.Now()
	err = s.postRepo.MarkPostAsPublished(ctx, postID, publishedAtTime)
	if err != nil {
		if err == sql.ErrNoRows { // Should be caught by GetPostByID, but safeguard
			return ErrPostNotFound
		}
		return fmt.Errorf("failed to mark post as published: %w", err)
	}
	return nil
}
