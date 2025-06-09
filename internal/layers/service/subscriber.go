package service

import (
	"context"
	"errors" // For basic error creation
	"fmt"
	"net/mail"
	"regexp"
	"strings"
	"time"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/middleware"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"github.com/google/uuid"
)

// Note: All error definitions have been moved to internal/errors package for centralization.
// Use apperrors.ErrAlreadySubscribed, apperrors.ErrSubscriptionNotFound, etc. instead.

// EmailRegex is reused from editor_service, consider moving to a common util if used in more places.
var subscriberEmailRegex = regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,4}$`)

const (
	DefaultSubscriptionListPageLimit = 10
	UnsubscribeEmailSubject          = "Subscription Confirmed & Unsubscribe Link"
	WelcomeBackEmailSubject          = "Welcome Back & Unsubscribe Link"
)

// SubscriberServiceInterface defines the operations for subscriber management.
// Note: The EmailServiceInterface dependency is implicitly expected by NewSubscriberService.
type SubscriberServiceInterface interface {
	SubscribeToNewsletter(ctx context.Context, email, newsletterID string) (*models.Subscriber, error)
	UnsubscribeByToken(ctx context.Context, token string) error
	ListActiveSubscribersByNewsletterID(ctx context.Context, editorAuthID string, newsletterID string, limit, offset int) ([]models.Subscriber, int, error)
	GetActiveSubscribersForNewsletter(ctx context.Context, newsletterID string) ([]models.Subscriber, error)
}

// SubscriberService manages subscriber operations for newsletters.
type SubscriberService struct {
	subscriberRepo repository.SubscriberRepository
	newsletterRepo repository.NewsletterRepository
	editorRepo     repository.EditorRepository // For authorization
	emailService   EmailService                 // Use direct email service instead of email worker
	appBaseURL     string                      // For generating unsubscribe links, e.g., "http://localhost:8080"
}

// NewSubscriberService creates a new SubscriberService.
func NewSubscriberService(
	subRepo repository.SubscriberRepository,
	newsRepo repository.NewsletterRepository,
	editorRepo repository.EditorRepository,
	emailService EmailService, // Use direct email service instead of email worker
	appBaseURL string,
) SubscriberServiceInterface {
	return &SubscriberService{
		subscriberRepo: subRepo,
		newsletterRepo: newsRepo,
		editorRepo:     editorRepo,
		emailService:   emailService,
		appBaseURL:     appBaseURL,
	}
}

// SubscribeToNewsletterRequest defines the input for subscribing to a newsletter.
type SubscribeToNewsletterRequest struct {
	Email        string `json:"email"`
	NewsletterID string // Usually from path parameter, not JSON body for this field
}

// SubscribeToNewsletterResponse defines the output after a successful subscription.
type SubscribeToNewsletterResponse struct {
	SubscriberID string                  `json:"subscriber_id"`
	Email        string                  `json:"email"`
	NewsletterID string                  `json:"newsletter_id"`
	Status       models.SubscriberStatus `json:"status"`
}

// SubscribeToNewsletter processes a subscription request.
// It creates a new subscriber record with a pending_confirmation status and triggers a confirmation email.
func (s *SubscriberService) SubscribeToNewsletter(ctx context.Context, email, newsletterID string) (*models.Subscriber, error) {
	email = strings.TrimSpace(strings.ToLower(email))
	newsletterID = strings.TrimSpace(newsletterID)

	if email == "" {
		return nil, fmt.Errorf("service: SubscribeToNewsletter: %w: email cannot be empty", apperrors.ErrValidation)
	}
	if _, err := mail.ParseAddress(email); err != nil || !subscriberEmailRegex.MatchString(email) {
		return nil, fmt.Errorf("service: SubscribeToNewsletter: %w '%s'", apperrors.ErrInvalidEmail, email)
	}
	if newsletterID == "" {
		return nil, fmt.Errorf("service: SubscribeToNewsletter: %w: newsletterID cannot be empty", apperrors.ErrValidation)
	}

	newsletter, err := s.newsletterRepo.GetNewsletterByID(ctx, newsletterID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNewsletterNotFound) {
			return nil, fmt.Errorf("service: SubscribeToNewsletter: newsletter '%s' %w", newsletterID, apperrors.ErrNotFound)
		}
		return nil, fmt.Errorf("service: SubscribeToNewsletter: checking newsletter: %w", err)
	}

	existingSub, err := s.subscriberRepo.GetSubscriberByEmailAndNewsletterID(ctx, email, newsletterID)
	if err != nil && !errors.Is(err, apperrors.ErrSubscriberNotFound) {
		return nil, fmt.Errorf("service: SubscribeToNewsletter: checking existing subscription: %w", err)
	}

	unsubscribeToken := uuid.NewString()
	now := time.Now().UTC()

	if existingSub != nil {
		if existingSub.Status == models.SubscriberStatusActive {
			return nil, fmt.Errorf("service: SubscribeToNewsletter: %w: email '%s' is already actively subscribed to newsletter '%s'", apperrors.ErrConflict, email, newsletter.Name)
		}
		if existingSub.Status == models.SubscriberStatusUnsubscribed {
			existingSub.Status = models.SubscriberStatusActive
			existingSub.SubscriptionDate = now
			existingSub.UnsubscribeToken = unsubscribeToken

			if err := s.subscriberRepo.UpdateSubscriberStatus(ctx, existingSub.ID, models.SubscriberStatusActive); err != nil {
				return nil, fmt.Errorf("service: SubscribeToNewsletter: reactivating subscriber status: %w", err)
			}
			if err := s.subscriberRepo.UpdateSubscriberUnsubscribeToken(ctx, existingSub.ID, unsubscribeToken); err != nil {
				return nil, fmt.Errorf("service: SubscribeToNewsletter: updating token for reactivated subscriber: %w", err)
			}

			// Generate unsubscribe link and extract recipient name
			unsubscribeLink := fmt.Sprintf("%s/api/subscriptions/unsubscribe?token=%s", s.appBaseURL, unsubscribeToken)
			recipientName := email
			if atIndex := strings.Index(email, "@"); atIndex > 0 {
				recipientName = email[:atIndex]
			}

			// Send confirmation email directly
			err := s.emailService.SendConfirmationEmailHTML(ctx, existingSub.Email, recipientName, unsubscribeLink)
			if err != nil {
				fmt.Printf("Warning: Failed to send confirmation email to subscriber %s: %v\n", existingSub.Email, err)
			}

			// Return a model representing the updated state.
			// The existingSub was modified in place and these changes were persisted.
			return existingSub, nil
		}
		return nil, fmt.Errorf("service: SubscribeToNewsletter: %w: email '%s' has an unexpected status for newsletter '%s'", apperrors.ErrConflict, email, newsletter.Name)
	}

	subscriber := models.Subscriber{
		Email:            email,
		NewsletterID:     newsletterID,
		SubscriptionDate: now,
		Status:           models.SubscriberStatusActive,
		UnsubscribeToken: unsubscribeToken,
	}

	subscriberIDVal, err := s.subscriberRepo.CreateSubscriber(ctx, subscriber)
	if err != nil {
		return nil, fmt.Errorf("service: SubscribeToNewsletter: creating subscriber: %w", err)
	}
	subscriber.ID = subscriberIDVal

	// Generate unsubscribe link and extract recipient name
	unsubscribeLink := fmt.Sprintf("%s/api/subscriptions/unsubscribe?token=%s", s.appBaseURL, unsubscribeToken)
	recipientName := email
	if atIndex := strings.Index(email, "@"); atIndex > 0 {
		recipientName = email[:atIndex]
	}

	// Send confirmation email directly
	err = s.emailService.SendConfirmationEmailHTML(ctx, subscriber.Email, recipientName, unsubscribeLink)
	if err != nil {
		// Critical: If we can't send the confirmation email, we should fail the subscription
		// The subscriber was already created in the database, so we need to clean up
		if updateErr := s.subscriberRepo.UpdateSubscriberStatus(ctx, subscriber.ID, models.SubscriberStatusUnsubscribed); updateErr != nil {
			fmt.Printf("ERROR: Failed to mark subscriber %s as unsubscribed after email sending failure: %v\n", subscriber.ID, updateErr)
		}
		return nil, fmt.Errorf("failed to send confirmation email to %s: %w", email, err)
	}

	return &subscriber, nil
}

// UnsubscribeByToken processes an unsubscription request using a token.
func (s *SubscriberService) UnsubscribeByToken(ctx context.Context, token string) error {
	token = strings.TrimSpace(token)
	if token == "" {
		return fmt.Errorf("service: UnsubscribeByToken: %w: unsubscribe token cannot be empty", apperrors.ErrValidation)
	}

	subscriber, err := s.subscriberRepo.GetSubscriberByUnsubscribeToken(ctx, token)
	if err != nil {
		if errors.Is(err, apperrors.ErrSubscriberNotFound) {
			return fmt.Errorf("service: UnsubscribeByToken: %w: invalid or expired token", apperrors.ErrTokenInvalid)
		}
		return fmt.Errorf("service: UnsubscribeByToken: retrieving subscriber by token: %w", err)
	}

	if subscriber.Status == models.SubscriberStatusUnsubscribed {
		return nil // Already unsubscribed
	}

	if err := s.subscriberRepo.UpdateSubscriberUnsubscribeToken(ctx, subscriber.ID, ""); err != nil {
		// Log original error for server visibility, return a generic one to user if needed for security.
		fmt.Printf("ERROR: service: UnsubscribeByToken: failed to invalidate token for subscriber %s: %v\n", subscriber.ID, err)
		return fmt.Errorf("service: UnsubscribeByToken: failed to update token state: %w", apperrors.ErrInternal) 
	}

	if err := s.subscriberRepo.UpdateSubscriberStatus(ctx, subscriber.ID, models.SubscriberStatusUnsubscribed); err != nil {
		fmt.Printf("ERROR: service: UnsubscribeByToken: failed to update status for subscriber %s: %v\n", subscriber.ID, err)
		return fmt.Errorf("service: UnsubscribeByToken: failed to update subscription status: %w", apperrors.ErrInternal)
	}

	return nil
}

// ConfirmSubscriptionRequest defines the input for confirming a subscription.
type ConfirmSubscriptionRequest struct {
	Token string
}

// GetActiveSubscribersForNewsletter retrieves all active subscribers for a given newsletter.
func (s *SubscriberService) GetActiveSubscribersForNewsletter(ctx context.Context, newsletterID string) ([]models.Subscriber, error) {
	if newsletterID == "" {
		return nil, errors.New("newsletter ID cannot be empty")
	}

	// First, check if the newsletter exists to avoid querying for subscribers of a non-existent newsletter.
	// This is optional but good practice.
	newsletter, err := s.newsletterRepo.GetNewsletterByID(ctx, newsletterID)
	if err != nil {
		// This could be a DB error or other issue fetching the newsletter.
		return nil, fmt.Errorf("error checking newsletter existence: %w", err)
	}
	if newsletter == nil {
		return nil, apperrors.ErrNewsletterNotFound // Or a more specific error.
	}

	// Get all active subscribers directly from repository (no in-memory filtering)
	activeSubscribers, err := s.subscriberRepo.GetAllActiveSubscribersByNewsletterID(ctx, newsletterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active subscribers: %w", err)
	}
	
	return activeSubscribers, nil
}

// getEditorFromContext retrieves the authenticated editor from context.
// This eliminates the need for additional database queries and ensures consistency with newsletter service.
func (s *SubscriberService) getEditorFromContext(ctx context.Context) (*models.Editor, error) {
	if editor, ok := ctx.Value(middleware.EditorContextKey).(*models.Editor); ok {
		return editor, nil
	}
	return nil, fmt.Errorf("service: getEditorFromContext: %w", apperrors.ErrForbidden)
}

func (s *SubscriberService) ListActiveSubscribersByNewsletterID(ctx context.Context, editorAuthID string, newsletterID string, limit, offset int) ([]models.Subscriber, int, error) {
	newsletterID = strings.TrimSpace(newsletterID)

	if newsletterID == "" {
		return nil, 0, fmt.Errorf("service: ListActiveSubscribers: %w: newsletterID cannot be empty", apperrors.ErrValidation)
	}

	// Use context-based authorization like newsletter service for consistency
	editor, err := s.getEditorFromContext(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("service: ListActiveSubscribers: authorization failed: %w", err)
	}

	retrievedNewsletter, err := s.newsletterRepo.GetNewsletterByID(ctx, newsletterID)
	if err != nil {
		if errors.Is(err, apperrors.ErrNewsletterNotFound) {
			return nil, 0, fmt.Errorf("service: ListActiveSubscribers: newsletter '%s' %w", newsletterID, apperrors.ErrNotFound)
		}
		return nil, 0, fmt.Errorf("service: ListActiveSubscribers: getting newsletter: %w", err)
	}
	if retrievedNewsletter.EditorID != editor.ID {
		return nil, 0, fmt.Errorf("service: ListActiveSubscribers: %w: editor does not own newsletter '%s'", apperrors.ErrForbidden, newsletterID)
	}

	if limit <= 0 {
		limit = DefaultSubscriptionListPageLimit
	}
	if offset < 0 {
		offset = 0
	}

	activeSubscribers, totalActiveCount, err := s.subscriberRepo.ListActiveSubscribersByNewsletterID(ctx, newsletterID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("service: ListActiveSubscribers: listing active subscribers from repo: %w", err)
	}

	return activeSubscribers, totalActiveCount, nil
}
