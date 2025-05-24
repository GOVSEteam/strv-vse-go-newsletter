package service

import (
	"context"
	"fmt"
	"os" // For App Base URL (can be improved with config struct)

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models" // Ensure models is imported
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/pkg/email"
	"github.com/google/uuid" // For UUID parsing
)

var _ models.Post // Trick to ensure models import is kept if auto-formatter is aggressive

// PublishingServiceInterface defines the contract for the publishing service.
type PublishingServiceInterface interface {
	PublishPostToSubscribers(ctx context.Context, postID string, editorFirebaseUID string) error
}

// PublishingService handles the logic for publishing posts to subscribers.
type PublishingService struct {
	newsletterService NewsletterServiceInterface // To get post details & mark as published
	subscriberService SubscriberServiceInterface // To get active subscribers
	emailService      email.EmailService         // To send emails
	// We might need PostRepository directly if NewsletterService doesn't expose enough post details
	// or if we want to decouple post fetching from NewsletterService for this specific flow.
	// For now, assuming NewsletterService can provide necessary post details (like its content and newsletterID).
}

// NewPublishingService creates a new PublishingService.
func NewPublishingService(
	newsletterService NewsletterServiceInterface,
	subscriberService SubscriberServiceInterface,
	emailService email.EmailService,
) PublishingServiceInterface {
	return &PublishingService{
		newsletterService: newsletterService,
		subscriberService: subscriberService,
		emailService:      emailService,
	}
}

// PublishPostToSubscribers orchestrates the process of sending a post to all active subscribers of its newsletter.
func (s *PublishingService) PublishPostToSubscribers(ctx context.Context, postID string, editorFirebaseUID string) error {
	// 1. Get Post details and verify ownership via NewsletterService
	//    This method should also return the newsletter_id associated with the post.
	//    Let's assume GetPostForPublishing also checks if the post is already published.
	postUUID, err := uuid.Parse(postID)
	if err != nil {
		return fmt.Errorf("invalid post ID format: %w", err)
	}

	post, err := s.newsletterService.GetPostForPublishing(ctx, postUUID, editorFirebaseUID)
	if err != nil {
		return fmt.Errorf("failed to get post for publishing: %w", err) // Handles not found, not owner, already published etc.
	}
	if post == nil { // Should be caught by error above, but as a safeguard
		return fmt.Errorf("post %s not found or not accessible for publishing", postID)
	}

	// 2. Get active subscribers for the newsletter
	activeSubscribers, err := s.subscriberService.GetActiveSubscribersForNewsletter(ctx, post.NewsletterID.String())
	if err != nil {
		return fmt.Errorf("failed to get active subscribers for newsletter %s: %w", post.NewsletterID, err)
	}

	if len(activeSubscribers) == 0 {
		// No active subscribers, still mark as published.
		// Log this information.
		fmt.Printf("No active subscribers for newsletter %s to send post %s\n", post.NewsletterID, postID)
		// Fall through to mark as published.
	}

	// 3. Iterate and send emails
	// TODO: Consider making email sending asynchronous (e.g., goroutines, message queue) for many subscribers.
	// For now, synchronous sending.
	appBaseURL := os.Getenv("APP_BASE_URL") // e.g., http://localhost:8080
	if appBaseURL == "" {
		appBaseURL = "http://localhost:8080" // Default if not set
		fmt.Println("Warning: APP_BASE_URL not set, defaulting to http://localhost:8080 for unsubscribe links.")
	}

	var emailSendErrors []error
	for _, subscriber := range activeSubscribers {
		if subscriber.UnsubscribeToken == "" {
			// This should ideally not happen if tokens are always generated.
			// Log and skip this subscriber for this email batch.
			fmt.Printf("Warning: Subscriber %s (ID: %s) missing unsubscribe token. Skipping email for post %s.\n", subscriber.Email, subscriber.ID, postID)
			continue
		}
		unsubscribeLink := fmt.Sprintf("%s/api/subscriptions/unsubscribe?token=%s", appBaseURL, subscriber.UnsubscribeToken)
		
		// Using subscriber's email as recipient name if no other name field exists
		err := s.emailService.SendNewsletterIssue(subscriber.Email, subscriber.Email, post.Title, post.Content, unsubscribeLink) // Use post.Content
		if err != nil {
			emailSendErrors = append(emailSendErrors, fmt.Errorf("failed to send to %s: %w", subscriber.Email, err))
			// Log individual errors but continue sending to others
			fmt.Printf("Error sending newsletter issue to %s for post %s: %v\n", subscriber.Email, postID, err)
		}
	}

	// 4. Mark post as published (e.g., set published_at timestamp)
	// This should happen regardless of email sending partial failures,
	// as the intent to publish was made. Failed emails should be logged/monitored.
	err = s.newsletterService.MarkPostAsPublished(ctx, editorFirebaseUID, postUUID) // Pass editorFirebaseUID (string) and postUUID (uuid.UUID)
	if err != nil {
		// This is a critical error if we can't mark it as published.
		// Combine with email errors if any.
		finalError := fmt.Errorf("failed to mark post %s as published: %w", postID, err)
		if len(emailSendErrors) > 0 {
			finalError = fmt.Errorf("%v; also encountered %d email sending errors", finalError, len(emailSendErrors))
		}
		return finalError
	}

	if len(emailSendErrors) > 0 {
		// Return a composite error or log that some emails failed.
		// For now, just returning a generic message indicating partial failure.
		return fmt.Errorf("post %s published, but %d email(s) failed to send", postID, len(emailSendErrors))
	}

	return nil
}

// GetPostForPublishingRequest and Response would be part of NewsletterService if we define it there.
// For now, assuming NewsletterService has a method like:
// GetPostForPublishing(ctx context.Context, postID string, editorFirebaseUID string) (*models.Post, error)
// which returns the post if the editor owns it and it's not already published.
// And models.Post has fields like NewsletterID, Title, Body.
