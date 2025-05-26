package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSubscriberStatus_Constants(t *testing.T) {
	// Test that the constants are defined correctly
	assert.Equal(t, SubscriberStatus("pending_confirmation"), SubscriberStatusPendingConfirmation)
	assert.Equal(t, SubscriberStatus("active"), SubscriberStatusActive)
	assert.Equal(t, SubscriberStatus("unsubscribed"), SubscriberStatusUnsubscribed)
}

func TestSubscriber_Creation(t *testing.T) {
	now := time.Now()
	subscriber := Subscriber{
		ID:                "test-id-123",
		Email:             "test@example.com",
		NewsletterID:      "newsletter-123",
		SubscriptionDate:  now,
		Status:            SubscriberStatusPendingConfirmation,
		ConfirmationToken: "token-123",
		TokenExpiryTime:   now.Add(24 * time.Hour),
		UnsubscribeToken:  "unsub-token-123",
	}

	assert.Equal(t, "test-id-123", subscriber.ID)
	assert.Equal(t, "test@example.com", subscriber.Email)
	assert.Equal(t, "newsletter-123", subscriber.NewsletterID)
	assert.Equal(t, SubscriberStatusPendingConfirmation, subscriber.Status)
	assert.Equal(t, "token-123", subscriber.ConfirmationToken)
	assert.Equal(t, "unsub-token-123", subscriber.UnsubscribeToken)
	assert.Nil(t, subscriber.ConfirmedAt) // Should be nil for pending confirmation
}

func TestSubscriber_ConfirmedState(t *testing.T) {
	now := time.Now()
	confirmedAt := now.Add(-time.Hour) // Confirmed an hour ago
	
	subscriber := Subscriber{
		ID:           "test-id-123",
		Email:        "test@example.com",
		NewsletterID: "newsletter-123",
		Status:       SubscriberStatusActive,
		ConfirmedAt:  &confirmedAt,
	}

	assert.Equal(t, SubscriberStatusActive, subscriber.Status)
	assert.NotNil(t, subscriber.ConfirmedAt)
	assert.True(t, subscriber.ConfirmedAt.Before(now))
}

func TestSubscriber_UnsubscribedState(t *testing.T) {
	subscriber := Subscriber{
		ID:           "test-id-123",
		Email:        "test@example.com",
		NewsletterID: "newsletter-123",
		Status:       SubscriberStatusUnsubscribed,
	}

	assert.Equal(t, SubscriberStatusUnsubscribed, subscriber.Status)
} 