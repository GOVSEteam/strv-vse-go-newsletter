package service

import (
	"context"
	"errors" // For basic error creation
	"fmt"
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
	ConfirmSubscription(ctx context.Context, req ConfirmSubscriptionRequest) error
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
	SubscriberID string `json:"subscriber_id"`
	Email        string `json:"email"`
	NewsletterID string `json:"newsletter_id"`
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
		// If user exists and is pending confirmation, maybe resend confirmation? Or tell them to check email.
		// If active or unsubscribed, it's still ErrAlreadySubscribed for simplicity of this flow.
		// More nuanced logic can be added if an active user tries to subscribe again (e.g., inform them they are already active).
		return nil, ErrAlreadySubscribed
	}

	confirmationToken := uuid.NewString()
	tokenExpiryDuration := 24 * time.Hour // Example: 24 hours expiry

	subscriber := models.Subscriber{
		Email:             req.Email,
		NewsletterID:      req.NewsletterID,
		SubscriptionDate:  time.Now().UTC(),
		Status:            models.SubscriberStatusPendingConfirmation,
		ConfirmationToken: confirmationToken,
		TokenExpiryTime:   time.Now().UTC().Add(tokenExpiryDuration),
		UnsubscribeToken:  uuid.NewString(), // Generate unsubscribe token
	}

	subscriberID, err := s.subscriberRepo.CreateSubscriber(ctx, subscriber)
	if err != nil {
		return nil, fmt.Errorf("failed to create subscriber: %w", err)
	}

	// TODO: The confirmation link should point to a frontend URL that then calls a backend confirmation endpoint.
	// For now, placeholder. This would be configured.
	// Example: "https://yourfrontend.com/confirm-subscription?token=" + confirmationToken
	// For a backend-only test, it might be "http://localhost:8080/api/subscribers/confirm?token=" + confirmationToken
	confirmationLink := "http://localhost:8080/api/subscribers/confirm?token=" + confirmationToken // Placeholder

	// The recipient name for the email; using email if no other name is available.
	err = s.emailService.SendConfirmationEmail(subscriber.Email, subscriber.Email, confirmationLink)
	if err != nil {
		// Log the error but don't fail the subscription itself. The user is in pending state.
		// A background job could retry sending emails or allow manual resend.
		// For now, we just log it.
		// Consider how to handle this critical step failing in a production system.
		fmt.Printf("Error sending confirmation email to %s: %v\n", subscriber.Email, err)
		// Depending on requirements, you might want to return an error here or proceed.
		// For this example, we proceed, as the subscriber is created.
	}

	// TODO: Send an email containing the unsubscribe link with subscriber.UnsubscribeToken
	// For example: unsubscribeLink := "http://localhost:8080/api/subscriptions/unsubscribe?token=" + subscriber.UnsubscribeToken
	// This link should be part of all emails sent to the subscriber (confirmation, newsletter issues).

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

	// TODO: Optionally, clear the unsubscribe token after use or mark it as used if it's single-use.
	// For now, we just update the status.
	err = s.subscriberRepo.UpdateSubscriberStatus(ctx, subscriber.ID, models.SubscriberStatusUnsubscribed)
	if err != nil {
		return fmt.Errorf("failed to update subscriber status for unsubscription: %w", err)
	}

	// TODO: Optionally, send an unsubscription confirmation email.
	return nil
}


// ConfirmSubscriptionRequest defines the input for confirming a subscription.
type ConfirmSubscriptionRequest struct {
	Token string
}

// ConfirmSubscription validates a confirmation token and activates the subscriber.
func (s *SubscriberService) ConfirmSubscription(ctx context.Context, req ConfirmSubscriptionRequest) error {
	if req.Token == "" {
		return errors.New("confirmation token cannot be empty")
	}

	subscriber, err := s.subscriberRepo.GetSubscriberByConfirmationToken(ctx, req.Token)
	if err != nil {
		// This is a server-side error from the repository
		return fmt.Errorf("error retrieving subscriber by token: %w", err)
	}
	if subscriber == nil {
		return ErrInvalidOrExpiredToken // Token not found
	}

	if subscriber.Status == models.SubscriberStatusActive && subscriber.ConfirmedAt != nil {
		return ErrAlreadyConfirmed
	}

	// Double check status, though GetByConfirmationToken should ideally only return pending ones.
	if subscriber.Status != models.SubscriberStatusPendingConfirmation {
		return ErrInvalidOrExpiredToken // Or a more specific error like "subscription not in pending state"
	}

	if time.Now().UTC().After(subscriber.TokenExpiryTime) {
		// TODO: Optionally, allow resending a new confirmation email or deleting the expired pending subscription.
		return ErrInvalidOrExpiredToken // Token expired
	}

	confirmationTime := time.Now().UTC()
	err = s.subscriberRepo.ConfirmSubscriber(ctx, subscriber.ID, confirmationTime)
	if err != nil {
		return fmt.Errorf("failed to confirm subscriber: %w", err)
	}

	// TODO: Optionally, send a "Welcome" or "Subscription Confirmed" email.
	// This email should also contain the unsubscribe link.

	return nil
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
