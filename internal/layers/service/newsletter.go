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
	UpdatePost(ctx context.Context, editorID string, postID string, title *string, content *string, publishedAt *time.Time) (*models.Post, error)
	DeletePost(ctx context.Context, editorID string, postID string) error
	PublishPost(ctx context.Context, editorID string, postID string) (*models.Post, error)
	UnpublishPost(ctx context.Context, editorID string, postID string) (*models.Post, error)
}

type newsletterService struct {
	newsletterRepo repository.NewsletterRepository
	postRepo       repository.PostRepository
	editorRepo     repository.EditorRepository
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

// --- Authorization Helper ---

// verifyNewsletterOwnershipAndGetEditor retrieves the editor and newsletter,
// verifies ownership, and returns both if successful.
func (s *newsletterService) verifyNewsletterOwnershipAndGetEditor(ctx context.Context, editorAuthID string, newsletterID string) (*models.Editor, *models.Newsletter, error) {
	editor, err := s.editorRepo.GetEditorByFirebaseUID(ctx, editorAuthID) // Assuming editorAuthID is FirebaseUID
	if err != nil {
		if errors.Is(err, apperrors.ErrEditorNotFound) {
			return nil, nil, fmt.Errorf("service: verifyNewsletterOwnership: %w", apperrors.ErrForbidden) // Editor for token not found
		}
		return nil, nil, fmt.Errorf("service: verifyNewsletterOwnership: getting editor: %w", err) // Internal error
	}

	newsletter, err := s.newsletterRepo.GetNewsletterByID(ctx, newsletterID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNewsletterNotFound) {
			return nil, nil, fmt.Errorf("service: verifyNewsletterOwnership: %w", apperrors.ErrNewsletterNotFound)
		}
		return nil, nil, fmt.Errorf("service: verifyNewsletterOwnership: getting newsletter: %w", err) // Internal error
	}

	if newsletter.EditorID != editor.ID {
		return nil, nil, fmt.Errorf("service: verifyNewsletterOwnership: %w", apperrors.ErrForbidden)
	}
	return editor, newsletter, nil
}

// verifyPostOwnershipAndGetEditor retrieves the editor and post,
// then verifies that the editor owns the newsletter to which the post belongs.
// Returns the editor and post if successful.
func (s *newsletterService) verifyPostOwnershipAndGetEditor(ctx context.Context, editorAuthID string, postID string) (*models.Editor, *models.Post, error) {
	editor, err := s.editorRepo.GetEditorByFirebaseUID(ctx, editorAuthID) // Assuming editorAuthID is FirebaseUID
	if err != nil {
		if errors.Is(err, apperrors.ErrEditorNotFound) {
			return nil, nil, fmt.Errorf("service: verifyPostOwnership: %w", apperrors.ErrForbidden) // Editor for token not found
		}
		return nil, nil, fmt.Errorf("service: verifyPostOwnership: getting editor: %w", err)
	}

	post, err := s.postRepo.GetPostByID(ctx, postID)
	if err != nil {
		if errors.Is(err, apperrors.ErrPostNotFound) {
			return nil, nil, fmt.Errorf("service: verifyPostOwnership: %w", apperrors.ErrPostNotFound)
		}
		return nil, nil, fmt.Errorf("service: verifyPostOwnership: getting post: %w", err)
	}

	// Now verify ownership of the newsletter the post belongs to
	_, _, err = s.verifyNewsletterOwnershipAndGetEditor(ctx, editorAuthID, post.NewsletterID)
	if err != nil {
		// This will return ErrForbidden if newsletter not owned, or ErrNewsletterNotFound
		return nil, nil, fmt.Errorf("service: verifyPostOwnership: newsletter ownership check failed: %w", err)
	}

	return editor, post, nil
}

// --- Newsletter Methods ---

func (s *newsletterService) ListNewslettersByEditorID(ctx context.Context, editorID string, limit int, offset int) ([]models.Newsletter, int, error) {
	// Assuming editorID is the authenticated editor's DB ID, validated by the handler/middleware.
	// No further ownership check needed here for this specific method's logic.
	if limit <= 0 {
		limit = 10 // Default limit
	}
	if offset < 0 {
		offset = 0 // Default offset
	}
	newsletters, total, err := s.newsletterRepo.ListNewslettersByEditorID(ctx, editorID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("service: ListNewslettersByEditorID: %w", err) // Wrap internal errors
	}
	return newsletters, total, nil
}

func (s *newsletterService) CreateNewsletter(ctx context.Context, editorID, name, description string) (*models.Newsletter, error) {
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

	// Check for name uniqueness for this editor
	existing, err := s.newsletterRepo.GetNewsletterByNameAndEditorID(ctx, name, editorID)
	if err != nil && !errors.Is(err, apperrors.ErrNewsletterNotFound) { // If error is something other than "not found"
		return nil, fmt.Errorf("service: CreateNewsletter: checking name uniqueness: %w", err)
	}
	if existing != nil { // Found an existing newsletter with the same name for this editor
		return nil, fmt.Errorf("service: CreateNewsletter: %w: newsletter name '%s' is already taken", apperrors.ErrConflict, name)
	}
	
	// editorID is assumed to be the authenticated user's actual database ID.
	return s.newsletterRepo.CreateNewsletter(ctx, editorID, name, description)
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

func (s *newsletterService) GetNewsletterForEditor(ctx context.Context, editorAuthID string, newsletterID string) (*models.Newsletter, error) {
	// editorAuthID is the authenticated editor's unique identifier (e.g., Firebase UID)
	// newsletterID is the ID of the newsletter to fetch.
	_, newsletter, err := s.verifyNewsletterOwnershipAndGetEditor(ctx, editorAuthID, newsletterID)
	if err != nil {
		return nil, err // verifyNewsletterOwnershipAndGetEditor already wraps errors appropriately
	}
	return newsletter, nil
}

func (s *newsletterService) UpdateNewsletter(ctx context.Context, editorAuthID string, newsletterID string, name *string, description *string) (*models.Newsletter, error) {
	_, newsletter, err := s.verifyNewsletterOwnershipAndGetEditor(ctx, editorAuthID, newsletterID)
	if err != nil {
		return nil, err // Handles ErrForbidden, ErrNewsletterNotFound, or internal errors
	}

	// Keep original values if not provided for update
	updatedName := newsletter.Name
	if name != nil {
		trimmedName := strings.TrimSpace(*name)
		if trimmedName == "" {
			return nil, fmt.Errorf("service: UpdateNewsletter: %w: name cannot be empty if provided", apperrors.ErrValidation)
		}
		if len(trimmedName) > MaxNewsletterNameLength {
			return nil, fmt.Errorf("service: UpdateNewsletter: %w: name exceeds max length of %d", apperrors.ErrValidation, MaxNewsletterNameLength)
		}
		// Check for name uniqueness if it's being changed
		if trimmedName != newsletter.Name {
			existing, errDb := s.newsletterRepo.GetNewsletterByNameAndEditorID(ctx, trimmedName, newsletter.EditorID)
			if errDb != nil && !errors.Is(errDb, apperrors.ErrNewsletterNotFound) {
				return nil, fmt.Errorf("service: UpdateNewsletter: checking name uniqueness: %w", errDb)
			}
			if existing != nil { // A newsletter with this new name already exists for this editor
				return nil, fmt.Errorf("service: UpdateNewsletter: %w: newsletter name '%s' is already taken", apperrors.ErrConflict, trimmedName)
			}
		}
		updatedName = trimmedName
	}

	updatedDescription := newsletter.Description
	if description != nil {
		trimmedDescription := strings.TrimSpace(*description)
		if len(trimmedDescription) > MaxNewsletterDescriptionLength {
			return nil, fmt.Errorf("service: UpdateNewsletter: %w: description exceeds max length of %d", apperrors.ErrValidation, MaxNewsletterDescriptionLength)
		}
		updatedDescription = trimmedDescription
	}
	
	// Pass name and description pointers to repository.
	// The repository will handle actual update if values changed.
	// Here we pass potentially modified values.
	var namePtr *string
	if name != nil { // if user intended to update name
		namePtr = &updatedName
	}
	var descPtr *string
	if description != nil { // if user intended to update description
		descPtr = &updatedDescription
	}

	// The editorID for the UpdateNewsletter in repo is the true owner ID (newsletter.EditorID)
	return s.newsletterRepo.UpdateNewsletter(ctx, newsletterID, newsletter.EditorID, namePtr, descPtr)
}

func (s *newsletterService) DeleteNewsletter(ctx context.Context, editorAuthID string, newsletterID string) error {
	_, newsletter, err := s.verifyNewsletterOwnershipAndGetEditor(ctx, editorAuthID, newsletterID)
	if err != nil {
		return err // Handles ErrForbidden, ErrNewsletterNotFound, or internal errors
	}

	// The editorID for the DeleteNewsletter in repo is the true owner ID (newsletter.EditorID)
	err = s.newsletterRepo.DeleteNewsletter(ctx, newsletterID, newsletter.EditorID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNewsletterNotFound) { // Should not happen if verify found it, but as safeguard
			return fmt.Errorf("service: DeleteNewsletter: %w", apperrors.ErrNewsletterNotFound)
		}
		return fmt.Errorf("service: DeleteNewsletter: %w", err)
	}
	return nil
}

// --- Post Methods ---

func (s *newsletterService) CreatePost(ctx context.Context, editorAuthID string, newsletterID string, title string, content string) (*models.Post, error) {
	_, newsletter, err := s.verifyNewsletterOwnershipAndGetEditor(ctx, editorAuthID, newsletterID)
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

func (s *newsletterService) GetPostForEditor(ctx context.Context, editorAuthID string, postID string) (*models.Post, error) {
	// editorAuthID is the authenticated editor's unique identifier (e.g., Firebase UID)
	// postID is the ID of the post to fetch.
	_, post, err := s.verifyPostOwnershipAndGetEditor(ctx, editorAuthID, postID)
	if err != nil {
		return nil, err // verifyPostOwnershipAndGetEditor already wraps errors appropriately
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

func (s *newsletterService) UpdatePost(ctx context.Context, editorAuthID string, postID string, title *string, content *string, publishedAtInput *time.Time) (*models.Post, error) {
	_, post, err := s.verifyPostOwnershipAndGetEditor(ctx, editorAuthID, postID)
	if err != nil {
		return nil, err
	}

	updated := false
	if title != nil {
		trimmedTitle := strings.TrimSpace(*title)
		if trimmedTitle == "" {
			return nil, fmt.Errorf("service: UpdatePost: %w: title cannot be empty if provided", apperrors.ErrValidation)
		}
		if len(trimmedTitle) > MaxPostTitleLength {
			return nil, fmt.Errorf("service: UpdatePost: %w: title exceeds max length of %d", apperrors.ErrValidation, MaxPostTitleLength)
		}
		if post.Title != trimmedTitle {
			post.Title = trimmedTitle
			updated = true
		}
	}

	if content != nil {
		trimmedContent := strings.TrimSpace(*content)
		if trimmedContent == "" { // Consider minimum length
			return nil, fmt.Errorf("service: UpdatePost: %w: content cannot be empty if provided", apperrors.ErrValidation)
		}
		if len(trimmedContent) < MinPostContentLength {
			 return nil, fmt.Errorf("service: UpdatePost: %w: content must be at least %d characters", apperrors.ErrValidation, MinPostContentLength)
		}
		if post.Content != trimmedContent {
			post.Content = trimmedContent
			updated = true
		}
	}
	
	// PublishedAt is handled by PublishPost/UnpublishPost methods now for clarity
	// This method is for content/title updates. If publishedAtInput is provided, it implies an explicit set/unset
	// which is complex. Better to use dedicated methods.
	// For this generic Update, we will *not* modify PublishedAt.
	// If a service needs to update content AND publish status, it should call UpdatePost then PublishPost.


	if !updated && publishedAtInput == nil { // No actual changes to title/content, and not trying to change publishedAt
		return post, nil // Nothing to update
	}
    
    // If a change occurred, update timestamp
    if updated {
	    post.UpdatedAt = time.Now().UTC()
    }


	updatedPost, err := s.postRepo.UpdatePost(ctx, post) // postRepo.UpdatePost updates fields based on the passed post model
	if err != nil {
		if errors.Is(err, apperrors.ErrPostNotFound) { // Should not happen if verifyPostOwnership passed
			return nil, fmt.Errorf("service: UpdatePost: %w", apperrors.ErrPostNotFound)
		}
		return nil, fmt.Errorf("service: UpdatePost: failed to update post: %w", err)
	}
	return updatedPost, nil
}

func (s *newsletterService) DeletePost(ctx context.Context, editorAuthID string, postID string) error {
	_, _, err := s.verifyPostOwnershipAndGetEditor(ctx, editorAuthID, postID)
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

func (s *newsletterService) PublishPost(ctx context.Context, editorAuthID string, postID string) (*models.Post, error) {
	_, post, err := s.verifyPostOwnershipAndGetEditor(ctx, editorAuthID, postID)
	if err != nil {
		return nil, err
	}

	if post.IsPublished() {
		return post, nil // Already published, no action needed, return current state
	}

	now := time.Now().UTC()
	post.PublishedAt = &now
	post.UpdatedAt = now

	// Repository's MarkPostAsPublished updates PublishedAt and UpdatedAt
	// Let's use UpdatePost for consistency as MarkPostAsPublished might be too specific if repo changes
	updatedPost, err := s.postRepo.UpdatePost(ctx, post)
	if err != nil {
		if errors.Is(err, apperrors.ErrPostNotFound) { // Should not happen
			return nil, fmt.Errorf("service: PublishPost: %w", apperrors.ErrPostNotFound)
		}
		return nil, fmt.Errorf("service: PublishPost: %w", err)
	}
	return updatedPost, nil
}

func (s *newsletterService) UnpublishPost(ctx context.Context, editorAuthID string, postID string) (*models.Post, error) {
	_, post, err := s.verifyPostOwnershipAndGetEditor(ctx, editorAuthID, postID)
	if err != nil {
		return nil, err
	}

	if !post.IsPublished() {
		return post, nil // Already unpublished, no action needed
	}

	post.PublishedAt = nil // Set to nil to unpublish
	post.UpdatedAt = time.Now().UTC()

	updatedPost, err := s.postRepo.UpdatePost(ctx, post)
	if err != nil {
		if errors.Is(err, apperrors.ErrPostNotFound) { // Should not happen
			return nil, fmt.Errorf("service: UnpublishPost: %w", apperrors.ErrPostNotFound)
		}
		return nil, fmt.Errorf("service: UnpublishPost: %w", err)
	}
	return updatedPost, nil
}
