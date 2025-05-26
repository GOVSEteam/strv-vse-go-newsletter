package testutils

import (
	"context"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/setup"
)

func SetupTestFirestore(t *testing.T) *firestore.Client {
	client := setup.GetFirestoreClient()
	return client
}

func CleanupTestFirestoreData(t *testing.T, client *firestore.Client) {
	ctx := context.Background()

	// Clean up test subscribers
	iter := client.Collection("subscribers").
		Where("email", ">=", "test_").
		Where("email", "<", "tesu").
		Documents(ctx)

	batch := client.Batch()
	batchSize := 0

	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		batch.Delete(doc.Ref)
		batchSize++

		// Firestore batch limit is 500
		if batchSize >= 500 {
			if _, err := batch.Commit(ctx); err != nil {
				t.Logf("Warning: Failed to cleanup Firestore batch: %v", err)
			}
			batch = client.Batch()
			batchSize = 0
		}
	}

	if batchSize > 0 {
		if _, err := batch.Commit(ctx); err != nil {
			t.Logf("Warning: Failed to cleanup final Firestore batch: %v", err)
		}
	}
} 