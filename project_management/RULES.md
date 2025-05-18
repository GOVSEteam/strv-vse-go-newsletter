# Project Rules & Technical Standards: Go Newsletter Platform

## Table of Contents
1. Technology Stack
2. Naming Conventions
3. Code Style & Structure
4. API Design Standards
5. Error Handling
6. Testing
7. Documentation
8. Deployment & Environment
9. General Guidelines

---

## 1. Technology Stack
- **Language:** Go (latest stable)
- **API:** REST (JSON over HTTP)
- **Database:**
  - Editors, Newsletters, Posts: PostgreSQL (no ORMs, use database/sql or low-level query builder)
  - Subscribers: Firebase (Firestore)
- **Authentication:** JWT (stateless, via Firebase Auth)
- **Email Service:** Resend, SendGrid, or AWS SES (team to choose and document)
- **Cloud Platform:** Railway (for deployment)
- **Documentation:** Swagger/OpenAPI for API docs, Markdown for project docs

---

## 2. Naming Conventions
- **Firebase/Cloud:** `strv-vse-go-newsletter-[last_name]-[first_name]`
- **Database Tables:** snake_case, plural (e.g., `editors`, `newsletters`)
- **Go Packages:** all lowercase, no underscores or hyphens
- **Variables/Functions:** camelCase for Go, snake_case for SQL
- **API Endpoints:** kebab-case, resource-oriented (e.g., `/newsletters/{id}/subscribe`)

---

## 3. Code Style & Structure
- Follow idiomatic Go (gofmt, golint)
- Modular package structure (separate handlers, services, models, etc.)
- Use interfaces for abstractions and testing
- Avoid global state; use dependency injection where possible
- Keep functions small and focused
- Use context for request-scoped values

---

## 4. API Design Standards
- RESTful principles: resource-based, stateless, predictable URLs
- Use appropriate HTTP methods: GET, POST, PATCH, DELETE
- All endpoints return JSON
- Use standard HTTP status codes
- Version API if needed (e.g., `/v1/`)
- Secure all protected endpoints with JWT middleware
- Pagination for list endpoints (if needed)
- Input validation for all user data

---

## 5. Error Handling
- Return clear, consistent error responses (JSON with `error` and `message` fields)
- Use appropriate HTTP status codes (400, 401, 404, 409, 500, etc.)
- Log errors server-side with context
- Never expose sensitive details in error messages
- Use transactions for multi-step DB operations

---

## 6. Testing
- Unit tests for all business logic (use Go's `testing` package)
- Integration tests for API endpoints
- Mock external services (email, Firebase) in tests
- Aim for high test coverage on core logic
- Use CI for automated test runs

---

## 7. Documentation
- Maintain up-to-date API docs (Swagger/OpenAPI)
- Project README with setup, architecture, and usage
- Inline code comments for complex logic
- Document all environment variables and configuration
- Provide migration/seed scripts for DB

---

## 8. Deployment & Environment
- Use Railway for deployment (document setup)
- Store secrets in Railway environment variables (never in code)
- Use `.env` files for local development (never commit to VCS)
- Provide Dockerfile for local dev (optional but recommended)
- Document deployment process and rollback steps

---

## 9. General Guidelines
- Treat project as production-ready (robustness, reliability, maintainability)
- Prefer clarity and simplicity over cleverness
- Communicate ambiguities or blockers early
- Document all team decisions and deviations from these rules
- Review and update rules as project evolves 