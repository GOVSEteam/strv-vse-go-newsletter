package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/config"
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
	emailService      EmailService               // For direct email sending
	config            *config.Config             // Application configuration
}

// Errors
var ErrPostAlreadyPublished = errors.New("post already published")

// NewPublishingService creates a new PublishingService.
func NewPublishingService(
	newsletterService NewsletterServiceInterface,
	subscriberService SubscriberServiceInterface,
	emailService EmailService, // For direct email sending
	cfg *config.Config,
) PublishingServiceInterface {
	return &PublishingService{
		newsletterService: newsletterService,
		subscriberService: subscriberService,
		emailService:      emailService,
		config:            cfg,
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
		return ErrPostAlreadyPublished
	}

	// 2. Get active subscribers for the newsletter
	// Use the efficient method that gets all active subscribers without pagination overhead
	activeSubscribers, err := s.subscriberService.GetActiveSubscribersForNewsletter(ctx, post.NewsletterID)
	if err != nil {
		return fmt.Errorf("failed to get active subscribers for newsletter %s: %w", post.NewsletterID, err)
	}

	if len(activeSubscribers) == 0 {
		fmt.Printf("No active subscribers for newsletter %s to send post %s\n", post.NewsletterID, postID)
		// Still mark as published even if no one to send to.
	} else {
		fmt.Printf("Found %d active subscribers for newsletter %s. Enqueuing emails for post %s...\n", len(activeSubscribers), post.NewsletterID, postID)

		// For now, send emails synchronously to ensure they work
		// TODO: Revert to async email worker once the database issues are resolved
		for _, subscriber := range activeSubscribers {
			if subscriber.UnsubscribeToken == "" {
				fmt.Printf("Warning: Subscriber %s (ID: %s) missing unsubscribe token. Skipping email for post %s.\n", subscriber.Email, subscriber.ID, postID)
				continue
			}

			// Generate unsubscribe link and extract recipient name
			unsubscribeLink := fmt.Sprintf("%s/api/subscriptions/unsubscribe?token=%s", s.config.AppBaseURL, subscriber.UnsubscribeToken)
			recipientName := subscriber.Email
			if atIndex := strings.Index(subscriber.Email, "@"); atIndex > 0 {
				recipientName = subscriber.Email[:atIndex]
			}

			// Send email directly using the email service
			err := s.emailService.SendNewsletterIssueHTML(ctx, subscriber.Email, recipientName, post.Title, post.Content, unsubscribeLink)
			if err != nil {
				fmt.Printf("Warning: Failed to send email to %s for post %s: %v\n", subscriber.Email, postID, err)
				// Continue with other subscribers instead of failing completely
			} else {
				fmt.Printf("Successfully sent email to %s for post %s\n", subscriber.Email, postID)
			}
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
