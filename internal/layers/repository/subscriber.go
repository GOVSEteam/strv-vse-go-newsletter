package repository

import (
	"context"
	"fmt"
	"time"

	"cloud.google.com/go/firestore"
	"cloud.google.com/go/firestore/apiv1/firestorepb"
	apperrors "github.com/GOVSEteam/strv-vse-go-newsletter/internal/errors"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/models"
	"google.golang.org/api/iterator"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const subscribersCollection = "subscribers"

// dbSubscriber is an internal struct used for Firestore document mapping.
// It contains firestore tags for field names.
type dbSubscriber struct {
	Email            string                 `firestore:"email"`
	NewsletterID     string                 `firestore:"newsletter_id"`
	SubscriptionDate time.Time              `firestore:"subscription_date"`
	Status           models.SubscriberStatus `firestore:"status"`
	UnsubscribeToken string                 `firestore:"unsubscribe_token,omitempty"`
	// ID is the Firestore document ID and is not stored as a field in the document.
}

// toDomain converts a dbSubscriber (and its Firestore document ID) to a models.Subscriber.
func (dbS *dbSubscriber) toDomain(docID string) models.Subscriber {
	return models.Subscriber{
		ID:               docID,
		Email:            dbS.Email,
		NewsletterID:     dbS.NewsletterID,
		SubscriptionDate: dbS.SubscriptionDate,
		Status:           dbS.Status,
		UnsubscribeToken: dbS.UnsubscribeToken,
	}
}

// fromDomain converts a models.Subscriber to a map suitable for creating/updating a Firestore document.
// The ID field from models.Subscriber is ignored as it's managed as the Firestore document ID.
func fromDomain(s models.Subscriber) map[string]interface{} {
	return map[string]interface{}{
		"email":             s.Email,
		"newsletter_id":     s.NewsletterID,
		"subscription_date": s.SubscriptionDate,
		"status":            s.Status,
		"unsubscribe_token": s.UnsubscribeToken,
	}
}

// SubscriberRepository defines the interface for subscriber data persistence.
type SubscriberRepository interface {
	CreateSubscriber(ctx context.Context, subscriber models.Subscriber) (string, error)
	GetSubscriberByEmailAndNewsletterID(ctx context.Context, email string, newsletterID string) (*models.Subscriber, error)
	ListSubscribersByNewsletterID(ctx context.Context, newsletterID string, limit int, offset int) ([]models.Subscriber, int, error)
	ListActiveSubscribersByNewsletterID(ctx context.Context, newsletterID string, limit int, offset int) ([]models.Subscriber, int, error)
	GetAllActiveSubscribersByNewsletterID(ctx context.Context, newsletterID string) ([]models.Subscriber, error)
	UpdateSubscriberStatus(ctx context.Context, subscriberID string, status models.SubscriberStatus) error
	UpdateSubscriberUnsubscribeToken(ctx context.Context, subscriberID string, newToken string) error
	GetSubscriberByUnsubscribeToken(ctx context.Context, token string) (*models.Subscriber, error)
	DeleteAllSubscribersByNewsletterID(ctx context.Context, newsletterID string) error
}

// firestoreSubscriberRepository implements SubscriberRepository using Firestore.
type firestoreSubscriberRepository struct {
	client *firestore.Client
}

// NewFirestoreSubscriberRepository creates a new firestoreSubscriberRepository.
func NewFirestoreSubscriberRepository(client *firestore.Client) SubscriberRepository {
	return &firestoreSubscriberRepository{client: client}
}

func (r *firestoreSubscriberRepository) CreateSubscriber(ctx context.Context, subscriber models.Subscriber) (string, error) {
	dataToWrite := fromDomain(subscriber)
	docRef, _, err := r.client.Collection(subscribersCollection).Add(ctx, dataToWrite)
	if err != nil {
		// Consider checking for specific gRPC codes if applicable for Add (e.g., permission denied)
		return "", fmt.Errorf("subscriber repo: CreateSubscriber: %w: %v", apperrors.ErrInternal, err)
	}
	return docRef.ID, nil
}

func (r *firestoreSubscriberRepository) GetSubscriberByEmailAndNewsletterID(ctx context.Context, email string, newsletterID string) (*models.Subscriber, error) {
	iter := r.client.Collection(subscribersCollection).
		Where("email", "==", email).
		Where("newsletter_id", "==", newsletterID).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			return nil, fmt.Errorf("subscriber repo: GetSubscriberByEmailAndNewsletterID: %w", apperrors.ErrSubscriberNotFound)
		}
		return nil, fmt.Errorf("subscriber repo: GetSubscriberByEmailAndNewsletterID: query: %w: %v", apperrors.ErrInternal, err)
	}

	var dbSub dbSubscriber
	if errData := doc.DataTo(&dbSub); errData != nil {
		return nil, fmt.Errorf("subscriber repo: GetSubscriberByEmailAndNewsletterID: decode: %w: %v", apperrors.ErrInternal, errData)
	}
	modelSub := dbSub.toDomain(doc.Ref.ID)
	return &modelSub, nil
}

func (r *firestoreSubscriberRepository) ListSubscribersByNewsletterID(ctx context.Context, newsletterID string, limit int, offset int) ([]models.Subscriber, int, error) {
	collRef := r.client.Collection(subscribersCollection)

	// Get total count
	// Build the aggregation query from the base collection reference, then apply filters for the aggregation.
	countQuery := collRef.Where("newsletter_id", "==", newsletterID)
	aggregationQuery := countQuery.NewAggregationQuery().WithCount("all")
	countSnapshot, err := aggregationQuery.Get(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("subscriber repo: ListSubscribersByNewsletterID: count query: %w: %v", apperrors.ErrInternal, err)
	}
	countValue, ok := countSnapshot["all"]
	if !ok {
		return nil, 0, fmt.Errorf("subscriber repo: ListSubscribersByNewsletterID: count aggregation did not return 'all' field: %w", apperrors.ErrInternal)
	}
	
	// Handle Firestore aggregation result properly
	var totalCount int
	if countResult, ok := countValue.(*firestorepb.Value); ok {
		if countResult.GetIntegerValue() != 0 {
			totalCount = int(countResult.GetIntegerValue())
		} else {
			totalCount = 0
		}
	} else if countInt, ok := countValue.(int64); ok {
		totalCount = int(countInt)
	} else {
		return nil, 0, fmt.Errorf("subscriber repo: ListSubscribersByNewsletterID: unexpected count value type: %T", countValue)
	}

	// Get paginated list
	// Note: For now we'll remove OrderBy to avoid composite index requirement
	// In production, you would create the composite index: newsletter_id + subscription_date
	query := collRef.
		Where("newsletter_id", "==", newsletterID).
		Offset(offset).
		Limit(limit)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var subscribers []models.Subscriber
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, 0, fmt.Errorf("subscriber repo: ListSubscribersByNewsletterID: iterate: %w: %v", apperrors.ErrInternal, err)
		}
		var dbSub dbSubscriber
		if errData := doc.DataTo(&dbSub); errData != nil {
			// Log potentially and consider if a single decode error should fail the whole list
			return nil, 0, fmt.Errorf("subscriber repo: ListSubscribersByNewsletterID: decode: %w: %v", apperrors.ErrInternal, errData)
		}
		subscribers = append(subscribers, dbSub.toDomain(doc.Ref.ID))
	}

	return subscribers, totalCount, nil
}

func (r *firestoreSubscriberRepository) ListActiveSubscribersByNewsletterID(ctx context.Context, newsletterID string, limit int, offset int) ([]models.Subscriber, int, error) {
	collRef := r.client.Collection(subscribersCollection)

	// Get total count of ACTIVE subscribers only
	countQuery := collRef.
		Where("newsletter_id", "==", newsletterID).
		Where("status", "==", models.SubscriberStatusActive)
	aggregationQuery := countQuery.NewAggregationQuery().WithCount("all")
	countSnapshot, err := aggregationQuery.Get(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("subscriber repo: ListActiveSubscribersByNewsletterID: count query: %w: %v", apperrors.ErrInternal, err)
	}
	countValue, ok := countSnapshot["all"]
	if !ok {
		return nil, 0, fmt.Errorf("subscriber repo: ListActiveSubscribersByNewsletterID: count aggregation did not return 'all' field: %w", apperrors.ErrInternal)
	}
	
	// Handle Firestore aggregation result properly
	var totalCount int
	if countResult, ok := countValue.(*firestorepb.Value); ok {
		if countResult.GetIntegerValue() != 0 {
			totalCount = int(countResult.GetIntegerValue())
		} else {
			totalCount = 0
		}
	} else if countInt, ok := countValue.(int64); ok {
		totalCount = int(countInt)
	} else {
		return nil, 0, fmt.Errorf("subscriber repo: ListActiveSubscribersByNewsletterID: unexpected count value type: %T", countValue)
	}

	// Get paginated list of ACTIVE subscribers only
	// Note: For now we'll remove OrderBy to avoid composite index requirement
	// In production, you would create the composite index: newsletter_id + status + subscription_date
	query := collRef.
		Where("newsletter_id", "==", newsletterID).
		Where("status", "==", models.SubscriberStatusActive).
		Offset(offset).
		Limit(limit)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var activeSubscribers []models.Subscriber
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, 0, fmt.Errorf("subscriber repo: ListActiveSubscribersByNewsletterID: iterate: %w: %v", apperrors.ErrInternal, err)
		}
		var dbSub dbSubscriber
		if errData := doc.DataTo(&dbSub); errData != nil {
			return nil, 0, fmt.Errorf("subscriber repo: ListActiveSubscribersByNewsletterID: decode: %w: %v", apperrors.ErrInternal, errData)
		}
		activeSubscribers = append(activeSubscribers, dbSub.toDomain(doc.Ref.ID))
	}

	return activeSubscribers, totalCount, nil
}

func (r *firestoreSubscriberRepository) GetAllActiveSubscribersByNewsletterID(ctx context.Context, newsletterID string) ([]models.Subscriber, error) {
	// Query for ALL active subscribers without pagination (used for bulk email operations)
	// Note: Removed OrderBy to avoid requiring a composite index for newsletter_id + status + subscription_date
	// For email publishing, the order doesn't matter
	query := r.client.Collection(subscribersCollection).
		Where("newsletter_id", "==", newsletterID).
		Where("status", "==", models.SubscriberStatusActive)

	iter := query.Documents(ctx)
	defer iter.Stop()

	var activeSubscribers []models.Subscriber
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("subscriber repo: GetAllActiveSubscribersByNewsletterID: iterate: %w: %v", apperrors.ErrInternal, err)
		}
		var dbSub dbSubscriber
		if errData := doc.DataTo(&dbSub); errData != nil {
			return nil, fmt.Errorf("subscriber repo: GetAllActiveSubscribersByNewsletterID: decode: %w: %v", apperrors.ErrInternal, errData)
		}
		activeSubscribers = append(activeSubscribers, dbSub.toDomain(doc.Ref.ID))
	}

	return activeSubscribers, nil
}

func (r *firestoreSubscriberRepository) UpdateSubscriberStatus(ctx context.Context, subscriberID string, newStatus models.SubscriberStatus) error {
	updates := []firestore.Update{
		{Path: "status", Value: newStatus},
		// {Path: "updated_at", Value: firestore.ServerTimestamp}, // If an updated_at field were present
	}
	_, err := r.client.Collection(subscribersCollection).Doc(subscriberID).Update(ctx, updates)
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			return fmt.Errorf("subscriber repo: UpdateSubscriberStatus: %w: id %s", apperrors.ErrSubscriberNotFound, subscriberID)
		}
		return fmt.Errorf("subscriber repo: UpdateSubscriberStatus: %w: %v", apperrors.ErrInternal, err)
	}
	return nil
}

func (r *firestoreSubscriberRepository) UpdateSubscriberUnsubscribeToken(ctx context.Context, subscriberID string, newToken string) error {
	updates := []firestore.Update{
		{Path: "unsubscribe_token", Value: newToken},
		// {Path: "updated_at", Value: firestore.ServerTimestamp}, // If an updated_at field were present
	}
	_, err := r.client.Collection(subscribersCollection).Doc(subscriberID).Update(ctx, updates)
	if err != nil {
		st, _ := status.FromError(err)
		if st.Code() == codes.NotFound {
			return fmt.Errorf("subscriber repo: UpdateSubscriberUnsubscribeToken: %w: id %s", apperrors.ErrSubscriberNotFound, subscriberID)
		}
		return fmt.Errorf("subscriber repo: UpdateSubscriberUnsubscribeToken: %w: %v", apperrors.ErrInternal, err)
	}
	return nil
}

func (r *firestoreSubscriberRepository) GetSubscriberByUnsubscribeToken(ctx context.Context, token string) (*models.Subscriber, error) {
	iter := r.client.Collection(subscribersCollection).
		Where("unsubscribe_token", "==", token).
		Limit(1).
		Documents(ctx)
	defer iter.Stop()

	doc, err := iter.Next()
	if err != nil {
		if err == iterator.Done {
			return nil, fmt.Errorf("subscriber repo: GetSubscriberByUnsubscribeToken: %w", apperrors.ErrSubscriberNotFound)
		}
		return nil, fmt.Errorf("subscriber repo: GetSubscriberByUnsubscribeToken: query: %w: %v", apperrors.ErrInternal, err)
	}

	var dbSub dbSubscriber
	if errData := doc.DataTo(&dbSub); errData != nil {
		return nil, fmt.Errorf("subscriber repo: GetSubscriberByUnsubscribeToken: decode: %w: %v", apperrors.ErrInternal, errData)
	}
	modelSub := dbSub.toDomain(doc.Ref.ID)
	return &modelSub, nil
}

func (r *firestoreSubscriberRepository) DeleteAllSubscribersByNewsletterID(ctx context.Context, newsletterID string) error {
	iter := r.client.Collection(subscribersCollection).Where("newsletter_id", "==", newsletterID).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return fmt.Errorf("subscriber repo: DeleteAllSubscribersByNewsletterID: iterate: %w: %v", apperrors.ErrInternal, err)
		}
		if _, err := doc.Ref.Delete(ctx); err != nil {
			return fmt.Errorf("subscriber repo: DeleteAllSubscribersByNewsletterID: delete: %w: %v", apperrors.ErrInternal, err)
		}
	}
	return nil
}
