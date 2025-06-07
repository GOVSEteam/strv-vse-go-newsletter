package testutils

import (
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TestSuite struct {
	DB        *pgxpool.Pool
	Firestore *firestore.Client
	Config    *config.Config
}

func NewTestSuite(t *testing.T) *TestSuite {
	cfg := LoadTestConfig(t)

	db := SetupTestDB(t)
	firestore := SetupTestFirestore(t)
	
	// Clean up any leftover test data before starting tests
	CleanupTestData(t, db)

	return &TestSuite{
		DB:        db,
		Firestore: firestore,
		Config:    cfg,
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