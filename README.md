# STRV VSE Go Newsletter Service

[![Go Report Card](https://goreportcard.com/badge/github.com/GOVSEteam/strv-vse-go-newsletter)](https://goreportcard.com/report/github.com/GOVSEteam/strv-vse-go-newsletter)

<!-- Add CI Badge later -->
<!-- [![CI Status](https://github.com/GOVSEteam/strv-vse-go-newsletter/actions/workflows/ci.yml/badge.svg)](https://github.com/GOVSEteam/strv-vse-go-newsletter/actions/workflows/ci.yml) -->
<!-- Add Coverage Badge later -->

API backend for the STRV Semestral Project - Go Newsletter Platform.

## Overview

This service provides the API endpoints for:

- Editor registration and authentication.
- Managing newsletters (create, rename, delete).
- Managing posts within newsletters.
- Subscribing users to newsletters via email.
- Unsubscribing users.
- Publishing newsletter posts to subscribers via email.

## Prerequisites

- [Go](https://golang.org/doc/install) (version 1.21 or later)

## Getting Started

1.  **Clone the repository:**

    ```bash
    git clone https://github.com/GOVSEteam/strv-vse-go-newsletter.git
    cd strv-vse-go-newsletter
    ```

2.  **Configure Environment:**

    - Copy the example configuration:

    ```bash
    cp configs/.env.example configs/.env
    ```

    - Edit `configs/.env` and fill in the required values (Database credentials, Firebase details, Email service keys, JWT secrets etc. - placeholders for now).

3.  **Build Dependencies (Local):**
    _(Docker setup recommended for consistency)_

    ```bash
    # TODO: Add docker-compose up command here once created
    echo "Run 'docker-compose up -d' (once docker-compose.yml is added)"
    ```

4.  **Run Database Migrations:**
    _(Requires a migration tool like golang-migrate/migrate)_

    ```bash
    # TODO: Add migration command here once tool is chosen
    echo "Run migration command (e.g., migrate -database ... -path ... up)"
    ```

5.  **Run the Application:**

    ```bash
    # Using Go directly (ensure dependencies in .env are set)
    go run ./cmd/server/main.go

    # Or using docker-compose (preferred)
    # docker-compose up app # Assuming 'app' service name
    ```

## Running Tests

```bash
# TODO: Add test command
go test ./... -v -race -cover
```

## Commit Conventions
- **Feature**: A new feature or enhancement to existing functionality.
- **Bugfix**: A bug fix or patch to existing functionality.
- **Refactor**: Code refactoring or cleanup without changing functionality.

## Branch naming rules
Commiting into main branch is not allowed. Changes should be made in separate branches and merged via pull requests.
- **Feature branches**: `feature/<description>` (e.g., `feature/user-auth`)
- **Bugfix branches**: `bugfix/<description>` (e.g., `bugfix/fix-login-issue`)
- **Refactor branches**: `refactor/<description>` (e.g., `refactor/code-cleanup`)

## Layered Architecture
The project follows a layered architecture pattern, which separates concerns and promotes maintainability. The main layers are:
- **Router**: Handles HTTP routing and delegates to handlers.
- **Handlers**: Handle HTTP requests, parse input/output, and call services.
- **Service**: Contains business logic, orchestrates repositories.
- **Repository**: Handles data persistence.

## Deployment
App is running on https://railway.com/. The production URL is:

`strv-vse-go-newsletter-production.up.railway.app`

You can test if the APP is running by this powershell command:

```powershell
Invoke-WebRequest -Uri https://strv-vse-go-newsletter-production.up.railway.app/healthz -Method GET
```

Or by opening the URL in your browser.

## Database
The application uses **PostgreSQL** as the primary database for storing data. The database connection details are specified in the `.env` file.

Given the size of the project and number of tables, automatic migrations are not implemented. If you want to change database schema, do that manually from the web administration console or, for more advanced use-cases, by running `railway connect` (you need to have the Railway CLI installed).

## Database Setup

### Running Migrations

1. Set your DATABASE_URL environment variable
2. Run migrations:
   ```bash
   go run cmd/migrate/main.go up
   ```

### Development Workflow

- Create new migration: `goose -dir migrations create migration_name sql`
- Apply migrations: `go run cmd/migrate/main.go up`
- Check status: `go run cmd/migrate/main.go status`


