package setup

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"os"
)

// ConnectDB returns a connected *sql.DB using env vars for config
func ConnectDB() *sql.DB {
	dbURL := os.Getenv("DATABASE_URL")
	if os.Getenv("RAILWAY_ENVIRONMENT") == "" {
		if publicURL := os.Getenv("DATABASE_PUBLIC_URL"); publicURL != "" {
			dbURL = publicURL
		}
	}

	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	if err := db.Ping(); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	return db
}
