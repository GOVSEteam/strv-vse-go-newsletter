package service

import (
	"context"
	"fmt"
	"os"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	// No direct import to internal/worker needed anymore
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
	emailJobQueuer    EmailJobQueuer             // Changed to EmailJobQueuer interface
}

// NewPublishingService creates a new PublishingService.
func NewPublishingService(
	newsletterService NewsletterServiceInterface,
	subscriberService SubscriberServiceInterface,
	emailJobQueuer EmailJobQueuer, // Uses the interface
) PublishingServiceInterface {
	return &PublishingService{
		newsletterService: newsletterService,
		subscriberService: subscriberService,
		emailJobQueuer:    emailJobQueuer,
	}
}

// PublishPostToSubscribers orchestrates the process of sending a post to all active subscribers of its newsletter.
func (s *PublishingService) PublishPostToSubscribers(ctx context.Context, postID string, editorFirebaseUID string) error {
	// 1. Get Post details and verify ownership via NewsletterService
	post, err := s.newsletterService.GetPostForEditor(ctx, editorFirebaseUID, postID)
	if err != nil {
		// Handles not found, not owner, etc.
		return fmt.Errorf("failed to get post %s for editor %s: %w", postID, editorFirebaseUID, err)
	}
	if post.IsPublished() {
		// Optional: Log or return a specific status/error if already published.
		// For now, we proceed to ensure it's marked as published and attempt sending emails if that logic changes.
		// However, if it is already published, we might not want to re-send emails.
		// Let's assume for now if newsletterService.PublishPost handles this idempotently, we can call it.
		// For safety and to avoid re-sending, let's return an error or a specific status.
		fmt.Printf("Post %s is already published. No emails will be sent.\n", postID)
		// Ensure it is indeed marked as published by calling PublishPost, which should be idempotent.
		_, pubErr := s.newsletterService.PublishPost(ctx, editorFirebaseUID, postID)
		if pubErr != nil {
			return fmt.Errorf("failed to confirm already published post %s for editor %s: %w", postID, editorFirebaseUID, pubErr)
		}
		return nil // Or a specific error like apperrors.ErrConflict with message "post already published"
	}

	// 2. Get active subscribers for the newsletter
	// Fetch all active subscribers for this newsletter. Pagination limit -1 and offset 0 to fetch all.
	// The service/repo should handle -1 as "all" or use a very large number.
	activeSubscribers, _, err := s.subscriberService.ListActiveSubscribersByNewsletterID(ctx, editorFirebaseUID, post.NewsletterID, -1, 0)
	if err != nil {
		return fmt.Errorf("failed to get active subscribers for newsletter %s: %w", post.NewsletterID, err)
	}

	if len(activeSubscribers) == 0 {
		fmt.Printf("No active subscribers for newsletter %s to send post %s\n", post.NewsletterID, postID)
		// Still mark as published even if no one to send to.
	} else {
		fmt.Printf("Found %d active subscribers for newsletter %s. Enqueuing emails for post %s...\n", len(activeSubscribers), post.NewsletterID, postID)
		appBaseURL := os.Getenv("APP_BASE_URL")
		if appBaseURL == "" {
			appBaseURL = "http://localhost:8080" // Default if not set
			fmt.Println("Warning: APP_BASE_URL not set, defaulting to http://localhost:8080 for unsubscribe links.")
		}

		for _, subscriber := range activeSubscribers {
			if subscriber.UnsubscribeToken == "" {
				fmt.Printf("Warning: Subscriber %s (ID: %s) missing unsubscribe token. Skipping email for post %s.\n", subscriber.Email, subscriber.ID, postID)
				continue
			}
			unsubscribeLink := fmt.Sprintf("%s/api/subscriptions/unsubscribe?token=%s", appBaseURL, subscriber.UnsubscribeToken)

			emailJob := models.EmailJob{
				To:           subscriber.Email,
				Subject:      post.Title,       // Use post.Title for subject
				Body:         post.Content,     // Use post.Content for body
				NewsletterID: post.NewsletterID, // For tracking
				// UnsubscribeLink: unsubscribeLink, // The EmailJob itself doesn't need this; body should contain it.
			}
			// The email body should be constructed to include the unsubscribe link.
			// For now, assuming post.Content is the full HTML body that includes this.
			// A more robust solution would involve email templates.
			emailJob.Body = fmt.Sprintf("%s<br><hr><p><small>To unsubscribe, click <a href=\"%s\">here</a>.</small></p>", post.Content, unsubscribeLink)

			s.emailJobQueuer.EnqueueJob(emailJob)
			// No direct error handling for enqueue here, assuming worker handles send failures.
			// If EnqueueJob could fail (e.g., queue full and non-blocking), that would need handling.
		}
		fmt.Printf("Finished enqueuing %d emails for post %s.\n", len(activeSubscribers), postID)
	}

	// 4. Mark post as published
	// This uses the PublishPost method from NewsletterService which should handle setting published_at.
	_, err = s.newsletterService.PublishPost(ctx, editorFirebaseUID, postID)
	if err != nil {
		// This is a critical error if we can't mark it as published.
		return fmt.Errorf("failed to mark post %s as published: %w", postID, err)
	}

	fmt.Printf("Post %s successfully marked as published.\n", postID)
	// The actual email sending is asynchronous. The error from this function now primarily relates to
	// fetching data or marking the post as published, not individual email send failures.
	// Those are handled by the worker.
	return nil
}

// GetPostForPublishingRequest and Response would be part of NewsletterService if we define it there.
// For now, assuming NewsletterService has a method like:
// GetPostForPublishing(ctx context.Context, postID string, editorFirebaseUID string) (*models.Post, error)
// which returns the post if the editor owns it and it's not already published.
// And models.Post has fields like NewsletterID, Title, Body.
