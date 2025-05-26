package models

import "time"

// SubscriberStatus defines the possible statuses of a newsletter subscription.
type SubscriberStatus string

const (
	// SubscriberStatusActive indicates the subscription is active.
	SubscriberStatusActive SubscriberStatus = "active"
	// SubscriberStatusUnsubscribed indicates the user has unsubscribed.
	SubscriberStatusUnsubscribed SubscriberStatus = "unsubscribed"
)

// Subscriber represents a subscriber to a newsletter in Firestore.
type Subscriber struct {
	ID               string           `json:"id" firestore:"id,omitempty"`                     // Firestore document ID
	Email            string           `json:"email" firestore:"email"`                         // Email of the subscriber
	NewsletterID     string           `json:"newsletter_id" firestore:"newsletter_id"`         // ID of the newsletter subscribed to
	SubscriptionDate time.Time        `json:"subscription_date" firestore:"subscription_date"` // Timestamp of subscription
	Status           SubscriberStatus `json:"status" firestore:"status"`                       // Status of the subscription
	UnsubscribeToken string           `json:"-" firestore:"unsubscribe_token,omitempty"`       // Token for one-click unsubscribe (omit from JSON)
}
