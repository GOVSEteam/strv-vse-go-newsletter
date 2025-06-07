package testutils

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/config"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/setup"
)

// SetupTestFirestore creates a Firestore client for testing.
// It prefers the Firestore emulator if FIRESTORE_EMULATOR_HOST is set,
// otherwise falls back to using service account credentials from config.
func SetupTestFirestore(t *testing.T) *firestore.Client {
	ctx := context.Background()

	// Check if Firestore emulator is available (test-specific override)
	emulatorHost := os.Getenv("FIRESTORE_EMULATOR_HOST")
	if emulatorHost != "" {
		t.Logf("Using Firestore emulator at %s", emulatorHost)
		
		// Create client for emulator
		client, err := firestore.NewClient(ctx, "newsletter-service-test")
		if err != nil {
			t.Fatalf("Failed to create Firestore emulator client: %v", err)
		}
		return client
	}

	// Load configuration using centralized config system
	cfg, err := config.Load()
	if err != nil {
		t.Fatalf("Failed to load test configuration: %v", err)
	}

	// Use service account from centralized config
	serviceAccountJSON := cfg.FirebaseServiceAccount
	if serviceAccountJSON == "" {
		t.Skip("Skipping Firestore tests: FIREBASE_SERVICE_ACCOUNT is not configured")
	}

	// Initialize Firebase app
	app, err := setup.NewFirebaseApp(ctx, serviceAccountJSON)
	if err != nil {
		t.Fatalf("Failed to initialize Firebase app for tests: %v", err)
	}

	// Create Firestore client
	client, err := setup.NewFirestoreClient(ctx, app)
	if err != nil {
		t.Fatalf("Failed to create Firestore client for tests: %v", err)
	}

	return client
}

// CleanupTestFirestoreData removes test data from Firestore.
// This function cleans up any data with test prefixes.
func CleanupTestFirestoreData(t *testing.T, client *firestore.Client) {
	if client == nil {
		t.Log("Warning: No Firestore client provided for cleanup")
		return
	}
	
	ctx := context.Background()

	// Clean up test subscribers - simple approach for test cleanup
	iter := client.Collection("subscribers").
		Where("email", ">=", "test_").
		Where("email", "<", "tesu").
		Documents(ctx)

	// Delete documents one by one - simpler than batching for test cleanup
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		if _, err := doc.Ref.Delete(ctx); err != nil {
			t.Logf("Warning: Failed to cleanup Firestore document %s: %v", doc.Ref.ID, err)
		}
	}
} 