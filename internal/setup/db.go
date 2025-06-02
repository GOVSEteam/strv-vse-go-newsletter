package setup

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// ConnectDB establishes a connection pool to the PostgreSQL database using pgx.
// It accepts a context and the databaseURL as parameters.
// It returns a *pgxpool.Pool and an error if the connection fails.
func ConnectDB(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	if databaseURL == "" {
		return nil, fmt.Errorf("database URL is required")
	}

	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Configure the connection pool
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = time.Hour
	// config.MaxConnIdleTime = 30 * time.Minute // Example: close idle connections after 30 mins
	// config.HealthCheckPeriod = 5 * time.Minute // Example: perform health checks periodically

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	if err := pool.Ping(ctx); err != nil {
		// Close the pool if ping fails to prevent a resource leak
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// log.Println("Successfully connected to the database and configured connection pool.")
	return pool, nil
}
