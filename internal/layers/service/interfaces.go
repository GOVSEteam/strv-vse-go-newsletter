package service

import (
	"context"

	"firebase.google.com/go/v4/auth"
)

// FirebaseAuthClient defines the interface for Firebase authentication operations.
// This interface abstracts the Firebase auth.Client to enable testing and follow
// the dependency inversion principle.
type FirebaseAuthClient interface {
	// CreateUser creates a new user in Firebase Auth with the given parameters
	CreateUser(ctx context.Context, user *auth.UserToCreate) (*auth.UserRecord, error)
	
	// VerifyIDToken verifies a Firebase ID token and returns the decoded token
	VerifyIDToken(ctx context.Context, idToken string) (*auth.Token, error)
}

// Ensure that Firebase's *auth.Client satisfies our interface at compile time
var _ FirebaseAuthClient = (*auth.Client)(nil) 