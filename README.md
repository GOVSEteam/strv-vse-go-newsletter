# STRV VSE Go Newsletter Service

[![Go Report Card](https://goreportcard.com/badge/github.com/GOVSEteam/strv-vse-go-newsletter)](https://goreportcard.com/report/github.com/GOVSEteam/strv-vse-go-newsletter)

API backend for the STRV Semestral Project - Go Newsletter Platform.

## Overview

This service provides API endpoints for:

- Editor registration and authentication (Firebase)
- Managing newsletters (create, rename, delete)
- Managing posts within newsletters
- Subscribing users to newsletters via email
- Publishing newsletter posts to subscribers via email

## Quick Start

1. **Clone and setup:**
   ```bash
   git clone https://github.com/GOVSEteam/strv-vse-go-newsletter.git
   cd strv-vse-go-newsletter
   cp .env.example .env
   ```

2. **Configure environment variables in `.env`:**
   ```bash
   # Database (PostgreSQL)
   DATABASE_URL=your_postgres_connection_string
   
   # Firebase Authentication
   FIREBASE_SERVICE_ACCOUNT=your_firebase_service_account_json
   
   # Email Service
   EMAIL_FROM=your_email@domain.com
   GOOGLE_APP_PASSWORD=your_app_password
   
   # Application
   APP_BASE_URL=http://localhost:8080
   ```

3. **Run database migrations:**
   ```bash
   go run ./cmd/migrate/main.go
   ```

4. **Start the server:**
   ```bash
   go run ./cmd/server/main.go
   ```

## API Documentation

- **Local Swagger UI**: http://localhost:8080/swagger/index.html
- **Production Swagger UI**: https://strv-vse-go-newsletter-production.up.railway.app/swagger/index.html
- **Postman Collection**: `postman/Newsletter_API_Collection.json`

## Testing

```bash
# Run all tests
go test ./...

# Run with coverage
go test ./... -v -race -cover

# Test API with Postman collection
newman run postman/Newsletter_API_Collection.json -e postman/Newsletter_API_Environment.json
```

## Architecture

**Layered Architecture:**
- **Router** → **Handlers** → **Services** → **Repository** → **Database**

**Key Packages:**
- `cmd/server` - Main application entry point
- `cmd/migrate` - Database migration tool
- `internal/layers/handler` - HTTP request handlers
- `internal/layers/service` - Business logic
- `internal/layers/repository` - Data access layer
- `internal/auth` - Firebase authentication
- `internal/pkg/email` - Email functionality

## Development

**Branch naming:**
- `feature/<description>` - New features
- `bugfix/<description>` - Bug fixes  
- `cleanup/<description>` - Code cleanup

**Commit types:** `feature`, `bugfix`, `refactor`, `cleanup`

## Deployment

**Production URL:** https://strv-vse-go-newsletter-production.up.railway.app

**Health check:**
```bash
curl https://strv-vse-go-newsletter-production.up.railway.app/healthz
```

**Database:** PostgreSQL on Railway (managed migrations via Railway console)

## Prerequisites

- Go 1.21+
- PostgreSQL database
- Firebase project (for authentication)
- Gmail account (for email service)
