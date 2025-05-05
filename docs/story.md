# User Stories: Go Newsletter Platform Backend API

**Version:** 1.0
**Project:** Go Newsletter Platform Backend API (Semester Project)

---

## Table of Contents

1.  **Introduction**
2.  **Definition of Done (DoD)**
3.  **Prioritization Approach**
4.  **User Stories & Acceptance Criteria**
    *   4.1 Epic: Project Setup & Foundation
    *   4.2 Epic: Editor Authentication & Management
    *   4.3 Epic: Newsletter Management
    *   4.4 Epic: Post Management
    *   4.5 Epic: Subscriber Management
    *   4.6 Epic: Email Sending
    *   4.7 Epic: API Gateway & Deployment
    *   4.8 Epic: Documentation & Quality
5.  **Dependency Overview**
6.  **High-Level Sprint/Iteration Mapping**

---

## 1. Introduction

This document contains a prioritized list of user stories derived from the PRD for the Go Newsletter Platform backend API project. Each story is intended to be small enough to be completed within approximately one day or less (representing roughly one story point in an agile context) and adheres to the INVEST criteria. The goal is to provide a clear, actionable backlog for the development team (the student developer).

## 2. Definition of Done (DoD)

A user story is considered "Done" when it meets all the following criteria:

*   All Acceptance Criteria (AC) for the story are met.
*   Code is written adhering to Go best practices (`gofmt`, `go vet`, linter clean).
*   Appropriate unit or integration tests are written and passing for the new functionality.
*   Code has been reviewed (self-review acceptable for solo project, peer/TA review if possible).
*   Relevant documentation (code comments, README sections if applicable) is updated.
*   The feature works correctly in the local development environment (e.g., via Docker Compose).
*   Code is committed and pushed to the main branch (or relevant feature branch).
*   *(For relevant stories)* The deployed application reflects the change correctly.

## 3. Prioritization Approach

Stories are prioritized based on a combination of factors, approximating a WSJF/RICE approach mentally:

1.  **Dependencies:** Foundational stories needed to unblock others are highest priority.
2.  **Core Value (MVP):** Stories essential for the minimal viable product (key flows like publish, subscribe) are prioritized next.
3.  **Project Requirements:** Adherence to specific technical constraints (sqlc, Firebase).
4.  **Effort/Complexity:** Already factored in by breaking stories down to ~1 point each.

Priority is indicated numerically (1 = Highest).

## 4. User Stories & Acceptance Criteria

---

### 4.1 Epic: Project Setup & Foundation

*   **Story ID:** FND-01
*   **Priority:** 1
*   **Story:** As a Developer, I want to initialize the Go module and project directory structure (`api-server`, `/cmd`, `/internal`, etc.), so that I have a standard layout for the codebase.
*   **AC:**
    *   Given a clean workspace
    *   When `git clone` and `cd` into the repo
    *   Then a `go.mod` file exists in the `api-server` root.
    *   And standard directories like `cmd/api-server`, `internal/auth`, `internal/newsletter`, `internal/db` are present.
*   **Dependencies:** None

*   **Story ID:** FND-02
*   **Priority:** 2
*   **Story:** As a Developer, I want to set up a local PostgreSQL instance using Docker, so that I have a database available for development and testing.
*   **AC:**
    *   Given Docker is installed
    *   When running `docker-compose up -d` (or equivalent command)
    *   Then a PostgreSQL container is running and accessible locally.
    *   And connection parameters (host, port, user, password, dbname) are defined and known.
*   **Dependencies:** None

*   **Story ID:** FND-03
*   **Priority:** 3
*   **Story:** As a Developer, I want to define the initial PostgreSQL schema for the `editors` table in a `.sql` file, so that the structure for storing editor data is defined.
*   **AC:**
    *   Given the project structure exists
    *   When viewing the `internal/db/schema/001_editors.sql` file (or similar)
    *   Then it contains the `CREATE TABLE editors` statement with `id` (VARCHAR, PK), `email` (VARCHAR, UNIQUE), `created_at`, `updated_at` columns.
*   **Dependencies:** FND-01

*   **Story ID:** FND-04
*   **Priority:** 4
*   **Story:** As a Developer, I want to install and configure `sqlc` with a `sqlc.yaml` file, so that I can generate type-safe Go code from SQL files.
*   **AC:**
    *   Given the project structure and Go module exist
    *   When viewing the `sqlc.yaml` file
    *   Then it correctly points to the schema and query directories.
    *   And specifies the Go package for generated code (`internal/db`).
    *   And `sqlc` is installed or installable via `go install`.
*   **Dependencies:** FND-01

*   **Story ID:** FND-05
*   **Priority:** 5
*   **Story:** As a Developer, I want to create basic SQL queries (`INSERT`, `SELECT by ID`, `SELECT by Email`, `DELETE`) for the `editors` table in a `.sql` file, so that `sqlc` can generate corresponding Go functions.
*   **AC:**
    *   Given the `editors` schema file exists (FND-03)
    *   When viewing the `internal/db/query/editor.sql` file (or similar)
    *   Then it contains named SQL queries for creating, getting (by ID, email), and deleting editors.
*   **Dependencies:** FND-03

*   **Story ID:** FND-06
*   **Priority:** 6
*   **Story:** As a Developer, I want to run `sqlc generate` successfully, so that type-safe Go code for interacting with the `editors` table is created in the specified package.
*   **AC:**
    *   Given `sqlc` is configured (FND-04) and schema/query files exist (FND-03, FND-05)
    *   When running the `sqlc generate` command (e.g., via `go generate` or Makefile)
    *   Then Go files (`models.go`, `editor.sql.go`, `db.go`) are generated in `internal/db` without errors.
    *   And the generated code compiles.
*   **Dependencies:** FND-04, FND-05

*   **Story ID:** FND-07
*   **Priority:** 7
*   **Story:** As a Developer, I want to set up basic application configuration loading from environment variables (e.g., for DB connection string, server port), so that the application is configurable without code changes.
*   **AC:**
    *   Given the Go project exists
    *   When the application starts
    *   Then it reads configuration values like `DATABASE_URL` and `PORT` from environment variables.
    *   And sensible defaults are used if variables are not set (for local dev).
*   **Dependencies:** FND-01

*   **Story ID:** FND-08
*   **Priority:** 8
*   **Story:** As a Developer, I want to integrate a basic structured logging library (e.g., `slog` or `zerolog`), so that application events and errors can be logged consistently in JSON format.
*   **AC:**
    *   Given the Go project exists
    *   When the application starts or handles a request/error
    *   Then log messages are outputted to stdout/stderr in JSON format.
    *   And basic log levels (INFO, ERROR, DEBUG) are supported.
*   **Dependencies:** FND-01

*   **Story ID:** FND-09
*   **Priority:** 9
*   **Story:** As a Developer, I want to set up a basic HTTP server using the `chi` router in `cmd/api-server/main.go`, so that the application can listen for and route incoming HTTP requests.
*   **AC:**
    *   Given the Go project exists
    *   When running `go run ./cmd/api-server`
    *   Then the application starts and listens on the configured port (from FND-07).
    *   And the `chi` router is initialized.
*   **Dependencies:** FND-01, FND-07

*   **Story ID:** FND-10
*   **Priority:** 10
*   **Story:** As a Developer, I want to create a basic `/healthz` endpoint, so that monitoring tools can check if the application is running.
*   **AC:**
    *   Given the HTTP server is running (FND-09)
    *   When sending a GET request to `/healthz`
    *   Then the application responds with a 200 OK status code and a simple body (e.g., `{"status": "ok"}`).
*   **Dependencies:** FND-09

---

### 4.2 Epic: Editor Authentication & Management

*   **Story ID:** AUTH-01
*   **Priority:** 11
*   **Story:** As a Developer, I want to set up the Firebase Admin SDK for Go, so that the backend can interact with Firebase services (Auth, Firestore).
*   **AC:**
    *   Given a Firebase project exists with Auth and Firestore enabled
    *   When the application starts
    *   Then the Firebase Admin SDK is initialized successfully using service account credentials (loaded from env var or file path).
*   **Dependencies:** FND-01, FND-07, External Firebase Setup

*   **Story ID:** AUTH-02
*   **Priority:** 12
*   **Story:** As a Developer, I want to implement a JWT validation middleware using the Firebase Admin SDK, so that incoming Firebase ID tokens can be verified on protected routes.
*   **AC:**
    *   Given the Firebase Admin SDK is initialized (AUTH-01) and `chi` router is set up (FND-09)
    *   When a request with a valid `Authorization: Bearer <Firebase ID Token>` header hits a route protected by this middleware
    *   Then the middleware successfully validates the token using Firebase Auth public keys.
    *   And the request is allowed to proceed (e.g., UID added to request context).
    *   When an invalid/expired token is used
    *   Then the middleware rejects the request with a 401 or 403 status code.
*   **Dependencies:** AUTH-01, FND-09

*   **Story ID:** AUTH-03
*   **Priority:** 13
*   **Story:** As a System, I want to store editor metadata (Firebase UID, email) in the PostgreSQL `editors` table upon successful Firebase signup, so that we have a local record associated with the Firebase user. *(Note: Assumes explicit backend call after frontend signup)*
*   **AC:**
    *   Given the `sqlc` code for editors exists (FND-06) and DB connection is available
    *   When a request (e.g., `POST /auth/signup-confirm`) provides a valid Firebase UID and email
    *   Then the system uses the generated `CreateEditor` sqlc function to insert a new record into the `editors` table.
    *   And the operation succeeds without violating constraints (e.g., unique email).
*   **Dependencies:** FND-06, FND-07, AUTH-01 (Potentially for token validation if endpoint is protected)

*   **Story ID:** AUTH-04
*   **Priority:** 14
*   **Story:** As an authenticated Editor, I want to retrieve my own basic details (UID, email) via a `GET /auth/me` endpoint, so that I can confirm my identity.
*   **AC:**
    *   Given the editor is authenticated via JWT middleware (AUTH-02)
    *   And the editor's UID is available (e.g., from request context)
    *   And the editor exists in the `editors` table (AUTH-03)
    *   When sending a GET request to `/auth/me`
    *   Then the system uses the `GetEditor` sqlc function to retrieve the editor's data.
    *   And responds with 200 OK and JSON body containing the editor's `id` (Firebase UID) and `email`.
*   **Dependencies:** AUTH-02, AUTH-03, FND-06

*   **Story ID:** AUTH-05
*   **Priority:** 25 (Lower, part of enhance phase)
*   **Story:** As an authenticated Editor, I want to delete my account data (MVP: Firebase Auth user and local `editors` record), so that my account is removed from the system.
*   **AC:**
    *   Given the editor is authenticated via JWT middleware (AUTH-02) and their UID is known
    *   And the editor exists in Firebase Auth and the `editors` table
    *   When sending a DELETE request to `/auth/me`
    *   Then the system calls the Firebase Admin SDK to delete the user from Firebase Auth.
    *   And the system uses the `DeleteEditor` sqlc function to delete the record from the `editors` table.
    *   And responds with a 200 OK or 204 No Content status code upon success.
*   **Dependencies:** AUTH-02, AUTH-03, FND-06

---

### 4.3 Epic: Newsletter Management

*   **Story ID:** NL-01
*   **Priority:** 15
*   **Story:** As a Developer, I want to define the PostgreSQL schema for the `newsletters` table, so that the structure for storing newsletter data is defined.
*   **AC:**
    *   Given the project structure exists
    *   When viewing the schema file (e.g., `internal/db/schema/002_newsletters.sql`)
    *   Then it contains `CREATE TABLE newsletters` with columns: `id` (UUID, PK), `slug` (VARCHAR, UNIQUE), `name`, `description`, `editor_id` (VARCHAR, FK-like reference), `created_at`, `updated_at`.
    *   And appropriate indexes (`slug`, `editor_id`) are defined.
*   **Dependencies:** FND-01

*   **Story ID:** NL-02
*   **Priority:** 16
*   **Story:** As a Developer, I want to create basic SQL queries (`INSERT`, `SELECT by ID`, `SELECT by EditorID`) for the `newsletters` table, so that `sqlc` can generate corresponding Go functions.
*   **AC:**
    *   Given the `newsletters` schema file exists (NL-01)
    *   When viewing the `internal/db/query/newsletter.sql` file
    *   Then it contains named SQL queries for creating and retrieving newsletters.
*   **Dependencies:** NL-01

*   **Story ID:** NL-03
*   **Priority:** 17
*   **Story:** As a Developer, I want to run `sqlc generate` successfully for the `newsletters` schema and queries, so that type-safe Go code is available.
*   **AC:**
    *   Given `sqlc` is configured and newsletter schema/query files exist (NL-01, NL-02)
    *   When running `sqlc generate`
    *   Then updated/new Go files are generated in `internal/db` without errors and compile.
*   **Dependencies:** FND-06, NL-01, NL-02

*   **Story ID:** NL-04
*   **Priority:** 18
*   **Story:** As an authenticated Editor, I want to create a new newsletter via `POST /api/newsletters`, so that I can start managing content for it.
*   **AC:**
    *   Given the editor is authenticated (AUTH-02) and their UID is known
    *   And the `sqlc` code for newsletters exists (NL-03)
    *   When sending a POST request to `/api/newsletters` with valid JSON body (`{"name": "...", "description": "..."}`)
    *   Then the system validates the input.
    *   And generates a unique URL-friendly slug from the name.
    *   And uses the `CreateNewsletter` sqlc function to insert the record, linking it to the editor's UID.
    *   And responds with 201 Created and JSON body containing the new newsletter's details (ID, slug, name, etc.).
*   **Dependencies:** AUTH-02, NL-03

*   **Story ID:** NL-05
*   **Priority:** 26 (Lower, part of enhance phase)
*   **Story:** As an authenticated Editor, I want to list all newsletters I own via `GET /api/newsletters`, so that I can see my created newsletters.
*   **AC:**
    *   Given the editor is authenticated (AUTH-02) and their UID is known
    *   And the editor owns 0 or more newsletters in the database
    *   When sending a GET request to `/api/newsletters`
    *   Then the system uses the `ListNewslettersByEditor` sqlc function (or similar).
    *   And responds with 200 OK and a JSON array of the editor's newsletters.
*   **Dependencies:** AUTH-02, NL-03, NL-04 (to have data)

---

### 4.4 Epic: Post Management

*   **Story ID:** POST-01
*   **Priority:** 19
*   **Story:** As a Developer, I want to define the PostgreSQL schema for the `posts` table (including `post_status` ENUM), so that the structure for storing post data is defined.
*   **AC:**
    *   Given the project structure exists
    *   When viewing the schema file (e.g., `internal/db/schema/003_posts.sql`)
    *   Then it contains `CREATE TYPE post_status AS ENUM ('draft', 'published', 'failed_to_send')`.
    *   And `CREATE TABLE posts` with columns: `id` (UUID, PK), `newsletter_id` (UUID, FK to newsletters ON DELETE CASCADE), `title`, `content`, `status` (post_status), `published_at` (TIMESTAMPTZ, nullable), `created_at`.
    *   And appropriate index (`newsletter_id`) is defined.
*   **Dependencies:** NL-01 (FK reference)

*   **Story ID:** POST-02
*   **Priority:** 20
*   **Story:** As a Developer, I want to create basic SQL queries (`INSERT`, `SELECT by ID`, `SELECT by NewsletterID`) for the `posts` table, so that `sqlc` can generate corresponding Go functions.
*   **AC:**
    *   Given the `posts` schema file exists (POST-01)
    *   When viewing the `internal/db/query/post.sql` file
    *   Then it contains named SQL queries for creating and retrieving posts.
*   **Dependencies:** POST-01

*   **Story ID:** POST-03
*   **Priority:** 21
*   **Story:** As a Developer, I want to run `sqlc generate` successfully for the `posts` schema and queries, so that type-safe Go code is available.
*   **AC:**
    *   Given `sqlc` is configured and post schema/query files exist (POST-01, POST-02)
    *   When running `sqlc generate`
    *   Then updated/new Go files are generated in `internal/db` without errors and compile.
*   **Dependencies:** FND-06, POST-01, POST-02

*   **Story ID:** POST-04
*   **Priority:** 22
*   **Story:** As an authenticated Editor, I want to publish a new post to a newsletter I own via `POST /api/newsletters/{newsletterId}/posts`, so that the content can be stored and sent to subscribers.
*   **AC:**
    *   Given the editor is authenticated (AUTH-02) and owns newsletter `{newsletterId}`
    *   And the `sqlc` code for posts exists (POST-03)
    *   When sending a POST request to `/api/newsletters/{newsletterId}/posts` with valid JSON body (`{"title": "...", "content": "..."}`)
    *   Then the system validates input and authorization (editor owns newsletter).
    *   And uses the `CreatePost` sqlc function to insert the record with status `'published'` and set `published_at` timestamp.
    *   And responds with 201 Created and JSON body containing the new post's details.
    *   And triggers the email sending process (dependency on EMAIL-02).
*   **Dependencies:** AUTH-02, NL-04, POST-03, EMAIL-02 (for full flow)

---

### 4.5 Epic: Subscriber Management

*   **Story ID:** SUB-01
*   **Priority:** 19 (Can be done in parallel with Post Schema)
*   **Story:** As a Developer, I want to define the conceptual structure for subscriber data in Firebase Firestore (Collection `subscribers`, fields: `email`, `newsletterSlug`, `unsubscribeToken`, `subscribedAt`), so that the data model is clear.
*   **AC:**
    *   Given the Firebase project exists with Firestore enabled
    *   When reviewing project documentation or code comments
    *   Then the structure for the `subscribers` collection and its documents is clearly defined.
*   **Dependencies:** External Firebase Setup

*   **Story ID:** SUB-02
*   **Priority:** 23
*   **Story:** As a User, I want to subscribe to a newsletter via `POST /subscribe` providing my email and the newsletter slug, so that I can be added to the mailing list.
*   **AC:**
    *   Given the Firebase Admin SDK is initialized (AUTH-01)
    *   And a newsletter with the specified `newsletterSlug` exists (can be checked via DB or assumed valid initially)
    *   When sending a POST request to `/subscribe` (public) with valid JSON (`{"email": "...", "newsletterSlug": "..."}`)
    *   Then the system generates a unique, secure `unsubscribeToken`.
    *   And writes a new document to the Firestore `subscribers` collection containing the email, slug, token, and current timestamp.
    *   And responds with 200 OK or 201 Created status code.
*   **Dependencies:** AUTH-01, SUB-01, NL-04 (to have valid slugs)

*   **Story ID:** SUB-03
*   **Priority:** 24
*   **Story:** As a Subscriber, I want to unsubscribe by visiting a unique URL `GET /unsubscribe/{unsubscribeToken}`, so that I am removed from the mailing list.
*   **AC:**
    *   Given the Firebase Admin SDK is initialized (AUTH-01)
    *   And a subscriber document exists in Firestore with the matching `unsubscribeToken`
    *   When sending a GET request to `/unsubscribe/{unsubscribeToken}` (public)
    *   Then the system finds the Firestore document using the token.
    *   And deletes the document (or marks it inactive).
    *   And responds with 200 OK and a simple success message/HTML page.
*   **Dependencies:** AUTH-01, SUB-01, SUB-02 (to have data)

---

### 4.6 Epic: Email Sending

*   **Story ID:** EMAIL-01
*   **Priority:** 20 (Can be done in parallel with Post Schema)
*   **Story:** As a Developer, I want to integrate the Resend Go SDK and configure it with an API key (via env var), so that the application can send emails via Resend.
*   **AC:**
    *   Given a Resend account and API key exist
    *   When the application needs to send an email
    *   Then it initializes the Resend client using the API key from environment variables.
*   **Dependencies:** FND-07, External Resend Setup

*   **Story ID:** EMAIL-02
*   **Priority:** 23 (Needs Post/Sub implemented)
*   **Story:** As a System, I want to fetch subscribers from Firestore and send the published post content via Resend when a post is published, so that subscribers receive the new content.
*   **AC:**
    *   Given a post is successfully created via POST-04 for `newsletterSlug`
    *   And active subscribers for that `newsletterSlug` exist in Firestore (SUB-02)
    *   And the Resend client is configured (EMAIL-01)
    *   When the post creation handler completes saving the post
    *   Then it queries Firestore for subscribers matching the `newsletterSlug`.
    *   And for each subscriber email, it calls the Resend API to send an email containing the post title and content.
    *   And logs success or failure for the email dispatch attempts (basic logging).
*   **Dependencies:** POST-04, SUB-02, EMAIL-01

---

### 4.7 Epic: API Gateway & Deployment

*   **Story ID:** DEPLOY-01
*   **Priority:** 15 (Can be started early)
*   **Story:** As a Developer, I want to create a basic `Dockerfile` for the `api-server`, so that I can build a container image for the application.
*   **AC:**
    *   Given the Go project exists
    *   When running `docker build .` in the `api-server` directory
    *   Then a Docker image is built successfully, containing the compiled Go binary and any necessary assets.
    *   And the image specifies the command to run the binary.
*   **Dependencies:** FND-01

*   **Story ID:** DEPLOY-02
*   **Priority:** 16
*   **Story:** As a Developer, I want to create a basic `Caddyfile` (or Caddy JSON config) to proxy requests to the backend service, so that requests can be routed correctly.
*   **AC:**
    *   Given the backend service will run on a specific port (e.g., 8080)
    *   When Caddy starts using this configuration
    *   Then requests to `/api/*` and `/auth/*` are proxied to the backend service (e.g., `api-server:8080`).
    *   And requests to `/healthz` are proxied.
*   **Dependencies:** FND-10 (for healthz route)

*   **Story ID:** DEPLOY-03
*   **Priority:** 17
*   **Story:** As a Developer, I want to set up a local `docker-compose.yml` file to run the `api-server`, PostgreSQL, and Caddy together, so that I can easily test the full stack locally.
*   **AC:**
    *   Given the Dockerfile (DEPLOY-01), Caddyfile (DEPLOY-02), and local Postgres setup (FND-02) exist
    *   When running `docker-compose up`
    *   Then the `api-server`, PostgreSQL, and Caddy containers start and are networked together.
    *   And requests to Caddy (e.g., `localhost:80/healthz`) are correctly routed and served by the `api-server`.
*   **Dependencies:** FND-02, DEPLOY-01, DEPLOY-02

*   **Story ID:** DEPLOY-04
*   **Priority:** 27 (End of Phase 1/Start of Phase 2)
*   **Story:** As a Developer, I want to deploy the containerized application (api-server + potentially Caddy) to a cloud platform (e.g., Render/Railway), so that the API is publicly accessible.
*   **AC:**
    *   Given a cloud platform account exists and is configured
    *   And the application image can be built (e.g., via buildpack or uploaded Dockerfile)
    *   When the deployment process is triggered
    *   Then the application service(s) start successfully on the cloud platform.
    *   And required environment variables (DB URL, Firebase creds, Resend key) are configured.
    *   And the deployed API's `/healthz` endpoint is accessible via its public URL and returns 200 OK.
*   **Dependencies:** DEPLOY-01, All MVP functional stories (to have something meaningful to deploy)

---

### 4.8 Epic: Documentation & Quality

*   **Story ID:** DOC-01
*   **Priority:** 28 (Phase 2)
*   **Story:** As a Developer, I want to write the basic project README.md including setup instructions and architecture overview, so that others (or future me) can understand and run the project.
*   **AC:**
    *   Given the project exists
    *   When viewing the `README.md` file
    *   Then it contains sections explaining: Project Purpose, Tech Stack, Local Setup Steps (env vars, docker-compose), Basic Architecture Diagram/Description.
*   **Dependencies:** Most setup stories (FND-* , DEPLOY-*)

*   **Story ID:** DOC-02
*   **Priority:** 29 (Phase 2)
*   **Story:** As a Developer, I want to generate an initial OpenAPI/Swagger specification file (e.g., using `swag` annotations or manual creation), so that the API contract is documented.
*   **AC:**
    *   Given API endpoints exist
    *   When running the generation tool or viewing the `openapi.yaml/json` file
    *   Then a valid OpenAPI spec file is present documenting at least the core MVP endpoints (`/healthz`, `/auth/me`, `/api/newsletters` POST, `/api/newsletters/{id}/posts` POST, `/subscribe`, `/unsubscribe`).
*   **Dependencies:** Implementation of corresponding API handlers.

*   **Story ID:** TEST-01
*   **Priority:** 18 (Interleaved with feature development)
*   **Story:** As a Developer, I want to write basic unit tests for a critical utility function or a simple handler logic branch, so that I establish a testing pattern and ensure core logic is verified.
*   **AC:**
    *   Given a non-trivial function exists (e.g., slug generation, input validation)
    *   When running `go test ./...`
    *   Then unit tests covering that function execute and pass.
    *   And test coverage increases for the relevant package.
*   **Dependencies:** Corresponding code implementation.

---

## 5. Dependency Overview

*   **Foundation (FND-\*):** Must be done first, largely sequentially.
*   **Auth Setup (AUTH-01, AUTH-02):** Needed before any protected endpoints.
*   **Schema/sqlc (NL-\*, POST-\*):** Must be done before handlers that use those tables. Run `sqlc generate` after schema/query changes.
*   **API Handlers (NL-04, POST-04, SUB-02, etc.):** Depend on relevant schema/sqlc stories and often auth middleware.
*   **Email Sending (EMAIL-02):** Depends on Post publishing (POST-04) and Resend setup (EMAIL-01).
*   **Deployment (DEPLOY-\*):** Depends on having a runnable application (Dockerfile) and working features.
*   **Documentation/Testing:** Can be interleaved but often depends on the features they document/test.

## 6. High-Level Sprint/Iteration Mapping

Based on the 2-week timeline, we can structure this into two rough iterations:

**Iteration 1: MVP Focus (Target Day 1-7)**

*   **Goal:** Get the core publish/subscribe flow working end-to-end locally and potentially deployed.
*   **Stories:**
    *   All FND-* (Foundation)
    *   AUTH-01, AUTH-02, AUTH-03, AUTH-04 (Core Auth)
    *   NL-01, NL-02, NL-03, NL-04 (Create Newsletter)
    *   POST-01, POST-02, POST-03, POST-04 (Publish Post)
    *   SUB-01, SUB-02, SUB-03 (Subscribe/Unsubscribe)
    *   EMAIL-01, EMAIL-02 (Email Sending)
    *   DEPLOY-01, DEPLOY-02, DEPLOY-03 (Local Stack)
    *   TEST-01 (Start Testing)
    *   *(Stretch)* DEPLOY-04 (Initial Deployment)

**Iteration 2: Enhancements & Polish (Target Day 8-14)**

*   **Goal:** Implement remaining features, improve quality, documentation, and finalize deployment.
*   **Stories:**
    *   AUTH-05 (Delete Account)
    *   NL-05 (List Newsletters) + other NL CRUD (Get/Update/Delete - could be separate stories if needed)
    *   Other POST CRUD (List/Get)
    *   Other SUB features (List Subscribers)
    *   DEPLOY-04 (if not done) + Deployment refinement
    *   DOC-01, DOC-02 (Documentation)
    *   Additional TEST-* stories for coverage
    *   Any remaining "Should Have" or "Could Have" features if time permits (e.g., Password Reset, Rate Limiting).

---

*End of User Story Document*