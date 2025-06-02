package setup

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
)

// NewFirebaseApp initializes and returns a new Firebase app.
// It requires a context and the Firebase service account JSON as a string.
func NewFirebaseApp(ctx context.Context, serviceAccountJSON string) (*firebase.App, error) {
	if serviceAccountJSON == "" {
		return nil, fmt.Errorf("Firebase service account JSON is required")
	}

	opt := option.WithCredentialsJSON([]byte(serviceAccountJSON))
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}
	return app, nil
}

// NewAuthClient initializes and returns a new Firebase Auth client from a Firebase app.
func NewAuthClient(ctx context.Context, app *firebase.App) (*auth.Client, error) {
	if app == nil {
		return nil, fmt.Errorf("Firebase app is required to create Auth client")
	}
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase Auth client: %w", err)
	}
	return client, nil
}

// NewFirestoreClient initializes and returns a new Firestore client from a Firebase app.
func NewFirestoreClient(ctx context.Context, app *firebase.App) (*firestore.Client, error) {
	if app == nil {
		return nil, fmt.Errorf("Firebase app is required to create Firestore client")
	}
	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firestore client: %w", err)
	}
	return client, nil
}
