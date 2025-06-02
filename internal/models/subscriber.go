package models

import (
	"fmt"
	"net/mail"
	"time"
)

// SubscriberStatus defines the possible statuses of a newsletter subscription.
type SubscriberStatus string

const (
	// SubscriberStatusActive indicates the subscription is active.
	SubscriberStatusActive SubscriberStatus = "active"
	// SubscriberStatusUnsubscribed indicates the user has unsubscribed.
	SubscriberStatusUnsubscribed SubscriberStatus = "unsubscribed"
)

// Subscriber represents a subscriber to a newsletter.
type Subscriber struct {
	ID               string           `json:"id"`
	Email            string           `json:"email"`
	NewsletterID     string           `json:"newsletter_id"`
	SubscriptionDate time.Time        `json:"subscription_date"`
	Status           SubscriberStatus `json:"status"`
	UnsubscribeToken string           `json:"-"` // Token for one-click unsubscribe (omit from JSON)
}

// Validate checks the subscriber's fields for validity.
func (s *Subscriber) Validate() error {
	if s.Email == "" {
		return fmt.Errorf("email is required")
	}
	if _, err := mail.ParseAddress(s.Email); err != nil {
		return fmt.Errorf("invalid email format: %w", err)
	}
	if s.NewsletterID == "" {
		return fmt.Errorf("newsletter_id is required")
	}
	if s.Status != SubscriberStatusActive && s.Status != SubscriberStatusUnsubscribed {
		return fmt.Errorf("invalid subscriber status: %s", s.Status)
	}
	return nil
}

// IsActive returns true if the subscriber's status is active.
func (s *Subscriber) IsActive() bool {
	return s.Status == SubscriberStatusActive
}

// CanUnsubscribe returns true if the subscriber is active and can unsubscribe.
// This can be extended with more complex logic if needed (e.g., based on subscription date).
func (s *Subscriber) CanUnsubscribe() bool {
	return s.IsActive()
}
