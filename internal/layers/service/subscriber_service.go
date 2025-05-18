package service

import (
	"context"
	"errors" // For basic error creation
	"fmt"
	"time"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/repository"
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

// SubscriberService handles business logic for subscriber management.
// Placeholder for now.
type SubscriberService struct {
	subscriberRepo repository.SubscriberRepository
	newsletterRepo repository.NewsletterRepository // Added newsletter repository dependency
	emailService   email.EmailService // Added EmailService dependency
	// We might add other dependencies like an EmailService later
}

// NewSubscriberService creates a new SubscriberService.
// Placeholder for now.
func NewSubscriberService(subRepo repository.SubscriberRepository, newsRepo repository.NewsletterRepository, emailSvc email.EmailService) *SubscriberService {
	return &SubscriberService{
		subscriberRepo: subRepo,
		newsletterRepo: newsRepo, // Store dependency
		emailService:   emailSvc, // Store EmailService dependency
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

// UnsubscribeFromNewsletter processes an unsubscription request.
// It updates the subscriber's status to Unsubscribed.
func (s *SubscriberService) UnsubscribeFromNewsletter(ctx context.Context, req UnsubscribeFromNewsletterRequest) error {
	if req.Email == "" {
		return errors.New("email cannot be empty")
	}
	if req.NewsletterID == "" {
		return errors.New("newsletter ID cannot be empty")
	}

	// Check if the newsletter itself exists. While not strictly necessary for unsubscription
	// (one might want to unsubscribe even if a newsletter was deleted),
	// it can prevent attempts to unsubscribe from non-existent entities if that's desired behavior.
	// For now, we'll skip this to allow unsubscription even if a newsletter is deleted.
	// newsletter, err := s.newsletterRepo.GetNewsletterByID(req.NewsletterID)
	// if err != nil {
	// 	return err // DB error
	// }
	// if newsletter == nil {
	// 	return ErrNewsletterNotFound // or a different error like "cannot unsubscribe from non-existent newsletter"
	// }

	// Find the subscriber record to get its ID
	existingSub, err := s.subscriberRepo.GetSubscriberByEmailAndNewsletterID(ctx, req.Email, req.NewsletterID)
	if err != nil {
		// This is a server-side error (e.g., DB connection)
		return err
	}
	if existingSub == nil {
		return ErrSubscriptionNotFound
	}

	// If already unsubscribed, we can consider it a success or a specific no-op response.
	// For simplicity, we'll proceed with the update, which is idempotent for status.
	if existingSub.Status == models.SubscriberStatusUnsubscribed {
		return nil // Already unsubscribed, no action needed
	}

	err = s.subscriberRepo.UpdateSubscriberStatus(ctx, existingSub.ID, models.SubscriberStatusUnsubscribed)
	if err != nil {
		// Handle potential errors from the update operation, e.g., the subscriber was deleted
		// between the Get and Update calls, though UpdateSubscriberStatus already checks for NotFound.
		return err // Propagate repository error
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

	return nil
} 