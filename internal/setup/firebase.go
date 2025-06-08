package setup

import (
	"context"
	"fmt"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/layers/service"
	"google.golang.org/api/option"
)

// NewFirebaseApp initializes a Firebase app with the provided service account JSON.
// This is the foundation for all other Firebase services.
func NewFirebaseApp(ctx context.Context, serviceAccountJSON string) (*firebase.App, error) {
	if serviceAccountJSON == "" {
		return nil, fmt.Errorf("firebase service account JSON is required")
	}

	opt := option.WithCredentialsJSON([]byte(serviceAccountJSON))
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	return app, nil
}

// NewAuthClient initializes a Firebase Auth client from the given Firebase app.
// This client is used for JWT token verification and user management.
func NewAuthClient(ctx context.Context, app *firebase.App) (*auth.Client, error) {
	client, err := app.Auth(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firebase Auth client: %w", err)
	}
	return client, nil
}

// NewFirestoreClient initializes a Firestore client from the given Firebase app.
// This client is used for NoSQL document database operations.
func NewFirestoreClient(ctx context.Context, app *firebase.App) (*firestore.Client, error) {
	client, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Firestore client: %w", err)
	}
	return client, nil
}

// FirebaseAuthAdapter adapts Firebase auth.Client to service.AuthClient interface.
type FirebaseAuthAdapter struct {
	client *auth.Client
}

// NewFirebaseAuthAdapter creates a new Firebase auth adapter for services.
func NewFirebaseAuthAdapter(client *auth.Client) service.AuthClient {
	return &FirebaseAuthAdapter{client: client}
}

// CreateUser creates a new user in Firebase and returns domain types.
func (a *FirebaseAuthAdapter) CreateUser(ctx context.Context, req service.CreateUserRequest) (*service.CreatedUser, error) {
	// Convert domain request to Firebase types
	userToCreate := &auth.UserToCreate{}
	userToCreate.Email(req.Email)
	userToCreate.Password(req.Password)

	// Create user in Firebase
	firebaseUser, err := a.client.CreateUser(ctx, userToCreate)
	if err != nil {
		return nil, err
	}

	// Convert Firebase response to domain types
	return &service.CreatedUser{
		UID:   firebaseUser.UID,
		Email: firebaseUser.Email,
	}, nil
}
