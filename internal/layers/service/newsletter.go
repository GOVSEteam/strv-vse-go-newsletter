package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"github.com/google/uuid"
)

const (
	MaxNewsletterNameLength        = 100
	MaxNewsletterDescriptionLength = 255
	MaxPostTitleLength             = 150
	MinPostContentLength           = 10 // Arbitrary minimum
)

type NewsletterServiceInterface interface {
	// Newsletter methods
	ListNewslettersByEditorID(ctx context.Context, editorID string, limit int, offset int) ([]models.Newsletter, int, error)
	CreateNewsletter(ctx context.Context, editorID, name, description string) (*models.Newsletter, error)
	GetNewsletterByID(ctx context.Context, newsletterID string) (*models.Newsletter, error) // For internal/service use, ownership checked by caller if needed
	GetNewsletterForEditor(ctx context.Context, editorID, newsletterID string) (*models.Newsletter, error) // For editor-specific get with ownership
	UpdateNewsletter(ctx context.Context, editorID string, newsletterID string, name *string, description *string) (*models.Newsletter, error)
	DeleteNewsletter(ctx context.Context, editorID string, newsletterID string) error

	// Post methods
	CreatePost(ctx context.Context, editorID string, newsletterID string, title string, content string) (*models.Post, error)
	GetPostByID(ctx context.Context, postID string) (*models.Post, error) // General get, ownership might be checked by caller
	GetPostForEditor(ctx context.Context, editorID string, postID string) (*models.Post, error) // For editor-specific get with ownership of post's newsletter
	ListPostsByNewsletterID(ctx context.Context, newsletterID string, limit int, offset int) ([]models.Post, int, error)
	UpdatePost(ctx context.Context, editorID string, postID string, title *string, content *string) (*models.Post, error)
	DeletePost(ctx context.Context, editorID string, postID string) error
	PublishPost(ctx context.Context, editorID string, postID string) (*models.Post, error)
	UnpublishPost(ctx context.Context, editorID string, postID string) (*models.Post, error)
}

type newsletterService struct {
	newsletterRepo repository.NewsletterRepository
	postRepo       repository.PostRepository
}

func NewNewsletterService(
	newsletterRepo repository.NewsletterRepository,
	postRepo repository.PostRepository,
) NewsletterServiceInterface {
	return &newsletterService{
		newsletterRepo: newsletterRepo,
		postRepo:       postRepo,
	}
}

// --- Authorization Helper ---

// getEditorFromContext retrieves the authenticated editor from context.
// This eliminates the need for additional database queries.
func (s *newsletterService) getEditorFromContext(ctx context.Context) (*models.Editor, error) {
	if editor, ok := ctx.Value("editor").(*models.Editor); ok {
		return editor, nil
	}
	return nil, fmt.Errorf("service: getEditorFromContext: %w", apperrors.ErrForbidden)
}

// verifyNewsletterOwnershipWithEditor verifies newsletter ownership using an already-fetched editor.
// This avoids redundant database calls when the editor is already available.
func (s *newsletterService) verifyNewsletterOwnershipWithEditor(ctx context.Context, editor *models.Editor, newsletterID string) (*models.Newsletter, error) {
	newsletter, err := s.newsletterRepo.GetNewsletterByID(ctx, newsletterID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNewsletterNotFound) {
			return nil, fmt.Errorf("service: verifyNewsletterOwnership: %w", apperrors.ErrNewsletterNotFound)
		}
		return nil, fmt.Errorf("service: verifyNewsletterOwnership: getting newsletter: %w", err) // Internal error
	}

	if newsletter.EditorID != editor.ID {
		return nil, fmt.Errorf("service: verifyNewsletterOwnership: %w", apperrors.ErrForbidden)
	}
	return newsletter, nil
}

// verifyPostOwnershipWithEditor verifies post ownership using an already-fetched editor.
// Returns the post if successful.
func (s *newsletterService) verifyPostOwnershipWithEditor(ctx context.Context, editor *models.Editor, postID string) (*models.Post, error) {
	post, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		if errors.Is(err, apperrors.ErrPostNotFound) {
			return nil, fmt.Errorf("service: verifyPostOwnership: %w", apperrors.ErrPostNotFound)
		}
		return nil, fmt.Errorf("service: verifyPostOwnership: getting post: %w", err)
	}

	// Use the already-fetched editor to verify newsletter ownership (no redundant DB call)
	_, err = s.verifyNewsletterOwnershipWithEditor(ctx, editor, post.NewsletterID)
	if err != nil {
		// This will return ErrForbidden if newsletter not owned, or ErrNewsletterNotFound
		return nil, fmt.Errorf("service: verifyPostOwnership: newsletter ownership check failed: %w", err)
	}

	return post, nil
}

// --- Newsletter Methods ---

func (s *newsletterService) ListNewslettersByEditorID(ctx context.Context, editorID string, limit int, offset int) ([]models.Newsletter, int, error) {
	editor, err := s.getEditorFromContext(ctx)
	if err != nil {
		return nil, 0, err
	}
	
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if offset < 0 {
		offset = 0 // Default offset
	}
	newsletters, total, err := s.newsletterRepo.ListNewslettersByEditorID(ctx, editor.ID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("service: ListNewslettersByEditorID: %w", err) // Wrap internal errors
	}
	return newsletters, total, nil
}

func (s *newsletterService) CreateNewsletter(ctx context.Context, editorID, name, description string) (*models.Newsletter, error) {
	editor, err := s.getEditorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	
	name = strings.TrimSpace(name)
	description = strings.TrimSpace(description)

	if name == "" {
		return nil, fmt.Errorf("service: CreateNewsletter: %w: name cannot be empty", apperrors.ErrValidation)
	}
	if len(name) > MaxNewsletterNameLength {
		return nil, fmt.Errorf("service: CreateNewsletter: %w: name exceeds max length of %d", apperrors.ErrValidation, MaxNewsletterNameLength)
	}
	if len(description) > MaxNewsletterDescriptionLength {
		return nil, fmt.Errorf("service: CreateNewsletter: %w: description exceeds max length of %d", apperrors.ErrValidation, MaxNewsletterDescriptionLength)
	}

	newsletter, err := s.newsletterRepo.CreateNewsletter(ctx, editor.ID, name, description)
	if err != nil {
		if errors.Is(err, apperrors.ErrConflict) {
			// Handle unique constraint violation from database
			return nil, fmt.Errorf("service: CreateNewsletter: %w: newsletter name '%s' is already taken", apperrors.ErrConflict, name)
		}
		return nil, fmt.Errorf("service: CreateNewsletter: %w", err)
	}
	return newsletter, nil
}

func (s *newsletterService) GetNewsletterByID(ctx context.Context, newsletterID string) (*models.Newsletter, error) {
	nl, err := s.newsletterRepo.GetNewsletterByID(ctx, newsletterID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNewsletterNotFound) {
			return nil, fmt.Errorf("service: GetNewsletterByID: %w", apperrors.ErrNewsletterNotFound)
		}
		return nil, fmt.Errorf("service: GetNewsletterByID: %w", err)
	}
	return nl, nil
}

func (s *newsletterService) GetNewsletterForEditor(ctx context.Context, editorID string, newsletterID string) (*models.Newsletter, error) {
	editor, err := s.getEditorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	
	newsletter, err := s.verifyNewsletterOwnershipWithEditor(ctx, editor, newsletterID)
	if err != nil {
		return nil, err
	}
	return newsletter, nil
}

func (s *newsletterService) UpdateNewsletter(ctx context.Context, editorID string, newsletterID string, name *string, description *string) (*models.Newsletter, error) {
	editor, err := s.getEditorFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Validate provided fields before attempting update
	var namePtr *string
	if name != nil {
		trimmedName := strings.TrimSpace(*name)
		if trimmedName == "" {
			return nil, fmt.Errorf("service: UpdateNewsletter: %w: name cannot be empty if provided", apperrors.ErrValidation)
		}
		if len(trimmedName) > MaxNewsletterNameLength {
			return nil, fmt.Errorf("service: UpdateNewsletter: %w: name exceeds max length of %d", apperrors.ErrValidation, MaxNewsletterNameLength)
		}
		namePtr = &trimmedName
	}

	var descPtr *string
	if description != nil {
		trimmedDescription := strings.TrimSpace(*description)
		if len(trimmedDescription) > MaxNewsletterDescriptionLength {
			return nil, fmt.Errorf("service: UpdateNewsletter: %w: description exceeds max length of %d", apperrors.ErrValidation, MaxNewsletterDescriptionLength)
		}
		descPtr = &trimmedDescription
	}

	// Repository atomically handles authorization and update
	updatedNewsletter, err := s.newsletterRepo.UpdateNewsletter(ctx, newsletterID, editor.ID, namePtr, descPtr)
	if err != nil {
		if errors.Is(err, apperrors.ErrNewsletterNotFound) {
			return nil, fmt.Errorf("service: UpdateNewsletter: %w", apperrors.ErrNewsletterNotFound)
		}
		return nil, fmt.Errorf("service: UpdateNewsletter: %w", err)
	}
	return updatedNewsletter, nil
}

func (s *newsletterService) DeleteNewsletter(ctx context.Context, editorID string, newsletterID string) error {
	editor, err := s.getEditorFromContext(ctx)
	if err != nil {
		return err
	}

	// Repository atomically handles authorization and deletion
	err = s.newsletterRepo.DeleteNewsletter(ctx, newsletterID, editor.ID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNewsletterNotFound) {
			return fmt.Errorf("service: DeleteNewsletter: %w", apperrors.ErrNewsletterNotFound)
		}
		return fmt.Errorf("service: DeleteNewsletter: %w", err)
	}
	return nil
}

// --- Post Methods ---

func (s *newsletterService) CreatePost(ctx context.Context, editorID string, newsletterID string, title string, content string) (*models.Post, error) {
	editor, err := s.getEditorFromContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("service: CreatePost: authorization failed: %w", err)
	}
	
	newsletter, err := s.verifyNewsletterOwnershipWithEditor(ctx, editor, newsletterID)
	if err != nil {
		return nil, fmt.Errorf("service: CreatePost: authorization failed: %w", err)
	}

	title = strings.TrimSpace(title)
	content = strings.TrimSpace(content)

	if title == "" {
		return nil, fmt.Errorf("service: CreatePost: %w: title cannot be empty", apperrors.ErrValidation)
	}
	if len(title) > MaxPostTitleLength {
		return nil, fmt.Errorf("service: CreatePost: %w: title exceeds max length of %d", apperrors.ErrValidation, MaxPostTitleLength)
	}
	if content == "" { // Consider minimum length as well
		return nil, fmt.Errorf("service: CreatePost: %w: content cannot be empty", apperrors.ErrValidation)
	}
	if len(content) < MinPostContentLength {
		 return nil, fmt.Errorf("service: CreatePost: %w: content must be at least %d characters", apperrors.ErrValidation, MinPostContentLength)
	}


	post := &models.Post{
		ID:           uuid.NewString(), // Repository expects ID to be set
		NewsletterID: newsletter.ID,    // Use the verified newsletter's ID
		Title:        title,
		Content:      content,
		// PublishedAt is nil by default (not published)
		CreatedAt: time.Now().UTC(),
		UpdatedAt: time.Now().UTC(),
	}

	createdPost, err := s.postRepo.CreatePost(ctx, post)
	if err != nil {
		// The repository CreatePost now handles specific errors like ForeignKeyViolation
		// and wraps them, e.g., into apperrors.ErrNotFound if newsletter_id is bad.
		// Or apperrors.ErrConflict if post ID (if we were to allow client-set IDs) was a duplicate.
		return nil, fmt.Errorf("service: CreatePost: failed to create post: %w", err)
	}
	return createdPost, nil
}

func (s *newsletterService) GetPostByID(ctx context.Context, postID string) (*models.Post, error) {
	// This is a general get, does not check ownership.
	// Useful for public access or when ownership is checked by the caller.
	post, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		if errors.Is(err, apperrors.ErrPostNotFound) {
			return nil, fmt.Errorf("service: GetPostByID: %w", apperrors.ErrPostNotFound)
		}
		return nil, fmt.Errorf("service: GetPostByID: %w", err)
	}
	return post, nil
}

func (s *newsletterService) GetPostForEditor(ctx context.Context, editorID string, postID string) (*models.Post, error) {
	editor, err := s.getEditorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	
	post, err := s.verifyPostOwnershipWithEditor(ctx, editor, postID)
	if err != nil {
		return nil, err
	}
	return post, nil
}


func (s *newsletterService) ListPostsByNewsletterID(ctx context.Context, newsletterID string, limit int, offset int) ([]models.Post, int, error) {
	// Optional: could add an ownership check here if only editor can list posts of their newsletter
	// For now, assuming it can be public or auth is handled by caller based on context.
	// First, verify newsletter exists.
	_, err := s.newsletterRepo.GetNewsletterByID(ctx, newsletterID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNewsletterNotFound) {
			return nil, 0, fmt.Errorf("service: ListPostsByNewsletterID: %w", apperrors.ErrNewsletterNotFound)
		}
		return nil, 0, fmt.Errorf("service: ListPostsByNewsletterID: checking newsletter: %w", err)
	}

	if limit <= 0 {
		limit = 10 // Default limit
	}
	if offset < 0 {
		offset = 0 // Default offset
	}

	posts, total, err := s.postRepo.ListPostsByNewsletterID(ctx, newsletterID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("service: ListPostsByNewsletterID: %w", err)
	}
	return posts, total, nil
}

func (s *newsletterService) UpdatePost(ctx context.Context, editorID string, postID string, title *string, content *string) (*models.Post, error) {
	editor, err := s.getEditorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	
	post, err := s.verifyPostOwnershipWithEditor(ctx, editor, postID)
	if err != nil {
		return nil, err
	}

	// Check if any changes are requested
	if title == nil && content == nil {
		return post, nil // No changes requested
	}

	// Validate title if provided
	if title != nil {
		trimmedTitle := strings.TrimSpace(*title)
		if trimmedTitle == "" {
			return nil, fmt.Errorf("service: UpdatePost: %w: title cannot be empty if provided", apperrors.ErrValidation)
		}
		if len(trimmedTitle) > MaxPostTitleLength {
			return nil, fmt.Errorf("service: UpdatePost: %w: title exceeds max length of %d", apperrors.ErrValidation, MaxPostTitleLength)
		}
		*title = trimmedTitle // Update the pointer value with trimmed version
	}

	// Validate content if provided
	if content != nil {
		trimmedContent := strings.TrimSpace(*content)
		if trimmedContent == "" {
			return nil, fmt.Errorf("service: UpdatePost: %w: content cannot be empty if provided", apperrors.ErrValidation)
		}
		if len(trimmedContent) < MinPostContentLength {
			return nil, fmt.Errorf("service: UpdatePost: %w: content must be at least %d characters", apperrors.ErrValidation, MinPostContentLength)
		}
		*content = trimmedContent // Update the pointer value with trimmed version
	}

	// Check if there are actual changes to avoid unnecessary updates
	if title != nil && post.Title == *title {
		title = nil // No change needed
	}
	if content != nil && post.Content == *content {
		content = nil // No change needed
	}

	// If no actual changes after validation, return current post
	if title == nil && content == nil {
		return post, nil
	}

	// Use the flexible repository method to update only the provided fields
	updates := repository.PostUpdate{
		Title:   title,
		Content: content,
	}
	
	updatedPost, err := s.postRepo.UpdatePost(ctx, postID, updates)

	if err != nil {
		if errors.Is(err, apperrors.ErrPostNotFound) {
			return nil, fmt.Errorf("service: UpdatePost: %w", apperrors.ErrPostNotFound)
		}
		return nil, fmt.Errorf("service: UpdatePost: failed to update post: %w", err)
	}
	return updatedPost, nil
}

func (s *newsletterService) DeletePost(ctx context.Context, editorID string, postID string) error {
	editor, err := s.getEditorFromContext(ctx)
	if err != nil {
		return err
	}
	
	_, err = s.verifyPostOwnershipWithEditor(ctx, editor, postID)
	if err != nil {
		return err
	}

	err = s.postRepo.DeletePost(ctx, postID)
	if err != nil {
		if errors.Is(err, apperrors.ErrPostNotFound) {
			return fmt.Errorf("service: DeletePost: %w", apperrors.ErrPostNotFound)
		}
		return fmt.Errorf("service: DeletePost: %w", err)
	}
	return nil
}

func (s *newsletterService) PublishPost(ctx context.Context, editorID string, postID string) (*models.Post, error) {
	editor, err := s.getEditorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	
	post, err := s.verifyPostOwnershipWithEditor(ctx, editor, postID)
	if err != nil {
		return nil, err
	}

	if post.IsPublished() {
		return post, nil // Already published, no action needed, return current state
	}

	now := time.Now().UTC()
	updatedPost, err := s.postRepo.SetPostPublished(ctx, postID, now)
	if err != nil {
		if errors.Is(err, apperrors.ErrPostNotFound) { // Should not happen
			return nil, fmt.Errorf("service: PublishPost: %w", apperrors.ErrPostNotFound)
		}
		return nil, fmt.Errorf("service: PublishPost: %w", err)
	}
	return updatedPost, nil
}

func (s *newsletterService) UnpublishPost(ctx context.Context, editorID string, postID string) (*models.Post, error) {
	editor, err := s.getEditorFromContext(ctx)
	if err != nil {
		return nil, err
	}
	
	post, err := s.verifyPostOwnershipWithEditor(ctx, editor, postID)
	if err != nil {
		return nil, err
	}

	if !post.IsPublished() {
		return post, nil // Already unpublished, no action needed
	}

	updatedPost, err := s.postRepo.SetPostUnpublished(ctx, postID)
	if err != nil {
		if errors.Is(err, apperrors.ErrPostNotFound) {
			return nil, fmt.Errorf("service: UnpublishPost: %w", apperrors.ErrPostNotFound)
		}
		return nil, fmt.Errorf("service: UnpublishPost: %w", err)
	}
	return updatedPost, nil
}
