package service

import (
	"context"
	"errors" // For basic error creation
	"fmt"
	"os"
	"time"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/pkg/email" // Added email package
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

type SubscriberServiceInterface interface {
	SubscribeToNewsletter(ctx context.Context, req SubscribeToNewsletterRequest) (*SubscribeToNewsletterResponse, error)
	UnsubscribeFromNewsletter(ctx context.Context, req UnsubscribeFromNewsletterRequest) error // Will likely be deprecated
	UnsubscribeByToken(ctx context.Context, token string) error
	GetActiveSubscribersForNewsletter(ctx context.Context, newsletterID string) ([]models.Subscriber, error) // Added for SUB-003
}

// SubscriberService handles business logic for subscriber management.
type SubscriberService struct {
	subscriberRepo repository.SubscriberRepository
	newsletterRepo repository.NewsletterRepository
	emailService   email.EmailService
}

// NewSubscriberService creates a new SubscriberService.
func NewSubscriberService(subRepo repository.SubscriberRepository, newsRepo repository.NewsletterRepository, emailSvc email.EmailService) SubscriberServiceInterface { // Return interface
	return &SubscriberService{
		subscriberRepo: subRepo,
		newsletterRepo: newsRepo,
		emailService:   emailSvc,
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
func (s *SubscriberService) SubscribeToNewsletter(ctx context.Context, req SubscribeToNewsletterRequest) (*SubscribeToNewsletterResponse, error) {
	if req.Email == "" {
		return nil, errors.New("email cannot be empty")
	}
	if req.NewsletterID == "" {
		return nil, errors.New("newsletter ID cannot be empty")
	}

	newsletter, err := s.newsletterRepo.GetNewsletterByID(req.NewsletterID)
	if err != nil {
		return nil, err
	}
	if newsletter == nil {
		return nil, ErrNewsletterNotFound
	}

	existingSub, err := s.subscriberRepo.GetSubscriberByEmailAndNewsletterID(ctx, req.Email, req.NewsletterID)
	if err != nil {
		return nil, err
	}
	if existingSub != nil {
		switch existingSub.Status {
		case models.SubscriberStatusUnsubscribed:
			// User previously unsubscribed: reactivate and resend confirmation
			existingSub.Status = models.SubscriberStatusActive
			existingSub.SubscriptionDate = time.Now().UTC()
			existingSub.UnsubscribeToken = uuid.NewString()

			err = s.subscriberRepo.UpdateSubscriberStatus(ctx, existingSub.ID, models.SubscriberStatusActive)
			if err != nil {
				return nil, fmt.Errorf("failed to reactivate subscriber: %w", err)
			}

			err = s.subscriberRepo.UpdateSubscriberUnsubscribeToken(ctx, existingSub.ID, existingSub.UnsubscribeToken)
			if err != nil {
				return nil, fmt.Errorf("failed to save new token to subscriber: %w", err)
			}

			// (Re-)build the unsubscribe link
			appBaseURL := os.Getenv("APP_BASE_URL")
			if appBaseURL == "" {
				appBaseURL = "http://localhost:8080"
			}
			unsubscribeLink := fmt.Sprintf("%s/api/subscriptions/unsubscribe?token=%s", appBaseURL, existingSub.UnsubscribeToken)

			// Resend confirmation/unsubscribe email
			if err := s.emailService.SendConfirmationEmail(
				existingSub.Email,
				existingSub.Email,
				unsubscribeLink,
			); err != nil {
				// log but don’t fail the overall flow
				fmt.Printf("Error resending confirmation email to %s: %v\n", existingSub.Email, err)
			}

			// Return the reactivated response
			return &SubscribeToNewsletterResponse{
				SubscriberID: existingSub.ID,
				Email:        existingSub.Email,
				NewsletterID: existingSub.NewsletterID,
				Status:       existingSub.Status,
			}, nil

		default:
			// already active status
			return nil, ErrAlreadySubscribed
		}
	}

	unsubscribeToken := uuid.NewString() // Generate unsubscribe token

	subscriber := models.Subscriber{
		Email:            req.Email,
		NewsletterID:     req.NewsletterID,
		SubscriptionDate: time.Now().UTC(),
		Status:           models.SubscriberStatusActive,
		UnsubscribeToken: unsubscribeToken,
	}

	subscriberID, err := s.subscriberRepo.CreateSubscriber(ctx, subscriber)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscriber: %w", err)
	}

	appBaseURL := os.Getenv("APP_BASE_URL") // e.g., http://localhost:8080
	if appBaseURL == "" {
		appBaseURL = "http://localhost:8080" // Default if not set
		fmt.Println("Warning: APP_BASE_URL not set, defaulting to http://localhost:8080 for unsubscribe links.")
	}

	unsubscribeLink := fmt.Sprintf("%s/api/subscriptions/unsubscribe?token=%s", appBaseURL, subscriber.UnsubscribeToken)

	// The recipient name for the email; using email if no other name is available.
	err = s.emailService.SendConfirmationEmail(subscriber.Email, subscriber.Email, unsubscribeLink)
	if err != nil {
		// For now, we just log it.
		// Consider how to handle this critical step failing in a production system.
		fmt.Printf("Error sending confirmation email to %s: %v\n", subscriber.Email, err)
		// Depending on requirements, you might want to return an error here or proceed.
		// For this example, we proceed, as the subscriber is created.
	}

	return &SubscribeToNewsletterResponse{
		SubscriberID: subscriberID,
		Email:        subscriber.Email,
		NewsletterID: subscriber.NewsletterID,
		Status:       subscriber.Status, // Will be pending_confirmation
	}, nil
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
	if token == "" {
		return errors.New("unsubscribe token cannot be empty")
	}

	subscriber, err := s.subscriberRepo.GetSubscriberByUnsubscribeToken(ctx, token)
	if err != nil {
		// This is a server-side error from the repository
		return fmt.Errorf("error retrieving subscriber by unsubscribe token: %w", err)
	}
	if subscriber == nil {
		return ErrInvalidOrExpiredToken // Using this error for simplicity, could be a more specific "unsubscribe token not found"
	}

	if subscriber.Status == models.SubscriberStatusUnsubscribed {
		return nil // Already unsubscribed
	}

	err = s.subscriberRepo.UpdateSubscriberUnsubscribeToken(ctx, subscriber.ID, "")
	if err != nil {
		return fmt.Errorf("failed to remove unsubscribe token: %w", err)
	}

	err = s.subscriberRepo.UpdateSubscriberStatus(ctx, subscriber.ID, models.SubscriberStatusUnsubscribed)
	if err != nil {
		return fmt.Errorf("failed to update subscriber status for unsubscription: %w", err)
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
