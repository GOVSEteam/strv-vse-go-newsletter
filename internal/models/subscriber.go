package models

import "time"

// SubscriberStatus defines the possible statuses of a newsletter subscription.
type SubscriberStatus string

const (
	// SubscriberStatusPendingConfirmation indicates the subscription is awaiting email confirmation.
	SubscriberStatusPendingConfirmation SubscriberStatus = "pending_confirmation"
	// SubscriberStatusActive indicates the subscription is active.
	SubscriberStatusActive SubscriberStatus = "active"
	// SubscriberStatusUnsubscribed indicates the user has unsubscribed.
	SubscriberStatusUnsubscribed SubscriberStatus = "unsubscribed"
)

// Subscriber represents a subscriber to a newsletter in Firestore.
type Subscriber struct {
	ID                 string           `json:"id" firestore:"id,omitempty"`                         // Firestore document ID
	Email              string           `json:"email" firestore:"email"`                            // Email of the subscriber
	NewsletterID       string           `json:"newsletter_id" firestore:"newsletter_id"`            // ID of the newsletter subscribed to
	SubscriptionDate   time.Time        `json:"subscription_date" firestore:"subscription_date"`  // Timestamp of subscription
	Status             SubscriberStatus `json:"status" firestore:"status"`                         // Status of the subscription
	ConfirmationToken  string           `json:"-" firestore:"confirmation_token,omitempty"`      // Token for email confirmation (omit from JSON responses)
	TokenExpiryTime    time.Time        `json:"-" firestore:"token_expiry_time,omitempty"`   // Expiry time for the confirmation token
	ConfirmedAt        *time.Time       `json:"confirmed_at,omitempty" firestore:"confirmed_at,omitempty"` // Timestamp of email confirmation
	// UnsubscribedAt    *time.Time     `json:"unsubscribed_at,omitempty" firestore:"unsubscribed_at,omitempty"` // Using status field primarily
} 