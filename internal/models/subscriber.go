package models

import (
	"fmt"
	"net/mail"
	"strings"
	"time"

	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
)

// SubscriberStatus defines the possible statuses of a newsletter subscription.
type SubscriberStatus string

const (
	// SubscriberStatusActive indicates the subscription is active.
	SubscriberStatusActive       SubscriberStatus = "active"
	// SubscriberStatusUnsubscribed indicates the user has unsubscribed.
	SubscriberStatusUnsubscribed SubscriberStatus = "unsubscribed"
)

// Subscriber represents a subscriber to a newsletter
type Subscriber struct {
	ID               string           `json:"id"`
	Email            string           `json:"email"`
	NewsletterID     string           `json:"newsletter_id"`      // Consistent snake_case naming
	SubscriptionDate time.Time        `json:"subscription_date"`  // Consistent snake_case naming
	Status           SubscriberStatus `json:"status"`
	UnsubscribeToken string           `json:"-"` // Token for one-click unsubscribe (omit from JSON)
}

// Validate checks the subscriber's fields for business validation
func (s *Subscriber) Validate() error {
	trimmedEmail := strings.TrimSpace(s.Email)
	if trimmedEmail == "" {
		return apperrors.WrapValidation(nil, "email is required")
	}
	if _, err := mail.ParseAddress(trimmedEmail); err != nil {
		return apperrors.ErrInvalidEmail
	}
	
	if strings.TrimSpace(s.NewsletterID) == "" {
		return apperrors.WrapValidation(nil, "newsletter ID is required")
	}
	
	if s.Status != SubscriberStatusActive && s.Status != SubscriberStatusUnsubscribed {
		return apperrors.WrapValidation(nil, fmt.Sprintf("invalid subscriber status: %s", s.Status))
	}
	
	return nil
}
