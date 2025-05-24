package repository

import (
	"context"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models" // Assuming this path is correct
	"google.golang.org/api/iterator"                              // For checking iterator.Done
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const subscribersCollection = "subscribers"

// SubscriberRepository defines the interface for subscriber data persistence.
type SubscriberRepository interface {
	CreateSubscriber(ctx context.Context, subscriber models.Subscriber) (string, error)
	GetSubscriberByEmailAndNewsletterID(ctx context.Context, email string, newsletterID string) (*models.Subscriber, error)
	UpdateSubscriberStatus(ctx context.Context, subscriberID string, status models.SubscriberStatus) error
	GetSubscriberByConfirmationToken(ctx context.Context, token string) (*models.Subscriber, error)
	ConfirmSubscriber(ctx context.Context, subscriberID string, confirmationTime time.Time) error
	GetSubscriberByUnsubscribeToken(ctx context.Context, token string) (*models.Subscriber, error)                 // New method
	GetActiveSubscribersByNewsletterID(ctx context.Context, newsletterID string) ([]models.Subscriber, error) // New method
	// TODO: Add IsSubscribed(ctx context.Context, email, newsletterID string) (bool, error)
}

// firestoreSubscriberRepository implements SubscriberRepository using Firestore.
type firestoreSubscriberRepository struct {
	client *firestore.Client
}

// NewFirestoreSubscriberRepository creates a new firestoreSubscriberRepository.
func NewFirestoreSubscriberRepository(client *firestore.Client) SubscriberRepository {
	if client == nil {
		// This should ideally be handled by a panic or a more robust error at startup
		// if the client isn't available, but for now, we'll return nil
		// or a repository that always errors. For simplicity, let's assume client is always provided.
		// log.Fatal("Firestore client is nil in NewFirestoreSubscriberRepository")
	}
	return &firestoreSubscriberRepository{client: client}
}

// CreateSubscriber adds a new subscriber document to the Firestore "subscribers" collection.
// It returns the ID of the newly created document or an error.
func (r *firestoreSubscriberRepository) CreateSubscriber(ctx context.Context, subscriber models.Subscriber) (string, error) {
	// We might want to check for existing exact email + newsletterID combination here
	// to prevent duplicates, or handle it at the service layer.
	// For now, we'll assume Firestore generates a unique ID or we ensure uniqueness elsewhere.

	docRef, _, err := r.client.Collection(subscribersCollection).Add(ctx, subscriber)
	if err != nil {
		// It's good practice to wrap or check for specific Firestore errors if needed.
		// For example, checking for codes.AlreadyExists if we were setting a specific document ID.
		return "", status.Errorf(codes.Internal, "failed to create subscriber in Firestore: %v", err)
	}
	return docRef.ID, nil
}

// GetSubscriberByEmailAndNewsletterID retrieves a subscriber by their email and a specific newsletter ID.
// Returns (nil, nil) if not found, or the subscriber and nil error if found.
func (r *firestoreSubscriberRepository) GetSubscriberByEmailAndNewsletterID(ctx context.Context, email string, newsletterID string) (*models.Subscriber, error) {
	iter := r.client.Collection(subscribersCollection).
		Where("email", "==", email).
		Where("newsletter_id", "==", newsletterID).
		Limit(1).
		Documents(ctx)

	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query Firestore for subscriber: %v", err)
	}

	var subscriber models.Subscriber
	if err := doc.DataTo(&subscriber); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to decode subscriber data: %v", err)
	}
	subscriber.ID = doc.Ref.ID // Populate the ID from the document reference
	return &subscriber, nil
}

// UpdateSubscriberStatus updates the status of a specific subscriber document in Firestore.
func (r *firestoreSubscriberRepository) UpdateSubscriberStatus(ctx context.Context, subscriberID string, newStatus models.SubscriberStatus) error {
	docRef := r.client.Collection(subscribersCollection).Doc(subscriberID)

	_, err := docRef.Update(ctx, []firestore.Update{
		{
			Path:  "status",
			Value: newStatus,
		},
		// We could also update an 'unsubscribed_at' timestamp here if needed.
		// {
		// 	Path: "unsubscribed_at",
		// 	Value: time.Now().UTC(),
		// },
	})

	if err != nil {
		if status.Code(err) == codes.NotFound {
			return status.Errorf(codes.NotFound, "subscriber with ID %s not found for update", subscriberID)
		}
		return status.Errorf(codes.Internal, "failed to update subscriber status in Firestore: %v", err)
	}
	return nil
}

// GetSubscriberByUnsubscribeToken retrieves a subscriber by their unsubscribe token.
// Returns (nil, nil) if not found.
func (r *firestoreSubscriberRepository) GetSubscriberByUnsubscribeToken(ctx context.Context, token string) (*models.Subscriber, error) {
	iter := r.client.Collection(subscribersCollection).
		Where("unsubscribe_token", "==", token).
		Limit(1).
		Documents(ctx)

	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query Firestore for subscriber by unsubscribe token: %v", err)
	}

	var subscriber models.Subscriber
	if err := doc.DataTo(&subscriber); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to decode subscriber data by unsubscribe token: %v", err)
	}
	subscriber.ID = doc.Ref.ID
	return &subscriber, nil
}

// GetActiveSubscribersByNewsletterID retrieves all active subscribers for a specific newsletter ID.
func (r *firestoreSubscriberRepository) GetActiveSubscribersByNewsletterID(ctx context.Context, newsletterID string) ([]models.Subscriber, error) {
	var subscribers []models.Subscriber
	iter := r.client.Collection(subscribersCollection).
		Where("newsletter_id", "==", newsletterID).
		Where("status", "==", models.SubscriberStatusActive). // Only active subscribers
		Documents(ctx)

	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to iterate active subscribers: %v", err)
		}

		var sub models.Subscriber
		if err := doc.DataTo(&sub); err != nil {
			// Log or handle individual decoding errors, but continue if possible
			// For now, we'll return an error on the first decode failure.
			return nil, status.Errorf(codes.Internal, "failed to decode active subscriber data: %v", err)
		}
		sub.ID = doc.Ref.ID
		subscribers = append(subscribers, sub)
	}

	return subscribers, nil
}

// GetSubscriberByConfirmationToken retrieves a subscriber by their confirmation token.
// Returns (nil, nil) if not found.
func (r *firestoreSubscriberRepository) GetSubscriberByConfirmationToken(ctx context.Context, token string) (*models.Subscriber, error) {
	iter := r.client.Collection(subscribersCollection).
		Where("confirmation_token", "==", token).
		Limit(1).
		Documents(ctx)

	defer iter.Stop()

	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, nil // Not found
	}
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to query Firestore for subscriber by token: %v", err)
	}

	var subscriber models.Subscriber
	if err := doc.DataTo(&subscriber); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to decode subscriber data by token: %v", err)
	}
	subscriber.ID = doc.Ref.ID
	return &subscriber, nil
}

// ConfirmSubscriber updates the subscriber's status to active, clears the token fields, and sets the confirmation time.
func (r *firestoreSubscriberRepository) ConfirmSubscriber(ctx context.Context, subscriberID string, confirmationTime time.Time) error {
	docRef := r.client.Collection(subscribersCollection).Doc(subscriberID)

	_, err := docRef.Update(ctx, []firestore.Update{
		{Path: "status", Value: models.SubscriberStatusActive},
		{Path: "confirmation_token", Value: firestore.Delete},
		{Path: "token_expiry_time", Value: firestore.Delete},
		{Path: "confirmed_at", Value: confirmationTime},
	})

	if err != nil {
		if status.Code(err) == codes.NotFound {
			return status.Errorf(codes.NotFound, "subscriber with ID %s not found for confirmation update", subscriberID)
		}
		return status.Errorf(codes.Internal, "failed to confirm subscriber in Firestore: %v", err)
	}
	return nil
}
