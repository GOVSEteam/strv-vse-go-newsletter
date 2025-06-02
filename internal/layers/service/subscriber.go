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
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models" // Added email package
	"github.com/google/uuid"
)

// ErrAlreadySubscribed is returned when a user tries to subscribe to a newsletter they are already subscribed to.
var ErrAlreadySubscribed = errors.New("email is already subscribed to this newsletter")

// ErrNewsletterNotFound is returned when a subscription attempt is made for a non-existent newsletter.
var ErrNewsletterNotFound = errors.New("newsletter not found")

// ErrSubscriptionNotFound is returned when an attempt to modify a subscription fails because it doesn't exist.
var ErrSubscriptionNotFound = errors.New("subscription not found for the given email and newsletter ID")

// ErrInvalidOrExpiredToken is returned when a confirmation token is invalid, not found, or expired.
var ErrInvalidOrExpiredToken = errors.New("confirmation token is invalid or expired")

// ErrAlreadyConfirmed is returned when a subscription is already confirmed.
var ErrAlreadyConfirmed = errors.New("subscription is already confirmed")

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
}

// SubscriberService handles business logic for subscriber management.
type SubscriberService struct {
	subscriberRepo repository.SubscriberRepository
	newsletterRepo repository.NewsletterRepository
	editorRepo     repository.EditorRepository // For authorization
	emailService   EmailService              // Corrected: Uses EmailService interface from this package
	appBaseURL     string                    // For generating unsubscribe links, e.g., "http://localhost:8080"
}

// NewSubscriberService creates a new SubscriberService.
func NewSubscriberService(
	subRepo repository.SubscriberRepository,
	newsRepo repository.NewsletterRepository,
	editorRepo repository.EditorRepository,
	emailSvc EmailService, // Corrected: Expecting EmailService interface from this package
	appBaseURL string,
) SubscriberServiceInterface {
	return &SubscriberService{
		subscriberRepo: subRepo,
		newsletterRepo: newsRepo,
		editorRepo:     editorRepo,
		emailService:   emailSvc,
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

			unsubscribeLink := fmt.Sprintf("%s/api/v1/subscriptions/unsubscribe/%s", s.appBaseURL, unsubscribeToken)
			bodyMsg := fmt.Sprintf("Welcome back! You have been re-subscribed to '%s'. If you wish to unsubscribe, please use this link: %s", newsletter.Name, unsubscribeLink)
			go func(emailAddr, subject, bodyContent string) {
				bgCtx := context.Background() // Use a background context for the goroutine
				if emailErr := s.emailService.SendEmail(bgCtx, emailAddr, subject, bodyContent); emailErr != nil {
					fmt.Printf("WARN: service: SubscribeToNewsletter: failed to send welcome back email to %s: %v\n", emailAddr, emailErr)
				}
			}(existingSub.Email, WelcomeBackEmailSubject, bodyMsg)

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

	unsubscribeLink := fmt.Sprintf("%s/api/v1/subscriptions/unsubscribe/%s", s.appBaseURL, unsubscribeToken)
	bodyMsg := fmt.Sprintf("Thank you for subscribing to '%s'! If you wish to unsubscribe, please use this link: %s", newsletter.Name, unsubscribeLink)
	go func(emailAddr, subject, bodyContent string) {
		bgCtx := context.Background() // Use a background context for the goroutine
		if emailErr := s.emailService.SendEmail(bgCtx, emailAddr, subject, bodyContent); emailErr != nil {
			fmt.Printf("WARN: service: SubscribeToNewsletter: failed to send confirmation email to %s: %v\n", emailAddr, emailErr)
		}
	}(subscriber.Email, UnsubscribeEmailSubject, bodyMsg)

	return &subscriber, nil
}

// UnsubscribeFromNewsletterRequest defines the input for unsubscribing from a newsletter.
type UnsubscribeFromNewsletterRequest struct {
	Email        string
	NewsletterID string
}

// UnsubscribeFromNewsletter processes an unsubscription request. (DEPRECATED in favor of UnsubscribeByToken)
// It updates the subscriber's status to Unsubscribed.
func (s *SubscriberService) UnsubscribeFromNewsletter(ctx context.Context, req UnsubscribeFromNewsletterRequest) error {
	// This method is kept for now but should be replaced by token-based unsubscription.
	// It's functionality is effectively replaced by UnsubscribeByToken.
	// Consider removing it in a future refactor if no longer called.
	if req.Email == "" {
		return errors.New("email cannot be empty (deprecated method)")
	}
	if req.NewsletterID == "" {
		return errors.New("newsletter ID cannot be empty (deprecated method)")
	}

	existingSub, err := s.subscriberRepo.GetSubscriberByEmailAndNewsletterID(ctx, req.Email, req.NewsletterID)
	if err != nil {
		return err
	}
	if existingSub == nil {
		return ErrSubscriptionNotFound
	}
	if existingSub.Status == models.SubscriberStatusUnsubscribed {
		return nil // Already unsubscribed
	}
	return s.subscriberRepo.UpdateSubscriberStatus(ctx, existingSub.ID, models.SubscriberStatusUnsubscribed)
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
	newsletter, err := s.newsletterRepo.GetNewsletterByID(newsletterID)
	if err != nil {
		// This could be a DB error or other issue fetching the newsletter.
		return nil, fmt.Errorf("error checking newsletter existence: %w", err)
	}
	if newsletter == nil {
		return nil, ErrNewsletterNotFound // Or a more specific error.
	}

	subscribers, err := s.subscriberRepo.GetActiveSubscribersByNewsletterID(ctx, newsletterID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active subscribers: %w", err)
	}
	return subscribers, nil
}

func (s *SubscriberService) ListActiveSubscribersByNewsletterID(ctx context.Context, editorAuthID string, newsletterID string, limit, offset int) ([]models.Subscriber, int, error) {
	newsletterID = strings.TrimSpace(newsletterID)
	editorAuthID = strings.TrimSpace(editorAuthID)

	if newsletterID == "" {
		return nil, 0, fmt.Errorf("service: ListActiveSubscribers: %w: newsletterID cannot be empty", apperrors.ErrValidation)
	}
	if editorAuthID == "" {
		return nil, 0, fmt.Errorf("service: ListActiveSubscribers: %w: editorAuthID cannot be empty for authorization", apperrors.ErrValidation)
	}

	editor, err := s.editorRepo.GetEditorByFirebaseUID(ctx, editorAuthID)
	if err != nil {
		if errors.Is(err, apperrors.ErrEditorNotFound) {
			return nil, 0, fmt.Errorf("service: ListActiveSubscribers: %w", apperrors.ErrForbidden)
		}
		return nil, 0, fmt.Errorf("service: ListActiveSubscribers: getting editor: %w", err)
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

	allSubscribersInPage, totalCountForAllStatuses, err := s.subscriberRepo.ListSubscribersByNewsletterID(ctx, newsletterID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("service: ListActiveSubscribers: listing subscribers from repo: %w", err)
	}

	activeSubscribers := make([]models.Subscriber, 0, len(allSubscribersInPage))
	for _, sub := range allSubscribersInPage {
		if sub.Status == models.SubscriberStatusActive {
			activeSubscribers = append(activeSubscribers, sub)
		}
	}
	return activeSubscribers, totalCountForAllStatuses, nil
}
