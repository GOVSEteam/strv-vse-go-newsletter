# STRV VSE Go Newsletter Service

[![Go Report Card](https://goreportcard.com/badge/github.com/GOVSEteam/strv-vse-go-newsletter)](https://goreportcard.com/report/github.com/GOVSEteam/strv-vse-go-newsletter)
[![Go Version](https://img.shields.io/badge/Go-1.23+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Railway Deploy](https://img.shields.io/badge/Deploy-Railway-0B0D0E?style=flat&logo=railway)](https://railway.app)
[![API Docs](https://img.shields.io/badge/API-Swagger-85EA2D?style=flat&logo=swagger)](https://strv-vse-go-newsletter-production.up.railway.app/swagger/index.html)
[![Ask DeepWiki](https://deepwiki.com/badge.svg)](https://deepwiki.com/GOVSEteam/strv-vse-go-newsletter)

API backend for the STRV Semestral Project - Go Newsletter Platform.

## Overview

This service provides API endpoints for:

- Editor registration and authentication (Firebase)
- Managing newsletters (create, update, delete, list, get)
- Managing posts within newsletters (create, update, delete, list, get, publish)
- Subscribing users to newsletters via email (with confirmation and unsubscribe)
- Publishing newsletter posts to subscribers via email (HTML, async)
- Listing newsletter subscribers (with pagination)

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
   DATABASE_PUBLIC_URL=your_public_postgres_url # (optional, for public access)

   # Firebase Authentication
   FIREBASE_SERVICE_ACCOUNT=your_firebase_service_account_json # (or use file or base64, see below)
   FIREBASE_SERVICE_ACCOUNT_BASE64=your_base64_encoded_service_account_json # (optional)
   FIREBASE_API_KEY=your_firebase_api_key

   # Email Service
   EMAIL_FROM=your_email@domain.com
   GOOGLE_APP_PASSWORD=your_app_password
   SMTP_HOST=smtp.gmail.com # (default)
   SMTP_PORT=587           # (default)

   # Application
   APP_BASE_URL=http://localhost:8080
   PORT=8080
   RAILWAY_ENVIRONMENT= # (optional, for Railway deployments)
   ```
   **Firebase Service Account loading:**
   - The service will try to load the service account in this order:
     1. `firebase-service-account.json` file in the root
     2. `FIREBASE_SERVICE_ACCOUNT_BASE64` (base64-encoded JSON)
     3. `FIREBASE_SERVICE_ACCOUNT` (plain JSON string)

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

This service implements **Clean Architecture** with dependency injection and layered design:

**Request Flow:**
- **Chi Router** → **Middleware** → **Handlers** → **Services** → **Repositories** → **Database/External APIs**

**Key Architectural Features:**
- ✅ **Dependency Injection**: Constructor-based DI, no global state
- ✅ **Clean Separation**: Pure domain models without persistence concerns  
- ✅ **Interface-Driven**: All dependencies injected via interfaces
- ✅ **Background Processing**: Async email delivery (currently direct, can be switched to worker)
- ✅ **Structured Logging**: Request correlation with Zap logger
- ✅ **Production Ready**: Health checks, graceful shutdown, panic recovery
- ✅ **Pagination**: All list endpoints support `limit` and `offset` query params
- ✅ **HTML Emails**: All emails (confirmation, newsletter) are sent as HTML
- ✅ **GDPR Support**: Unsubscribe and data deletion endpoints

**Core Packages:**
- `cmd/server` - Main application entry point with DI setup
- `cmd/migrate` - Database migration tool
- `internal/config` - Centralized configuration management
- `internal/layers/handler` - HTTP request handlers
- `internal/layers/service` - Business logic layer
- `internal/layers/repository` - Data access layer (PostgreSQL, Firestore)
- `internal/middleware` - Authentication, logging, recovery, CORS
- `internal/models` - Pure domain models
- `internal/errors` - Centralized error definitions

**Technology Stack:**
- **Router**: Chi v5 with middleware chains
- **Database**: PostgreSQL with pgx/v5 connection pooling
- **NoSQL**: Firestore for subscriber management
- **Authentication**: Firebase Auth with JWT verification
- **Email**: Gmail SMTP with HTML templates
- **Logging**: Structured logging with Zap

## Major API Endpoints

### Public
- `POST   /api/editor/signup` — Register new editor
- `POST   /api/editor/signin` — Editor login
- `POST   /api/editor/password-reset` — Request password reset
- `POST   /api/newsletters/{newsletterID}/subscribe` — Subscribe to newsletter
- `GET    /api/subscriptions/unsubscribe` — Unsubscribe via token

### Protected (require editor JWT)
- `GET    /api/newsletters` — List newsletters (with pagination)
- `POST   /api/newsletters` — Create newsletter
- `GET    /api/newsletters/{newsletterID}` — Get newsletter by ID
- `PATCH  /api/newsletters/{newsletterID}` — Update newsletter
- `DELETE /api/newsletters/{newsletterID}` — Delete newsletter
- `GET    /api/newsletters/{newsletterID}/subscribers` — List subscribers (with pagination)
- `POST   /api/newsletters/{newsletterID}/posts` — Create post
- `GET    /api/newsletters/{newsletterID}/posts` — List posts (with pagination)
- `GET    /api/posts/{postID}` — Get post by ID
- `PUT    /api/posts/{postID}` — Update post
- `DELETE /api/posts/{postID}` — Delete post
- `POST   /api/posts/{postID}/publish` — Publish post (sends to all active subscribers)

### Health
- `GET    /health` — Health check (returns OK if DB is up)

## Deployment

**Production URL:** https://strv-vse-go-newsletter-production.up.railway.app

**Health check:**
```bash
curl https://strv-vse-go-newsletter-production.up.railway.app/health
```

**Database:** PostgreSQL on Railway (managed migrations via Railway console)

## Prerequisites

- Go 1.21+
- PostgreSQL database
- Firebase project (for authentication)
- Gmail account (for email service)

## Notes
- All environment variables are loaded via `internal/config` and validated at startup.
- Firebase service account can be provided as a file, base64 env, or plain env var.
- All email sending is currently direct (sync), but can be switched to async worker.
- All list endpoints support pagination via `limit` and `offset` query params.
- All endpoints return JSON responses.
