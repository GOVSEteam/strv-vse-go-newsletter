package firebase_auth

import (
	"context"
	"firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
	"log"
	"os"
)

var authClient *auth.Client

func InitFirebase() {
	credJSON := os.Getenv("FIREBASE_SERVICE_ACCOUNT")
	if credJSON == "" {
		log.Fatal("FIREBASE_SERVICE_ACCOUNT env-var is not set")
	}

	opt := option.WithCredentialsJSON([]byte(credJSON)) // ← replaces WithCredentialsFile
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing Firebase app: %v", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v", err)
	}
	authClient = client
}

func GetAuthClient() *auth.Client {
	if authClient == nil {
		InitFirebase()
	}
	return authClient
}
