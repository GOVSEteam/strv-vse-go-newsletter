package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"

	"github.com/GOVSEteam/strv-vse-go-newsletter/internal/config"
	_ "github.com/lib/pq" // Register pq driver for database/sql
	"github.com/pressly/goose/v3"
)

func main() {
	var dir = flag.String("dir", "migrations", "directory with migration files")
	flag.Parse()

	// Load configuration using centralized config system
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	dbURL := cfg.GetDatabaseURL()
	if dbURL == "" {
		log.Fatal("DATABASE_URL is required but not configured")
	}

	// Use pq with database/sql compatibility
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	args := flag.Args()
	if len(args) < 1 {
		log.Fatal("Usage: go run cmd/migrate/main.go [up|down|status|version]")
	}

	command := args[0]

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatalf("Failed to set dialect: %v", err)
	}

	switch command {
	case "up":
		if err := goose.Up(db, *dir); err != nil {
			log.Fatalf("Migration up failed: %v", err)
		}
		fmt.Println("Migrations applied successfully")
	case "down":
		if err := goose.Down(db, *dir); err != nil {
			log.Fatalf("Migration down failed: %v", err)
		}
		fmt.Println("Migration rolled back successfully")
	case "status":
		if err := goose.Status(db, *dir); err != nil {
			log.Fatalf("Migration status failed: %v", err)
		}
	case "version":
		version, err := goose.GetDBVersion(db)
		if err != nil {
			log.Fatalf("Failed to get version: %v", err)
		}
		fmt.Printf("Current version: %d\n", version)
	default:
		log.Fatalf("Unknown command: %s", command)
	}
}
