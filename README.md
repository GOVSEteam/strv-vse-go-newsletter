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
- [Docker](https://docs.docker.com/get-docker/) & [Docker Compose](https://docs.docker.com/compose/install/)
- (Optional) [Task](https://taskfile.dev/installation/) or Make for running scripts

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

## Project Structure
- `cmd/` — Application entry points (e.g., main.go for starting the server)
- `internal/layers/router/` — HTTP router setup: configures routes and wires handlers to endpoints
- `internal/layers/handler/` — HTTP handlers: handle HTTP requests, parse input, and return responses
- `internal/layers/service/` — Business logic: implements core application logic and orchestrates repositories
- `internal/layers/repository/` — Data access: manages data storage and retrieval (e.g., in-memory, database)
- `internal/domain/` — Domain models and entities for the application's core business objects

## Deployment
App is running on https://railway.com/ with automatic deployment on repository commit. The production URL is:

`strv-vse-go-newsletter-production.up.railway.app`

You can test if the APP is running by this powershell command:

```powershell
Invoke-WebRequest -Uri https://strv-vse-go-newsletter-production.up.railway.app/healthz -Method GET
```

Or by opening the URL in your browser.
