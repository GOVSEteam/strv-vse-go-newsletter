#!/bin/bash
set -e

echo "Running database migrations..."
go run cmd/migrate/main.go up
echo "Migrations completed successfully" 