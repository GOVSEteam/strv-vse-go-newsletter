package testutils

import (
	"database/sql"
	"testing"

	"cloud.google.com/go/firestore"
)

type TestSuite struct {
	DB        *sql.DB
	Firestore *firestore.Client
	Config    *TestConfig
}

func NewTestSuite(t *testing.T) *TestSuite {
	config := LoadTestConfig(t)
	config.Validate(t)

	return &TestSuite{
		DB:        SetupTestDB(t),
		Firestore: SetupTestFirestore(t),
		Config:    config,
	}
}

func (ts *TestSuite) Cleanup(t *testing.T) {
	CleanupTestData(t, ts.DB)
	CleanupTestFirestoreData(t, ts.Firestore)

	if ts.DB != nil {
		ts.DB.Close()
	}
	if ts.Firestore != nil {
		ts.Firestore.Close()
	}
} 