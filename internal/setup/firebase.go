package setup

import (
	"context"

	"cloud.google.com/go/firestore"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"google.golang.org/api/option"
	"log"
	"os"
)

var authClient *auth.Client
var firestoreClient *firestore.Client

func InitFirebase() {
	credJSON := os.Getenv("FIREBASE_SERVICE_ACCOUNT")
	if credJSON == "" {
		log.Fatal("FIREBASE_SERVICE_ACCOUNT env-var is not set")
	}

	opt := option.WithCredentialsJSON([]byte(credJSON))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		log.Fatalf("error initializing Firebase app: %v", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		log.Fatalf("error getting Auth client: %v", err)
	}
	authClient = client

	fsClient, err := app.Firestore(context.Background())
	if err != nil {
		log.Fatalf("error getting Firestore client: %v", err)
	}
	firestoreClient = fsClient
	log.Println("Firebase Auth and Firestore clients initialized.")
}

func GetAuthClient() *auth.Client {
	if authClient == nil {
		log.Println("Auth client not initialized, calling InitFirebase().")
		InitFirebase()
	}
	return authClient
}

func GetFirestoreClient() *firestore.Client {
	if firestoreClient == nil {
		log.Println("Firestore client not initialized, calling InitFirebase().")
		InitFirebase()
	}
	return firestoreClient
}
