# Product Requirements Document: Go Newsletter Platform Backend API

**Version:** 1.0
**Status:** Draft

---

## Table of Contents

1.  **Introduction & Overview**
2.  **Goals & Objectives**
3.  **Scope**
    *   3.1 In Scope
    *   3.2 Out of Scope
4.  **User Personas & Goals**
    *   4.1 Editor
    *   4.2 Subscriber
5.  **Feature Breakdown & Prioritization (Kano Model)**
6.  **Functional Requirements**
    *   6.1 Editor Management
    *   6.2 Newsletter Management
    *   6.3 Post Management
    *   6.4 Subscriber Management
    *   6.5 Emailing
    *   6.6 API & System
7.  **Non-Functional Requirements**
    *   7.1 Performance
    *   7.2 Scalability
    *   7.3 Reliability
    *   7.4 Security
    *   7.5 Maintainability
    *   7.6 Usability (API)
    *   7.7 Compliance
8.  **User Workflows & Journeys**
    *   8.1 Editor Registration & Login
    *   8.2 Editor Creates & Publishes Post
    *   8.3 User Subscribes to Newsletter
    *   8.4 User Unsubscribes from Newsletter
    *   8.5 Editor Account Deletion
9.  **Proposed Architecture**
    *   9.1 Conceptual Diagram
    *   9.2 Key Components & Rationale
    *   9.3 Technical Constraints & Dependencies
10. **API Specification**
11. **Acceptance Criteria (Gherkin Syntax)**
    *   11.1 Editor Signup
    *   11.2 Editor Login
    *   11.3 Create Newsletter
    *   11.4 Publish Post
    *   11.5 Subscribe to Newsletter
    *   11.6 Unsubscribe from Newsletter
    *   11.7 Editor Deletion (Simplified MVP)
12. **Release Strategy & Incremental Roadmap**
    *   12.1 Phase 0: Foundation & Setup
    *   12.2 Phase 1: MVP Core Functionality (Target: ~7 Days)
    *   12.3 Phase 2: Enhancements & Refinement (Target: Remaining ~7 Days)
    *   12.4 Deferred / Future Considerations
13. **Open Questions**
14. **RAID Log (Risks, Assumptions, Issues, Dependencies)**

---

## 1. Introduction & Overview

This document defines the requirements for the backend API of the "Go Newsletter Platform," a semester project. The goal is to build a functional backend service using Go, PostgreSQL (without an ORM, leveraging `sqlc`), Firebase Firestore, and Firebase Authentication. The API will serve web and mobile clients (via REST) enabling editors to manage newsletters and publish posts, and users to subscribe and unsubscribe.

This PRD synthesizes information from the "Project Strategy & Implementation Plan" (v1.0), the "Architecture Document" (v1.0), and the "Comprehensive Review" feedback, reflecting the decision to build a **single deployable Go binary (`api-server`) with internal packages (`auth`, `newsletter`)** fronted by a Caddy API gateway, to mitigate risks associated with the tight 2-week timeline.

## 2. Goals & Objectives

*   **Functional:** Deliver a working backend API that fulfills the core requirements of editor management, newsletter/post management, subscriber management, and email publishing.
*   **Technical Learning:** Demonstrate understanding of Go for backend development, REST API design, database interaction without an ORM (`sqlc`), integration with external services (Firebase Auth, Firestore, Resend), and basic microservice concepts (even within a single binary structure).
*   **Project:** Successfully complete the semester project requirements within the strict 2-week deadline.
*   **Quality:** Deliver a well-architected, documented, and testable codebase suitable for a final release, not just a prototype.

## 3. Scope

### 3.1 In Scope

*   **Backend API:** Implementation of RESTful endpoints for all defined features.
*   **Editor Identity:** Registration, login, account deletion via Firebase Authentication integration. Editor metadata storage in PostgreSQL.
*   **Newsletter/Post Management:** CRUD operations for newsletters and posts (owned by editors), stored in PostgreSQL.
*   **Subscriber Management:** Public subscription endpoint, unsubscribe via tokenized link, storage in Firebase Firestore.
*   **Email Publishing:** Sending published posts to confirmed newsletter subscribers via Resend.
*   **Email Confirmation:** Sending a confirmation email upon subscription (TBD if includes explicit confirmation step or just welcome).
*   **API Gateway:** Configuration of Caddy (or similar) for routing, TLS termination (optional, depending on deployment platform).
*   **Database Interaction:** Using `sqlc` for type-safe SQL execution against PostgreSQL.
*   **Authentication:** Stateless JWT (Firebase ID Tokens) validation on protected API endpoints.
*   **Basic Observability:** Structured logging, basic health check endpoint.
*   **Documentation:** Project README, Decision Log, OpenAPI/Swagger specification.
*   **Deployment:** Containerization (Docker) and deployment to a cloud platform (e.g., Render, Railway).

### 3.2 Out of Scope

*   **Frontend Implementation:** No web or mobile client development.
*   **Advanced Email Features:** Email template design/customization, bounce/complaint handling, open/click tracking.
*   **Post Scheduling:** Functionality to schedule posts for future publication.
*   **GraphQL API:** An alternative API style to REST.
*   **Advanced Analytics:** Tracking newsletter/post performance.
*   **Complex Authorization:** Role-based access control beyond simple ownership.
*   **Full Transactional Guarantees Across Services:** Complex distributed transaction management (simplified approach for deletion).
*   **UI/UX Design:** Visual or interaction design aspects.
*   **Specific Cloud Provider Infrastructure:** Detailed IaC (Infrastructure as Code) beyond Docker setup.

## 4. User Personas & Goals

### 4.1 Editor (e.g., Content Creator, Blogger)

*   **Goal:** Easily manage newsletters, write/publish content, and grow an audience.
*   **Needs:** Simple registration/login, intuitive interface (API) for creating/managing newsletters and posts, reliable publishing mechanism, ability to see who subscribed (basic list), control over their account.

### 4.2 Subscriber (e.g., Reader, Follower)

*   **Goal:** Receive interesting content via email from newsletters they follow.
*   **Needs:** Easy way to subscribe with just an email, confirmation of subscription, clear way to unsubscribe from emails, trust that their email won't be misused.

## 5. Feature Breakdown & Prioritization (Kano Model)

Features are classified based on the Kano model relative to the project goals and constraints.

*   **Basic (Must-Haves for MVP):**
    *   Editor Registration (via Firebase Auth)
    *   Editor Login (via Firebase Auth, obtain JWT)
    *   JWT Validation Middleware
    *   Create Newsletter (Authenticated Editor)
    *   Publish Post to Newsletter (Authenticated Editor)
    *   Subscribe to Newsletter (Public Endpoint)
    *   Unsubscribe from Newsletter (via Email Link/Token)
    *   Email Sending on Publish (Basic, via Resend)
    *   PostgreSQL Schema Setup & `sqlc` Integration
    *   Firebase Firestore Setup for Subscribers
    *   Basic API Gateway Routing (Caddy)
    *   Basic Deployment Setup (Docker, Cloud Platform)
    *   Minimal README Documentation
    *   Health Check Endpoint (`/healthz`)

*   **Performance (Expected for Quality):**
    *   List Own Newsletters (Authenticated Editor)
    *   List Posts for Own Newsletter (Authenticated Editor)
    *   Get Editor Details (`/me`)
    *   Reasonable API Response Times (<500ms P95 for reads, <2s for writes)
    *   Timely Email Delivery (within minutes of publish)
    *   Structured Logging
    *   Generated OpenAPI/Swagger Documentation

*   **Excitement (Delighters / Stretch Goals if time permits):**
    *   Update Newsletter Details
    *   Delete Newsletter & Associated Posts
    *   Get Newsletter Details
    *   Get Post Details
    *   List Subscribers for Own Newsletter
    *   Editor Account Deletion (with data cleanup)
    *   Password Reset Flow (via Firebase)
    *   Email Confirmation on Subscribe (if not in MVP)
    *   Basic Rate Limiting on Public Endpoints
    *   Improved Error Handling/Reporting (e.g., email send failures)
    *   CI/CD Pipeline Setup

## 6. Functional Requirements

### 6.1 Editor Management

*   **FR1.1 (Register):** System shall allow a new user to register as an Editor using email and password via Firebase Authentication. Editor metadata (Firebase UID, email) shall be stored in the PostgreSQL `editors` table.
*   **FR1.2 (Login):** System shall allow a registered Editor to log in using email and password via Firebase Authentication, receiving a Firebase ID Token (JWT). (Note: The backend API *validates* tokens, Firebase handles issuance).
*   **FR1.3 (Get Self):** An authenticated Editor shall be able to retrieve their basic profile information (Firebase UID, email) via a `GET /auth/me` endpoint.
*   **FR1.4 (Delete Account - MVP Simplified):** An authenticated Editor shall be able to initiate account deletion via `DELETE /auth/me`. The system *must* delete the user from Firebase Authentication and the PostgreSQL `editors` table. (Stretch Goal: Trigger cleanup of associated Newsletter data).
*   **FR1.5 (Password Reset - Stretch):** System shall support initiating and completing a password reset flow via Firebase Authentication (backend may need endpoints to proxy or handle confirmation depending on Firebase flow).

### 6.2 Newsletter Management

*   **FR2.1 (Create):** An authenticated Editor shall be able to create a new newsletter by providing a name and description (`POST /api/newsletters`). The system shall generate a unique, URL-friendly `slug`, store the newsletter in PostgreSQL associated with the editor's ID.
*   **FR2.2 (List Own):** An authenticated Editor shall be able to list all newsletters they own (`GET /api/newsletters`).
*   **FR2.3 (Get Details - Stretch):** An authenticated Editor shall be able to retrieve details of a specific newsletter they own (`GET /api/newsletters/{newsletterId}`).
*   **FR2.4 (Update - Stretch):** An authenticated Editor shall be able to update the name and description of a newsletter they own (`PATCH /api/newsletters/{newsletterId}`). (Slug updates are out of scope initially due to complexity).
*   **FR2.5 (Delete - Stretch):** An authenticated Editor shall be able to delete a newsletter they own (`DELETE /api/newsletters/{newsletterId}`). Deleting a newsletter *must* also delete all associated posts (via DB cascade).

### 6.3 Post Management

*   **FR3.1 (Publish):** An authenticated Editor shall be able to publish a new post to a specific newsletter they own by providing a title and content (`POST /api/newsletters/{newsletterId}/posts`). The post shall be stored in PostgreSQL with status 'published' and associated with the newsletter. This action triggers email sending (FR5.1).
*   **FR3.2 (List Posts):** An authenticated Editor shall be able to list published posts for a specific newsletter they own (`GET /api/newsletters/{newsletterId}/posts`).
*   **FR3.3 (Get Details - Stretch):** An authenticated Editor shall be able to retrieve the details (title, content) of a specific post belonging to a newsletter they own (`GET /api/posts/{postId}`).

### 6.4 Subscriber Management

*   **FR4.1 (Subscribe):** Any user shall be able to subscribe to a newsletter by providing their email address and the newsletter's `slug` via a public endpoint (`POST /subscribe`). The system shall store the subscription details (email, newsletter slug, unique unsubscribe token) in Firebase Firestore.
*   **FR4.2 (Unsubscribe):** A subscriber shall be able to unsubscribe from a newsletter by accessing a unique link containing an unsubscribe token (`GET /unsubscribe/{unsubscribeToken}`). The system shall use the token to identify and remove/mark the subscription as inactive in Firebase Firestore.
*   **FR4.3 (List Subscribers - Stretch):** An authenticated Editor shall be able to list the emails of subscribers for a newsletter they own (`GET /api/newsletters/{newsletterId}/subscribers`).

### 6.5 Emailing

*   **FR5.1 (Publish Notification):** When a post is successfully published (FR3.1), the system shall retrieve the list of active subscribers for that newsletter from Firestore and send an email containing the post title and content to each subscriber via the Resend service.
*   **FR5.2 (Subscription Confirmation - Stretch/MVP?):** Upon successful subscription (FR4.1), the system may send a confirmation/welcome email to the subscriber via Resend. (Decision needed: Simple welcome or double opt-in link?).

### 6.6 API & System

*   **FR6.1 (Authentication Middleware):** All protected endpoints must validate the incoming Firebase ID Token (JWT) in the `Authorization: Bearer <token>` header. Unauthenticated or invalid token requests must be rejected with appropriate HTTP status codes (401/403).
*   **FR6.2 (API Documentation):** The API shall be documented using the OpenAPI v3 standard (Swagger). This documentation should be accessible, potentially via a `/docs` endpoint served by Caddy.
*   **FR6.3 (Health Check):** The system shall expose a basic health check endpoint (`/healthz` or similar) that returns a 200 OK status if the service is running.
*   **FR6.4 (Configuration):** All external service credentials, database connection strings, ports, and other environment-specific settings must be configurable via Environment Variables.
*   **FR6.5 (Logging):** The application shall produce structured logs (e.g., JSON) for requests, errors, and key events (like email sending attempts/failures).

## 7. Non-Functional Requirements

*   **NFR7.1 (Performance):**
    *   API read endpoints (GET): P95 response time < 500ms under expected load.
    *   API write endpoints (POST, PATCH, DELETE): P95 response time < 2000ms under expected load.
    *   Email delivery: Emails should be dispatched to Resend within seconds of publishing, aiming for delivery to inbox within 5 minutes (acknowledging external dependency).
*   **NFR7.2 (Scalability):**
    *   The `api-server` should be stateless to allow horizontal scaling (multiple instances behind Caddy/load balancer).
    *   Leverage scalability of managed services: Firebase Auth, Firestore, Resend, Cloud Platform hosting.
    *   PostgreSQL can be scaled vertically initially.
*   **NFR7.3 (Reliability):**
    *   Target availability: 99.5% for the core API during the project evaluation period.
    *   Graceful handling of external service failures (e.g., log error if Resend is down, potentially mark post as `failed_to_send`).
    *   Database connections should be resilient (e.g., use connection pooling).
    *   Basic health checks for monitoring.
*   **NFR7.4 (Security):**
    *   All external communication must use HTTPS (handled by Caddy/Platform).
    *   Secure JWT validation using Firebase Admin SDK public keys (handle key rotation).
    *   Prevent SQL Injection (primary mitigation via `sqlc`).
    *   Manage secrets securely using environment variables (no hardcoding).
    *   Run application container as a non-root user.
    *   Consider basic rate limiting on public endpoints (`/subscribe`, `/unsubscribe`) to prevent abuse (Stretch Goal).
    *   Protect internal endpoints if ever exposed (though deletion logic is now internal to the single binary).
*   **NFR7.5 (Maintainability):**
    *   Codebase organized into logical packages (`auth`, `newsletter`, `internal/db`, `internal/email`, etc.) within the single `api-server` module.
    *   Adhere to Go best practices and formatting (`gofmt`).
    *   Use `sqlc` to keep SQL queries separate and generate type-safe Go code.
    *   Clear, structured logging.
    *   Comprehensive README for setup and deployment.
    *   OpenAPI specification for API contract.
    *   Aim for reasonable test coverage (unit/integration tests for critical paths).
*   **NFR7.6 (Usability - API):**
    *   Adhere to RESTful principles (standard HTTP verbs, resource-based URLs, JSON).
    *   Consistent request/response formats.
    *   Clear and consistent error message structure.
*   **NFR7.7 (Compliance):**
    *   Ensure unsubscribe functionality is robust and honors requests promptly (GDPR consideration).
    *   Store minimal necessary user data.

## 8. User Workflows & Journeys

### 8.1 Editor Registration & Login

1.  User navigates to Frontend (OOS).
2.  User provides email/password, Frontend interacts with Firebase Auth SDK for signup/login.
3.  *On Signup:* Firebase creates user, returns success. Backend `POST /auth/signup` called (by frontend or Firebase trigger TBD) to store editor metadata in PostgreSQL.
4.  *On Login:* Firebase returns JWT to Frontend.
5.  Frontend stores JWT and includes it in subsequent `Authorization: Bearer` headers for API calls.

### 8.2 Editor Creates & Publishes Post

1.  Editor (authenticated, has JWT) uses Frontend (OOS).
2.  Frontend calls `POST /api/newsletters` with name/description -> Backend creates newsletter, returns details.
3.  Frontend calls `POST /api/newsletters/{id}/posts` with title/content -> Backend:
    *   Validates JWT & authorization (editor owns newsletter).
    *   Stores post in PostgreSQL (status='published').
    *   Retrieves subscriber emails from Firestore for this newsletter.
    *   Asynchronously (preferred) or synchronously calls Resend API for each subscriber.
    *   Logs success/failure of email dispatch.
    *   Returns post details or confirmation to Frontend.

### 8.3 User Subscribes to Newsletter

1.  User visits public page/form (OOS).
2.  User enters email, selects newsletter (slug identified).
3.  Frontend calls public `POST /subscribe` with email and newsletterSlug -> Backend:
    *   Validates input.
    *   Generates unique unsubscribe token.
    *   Stores email, slug, token in Firestore.
    *   (Optional/Stretch) Triggers confirmation/welcome email via Resend.
    *   Returns success to Frontend.

### 8.4 User Unsubscribes from Newsletter

1.  User clicks unsubscribe link in email (e.g., `https://yourdomain.com/unsubscribe/{token}`).
2.  Browser hits `GET /unsubscribe/{token}` -> Backend:
    *   Extracts token.
    *   Finds subscriber record in Firestore matching the token.
    *   Removes record or marks as unsubscribed.
    *   Returns a success confirmation page/message (simple HTML or redirect).

### 8.5 Editor Account Deletion (Simplified MVP Workflow)

1.  Editor (authenticated) initiates deletion via Frontend (OOS).
2.  Frontend calls `DELETE /auth/me` -> Backend:
    *   Validates JWT.
    *   Extracts editor's Firebase UID.
    *   Uses Firebase Admin SDK to delete user from Firebase Auth.
    *   Executes `DELETE FROM editors WHERE id = $1` in PostgreSQL.
    *   Returns success/failure to Frontend.
    *   *(Deferred/Stretch: Implement robust deletion of newsletters/posts, potentially via background job or explicit user confirmation of data loss)*

## 9. Proposed Architecture

Reflecting the feedback and timeline constraints, the architecture is simplified to a **single Go binary (`api-server`)** deployable as a container, fronted by Caddy.

### 9.1 Conceptual Diagram (Mermaid)

```mermaid
graph LR
    subgraph "Client Domain"
        Client[Web/Mobile Client]
    end

    subgraph "Cloud Platform / Network Edge"
        Caddy[Caddy API Gateway <br/> (TLS, Routing)]
    end

    subgraph "Backend Application (Single Container)"
        Server[api-server (Go Binary)]
        PkgAuth[auth package <br/> (Handlers, Logic)]
        PkgNL[newsletter package <br/> (Handlers, Logic)]
        PkgDB[internal/db <br/> (sqlc generated)]
        PkgEmail[internal/email <br/> (Resend Client)]
        PkgFBAuth[internal/auth <br/> (Firebase Admin SDK)]
    end

    subgraph "Data Stores"
        PG[PostgreSQL <br/> (Editors, Newsletters, Posts)]
        FS[Firebase Firestore <br/> (Subscribers)]
    end

    subgraph "External Services"
        FirebaseAuth[Firebase Authentication]
        Resend[Resend API]
    end

    Client -- HTTPS REST --> Caddy
    Caddy -- HTTP /docs --> Server(Static Files)
    Caddy -- HTTP /api/*, /auth/* --> Server

    Server -- Uses --> PkgAuth
    Server -- Uses --> PkgNL
    Server -- Uses --> PkgDB
    Server -- Uses --> PkgEmail
    Server -- Uses --> PkgFBAuth

    PkgAuth -- Validates Token With --> FirebaseAuth
    PkgAuth -- Stores/Deletes --> PG(editors table)
    PkgAuth -- Deletes User --> FirebaseAuth

    PkgNL -- CRUD --> PG(newsletters, posts tables)
    PkgNL -- CRUD --> FS(subscribers collection)
    PkgNL -- Validates Token With --> FirebaseAuth
    PkgNL -- Sends Via --> Resend
    PkgNL -- Uses --> PkgDB
    PkgNL -- Uses --> PkgEmail

    PkgDB -- SQL (via pgx) --> PG

    PkgEmail -- API Call --> Resend

    PkgFBAuth -- SDK Calls --> FirebaseAuth

    Client -- Interacts With --> FirebaseAuth (for Login/Signup Flow)

```

### 9.2 Key Components & Rationale

*   **`api-server` (Go Binary):** Single deployment unit containing all backend logic.
    *   *Rationale:* Significantly reduces deployment complexity and operational overhead compared to multiple microservices, critical for the 2-week timeline. Internal Go packages (`auth`, `newsletter`) maintain logical separation, allowing potential future splitting if needed.
*   **Caddy:** API Gateway.
    *   *Rationale:* Simple configuration, handles HTTPS automatically, provides routing, hides internal structure. Standard practice.
*   **PostgreSQL + `sqlc`:** Database for relational data.
    *   *Rationale:* Meets "PostgreSQL" and "no ORM" requirements. `sqlc` provides type safety and reduces boilerplate over raw `database/sql`.
*   **Firebase Firestore:** Database for subscriber data.
    *   *Rationale:* Meets "Firebase for subscribers" requirement. Scalable NoSQL store suitable for potentially large, less relational subscriber lists.
*   **Firebase Authentication:** External identity provider.
    *   *Rationale:* Meets "stateless JWT auth" requirement and allowance for 3rd party. Drastically reduces development time and security risks compared to custom auth. Backend validates tokens using Admin SDK.
*   **Resend:** Email sending service.
    *   *Rationale:* Provides reliable email delivery via API, has a Go SDK, suitable free tier.

### 9.3 Technical Constraints & Dependencies

*   **Language:** Go (latest stable)
*   **Database:** PostgreSQL (v13+), Firebase Firestore
*   **DB Interaction:** `sqlc` + `pgx` driver (No ORM)
*   **Authentication:** Firebase Authentication JWTs
*   **API Style:** RESTful JSON
*   **Timeline:** Strict 2 weeks.
*   **Dependencies:** Go compiler, Docker, Caddy, Access to Firebase Project (Auth, Firestore enabled), Resend API Key, Cloud Platform Account (Render/Railway recommended).

## 10. API Specification

A detailed OpenAPI v3 (Swagger) specification (`openapi.yaml` or `openapi.json`) will be created and maintained, documenting all public endpoints defined in the Functional Requirements (Section 6). It will include paths, methods, parameters, request/response bodies, status codes, and authentication requirements. This spec will be the source of truth for API consumers. It should be generated from code comments (e.g., using `swag`) or written manually and potentially served via Caddy at a `/docs` endpoint.

*(Placeholder for link to final OpenAPI spec)*

## 11. Acceptance Criteria (Gherkin Syntax)

*(Key examples provided; more needed for full coverage)*

### 11.1 Editor Signup

```gherkin
Feature: Editor Authentication

  Scenario: New user registers successfully
    Given the user does not exist in Firebase Auth or the editors table
    When the user signs up via Firebase Authentication with email "new.editor@example.com" and password "password123"
    And the backend receives a request (e.g., triggered hook or explicit call) associated with the new Firebase UID
    Then the system should store the editor's Firebase UID and "new.editor@example.com" in the PostgreSQL editors table
    And the overall signup process should succeed
```

### 11.2 Editor Login

```gherkin
Feature: Editor Authentication

  Scenario: Registered editor logs in and backend validates token
    Given the editor "editor@example.com" exists and is registered
    When the editor logs in via Firebase Authentication and obtains a valid JWT
    And the editor sends a GET request to the protected "/auth/me" endpoint with the JWT in the Authorization header
    Then the system should validate the JWT successfully
    And the system should respond with a 200 OK status code
    And the response body should contain the editor's ID (Firebase UID) and email "editor@example.com"
```

### 11.3 Create Newsletter

```gherkin
Feature: Newsletter Management

  Scenario: Authenticated editor creates a new newsletter successfully
    Given the editor is authenticated with a valid JWT
    When the editor sends a POST request to "/api/newsletters" with JSON body: {"name": "My Awesome Newsletter", "description": "News about awesome things"}
    Then the system should respond with a 201 Created status code
    And the response body should contain the new newsletter's details including a unique ID, the name "My Awesome Newsletter", the description, and a generated slug like "my-awesome-newsletter"
    And a corresponding record should exist in the newsletters table in PostgreSQL linked to the editor's ID
```

### 11.4 Publish Post

```gherkin
Feature: Post Management

  Scenario: Authenticated editor publishes a post successfully
    Given the editor is authenticated with a valid JWT and owns newsletter with ID "newsletter-uuid-123"
    And newsletter "newsletter-uuid-123" has subscribers "sub1@example.com", "sub2@example.com" in Firestore
    When the editor sends a POST request to "/api/newsletters/newsletter-uuid-123/posts" with JSON body: {"title": "Big News!", "content": "Something important happened."}
    Then the system should respond with a 201 Created status code
    And the response body should contain the new post's details including ID, title, content, and status "published"
    And a corresponding record should exist in the posts table in PostgreSQL linked to newsletter "newsletter-uuid-123"
    And the system should attempt to send an email containing "Big News!" and "Something important happened." to "sub1@example.com" and "sub2@example.com" via Resend
```

### 11.5 Subscribe to Newsletter

```gherkin
Feature: Subscriber Management

  Scenario: User subscribes to a public newsletter
    Given newsletter with slug "public-news" exists
    When a POST request is sent to "/subscribe" with JSON body: {"email": "new.subscriber@email.com", "newsletterSlug": "public-news"}
    Then the system should respond with a 200 OK or 201 Created status code
    And a record should exist in the Firebase Firestore 'subscribers' collection containing "new.subscriber@email.com", "public-news", and a unique unsubscribe token
    # Optional: And an email should be queued/sent to "new.subscriber@email.com"
```

### 11.6 Unsubscribe from Newsletter

```gherkin
Feature: Subscriber Management

  Scenario: User unsubscribes using a valid token
    Given a subscriber exists in Firestore with email "sub.to.remove@email.com" for newsletter "public-news" and unsubscribe token "valid-unsubscribe-token-abc"
    When a GET request is sent to "/unsubscribe/valid-unsubscribe-token-abc"
    Then the system should find the Firestore record by the token
    And the system should remove the record or mark it as inactive
    And the system should respond with a success indication (e.g., 200 OK with HTML page, or redirect)
```

### 11.7 Editor Deletion (Simplified MVP)

```gherkin
Feature: Editor Management

  Scenario: Authenticated editor deletes their own account (MVP)
    Given the editor "editor.to.delete@example.com" with Firebase UID "firebase-uid-to-delete" is authenticated with a valid JWT
    And a record exists in the PostgreSQL 'editors' table for "firebase-uid-to-delete"
    When the editor sends a DELETE request to "/auth/me"
    Then the system should validate the JWT
    And the system should delete the user "firebase-uid-to-delete" from Firebase Authentication
    And the system should delete the record for "firebase-uid-to-delete" from the PostgreSQL 'editors' table
    And the system should respond with a 200 OK or 204 No Content status code
```

## 12. Release Strategy & Incremental Roadmap

The project will follow an incremental approach focusing on delivering a core MVP within the first week, allowing the second week for enhancements, testing, documentation, and deployment polish.

### 12.1 Phase 0: Foundation & Setup (Days 1-2)

*   Git repository setup, project structure (`api-server` module).
*   Go module initialization.
*   Firebase project setup (Auth, Firestore), obtain credentials.
*   PostgreSQL setup (local Docker), initial schema (`schema.sql`).
*   `sqlc` setup (`sqlc.yaml`, generate initial code).
*   Basic `chi` router setup in `main.go`.
*   Environment variable configuration loading.
*   Basic structured logging implementation.
*   Initial Dockerfile for `api-server`.
*   Local Caddyfile configuration.

### 12.2 Phase 1: MVP Core Functionality (Target: ~7 Days Total)

*   **Auth:** Implement Firebase Auth signup/login flow integration (token validation middleware, `/auth/me`). Store editor metadata in PG.
*   **Newsletter:** Implement `POST /api/newsletters` (Create).
*   **Post:** Implement `POST /api/newsletters/{id}/posts` (Publish). Store post in PG.
*   **Subscription:** Implement `POST /subscribe` (store in Firestore), `GET /unsubscribe/{token}` (remove from Firestore).
*   **Emailing:** Basic Resend integration for sending email on post publish (FR5.1).
*   **System:** Implement `/healthz` endpoint.
*   **Deployment:** Initial deployment to Render/Railway via Docker.
*   **Testing:** Basic unit/integration tests for core flows.

### 12.3 Phase 2: Enhancements & Refinement (Target: Remaining ~7 Days)

*   **CRUD:** Implement remaining Newsletter/Post GET/LIST/PATCH/DELETE endpoints (Stretch features from Sec 5).
*   **Auth:** Implement simplified editor deletion (FR1.4 MVP). Implement Password Reset (Stretch).
*   **Subscribers:** Implement `GET /api/newsletters/{id}/subscribers` (List for Editor - Stretch).
*   **Emailing:** Consider welcome/confirmation email (FR5.2 - Stretch). Improve error handling for email sending (mark post `failed_to_send`).
*   **Documentation:** Generate/Finalize OpenAPI spec. Write comprehensive README.md. Finalize Decision Log.
*   **Testing:** Increase test coverage. Add basic load testing checks (`hey`/`k6`).
*   **Observability:** Enhance logging, potentially add basic metrics middleware.
*   **Security:** Implement basic rate limiting (Stretch). Ensure container runs non-root.
*   **Refinement:** Code cleanup, dependency updates, address review feedback. Final deployment testing.

### 12.4 Deferred / Future Considerations

*   Post Scheduling
*   GraphQL API
*   Advanced Email Features (Templates, Bounce Handling)
*   Robust Data Consistency for Deletion (e.g., background jobs)
*   Full CI/CD Pipeline Automation
*   User Roles / Permissions
*   Analytics / Reporting

## 13. Open Questions

*   Subscription Confirmation: Is a simple welcome email sufficient for MVP, or is a double opt-in confirmation link required? (Assume simple welcome for MVP unless specified otherwise).
*   Editor Signup Backend Trigger: Does Firebase Auth provide a trigger/webhook on user creation that the backend can listen to, or should the frontend explicitly call a backend endpoint like `/auth/signup-confirm` after successful Firebase signup to store editor metadata? (Assume explicit call for now).
*   Error Handling Specifics: Define standard structure for API error responses. How should partial failures (e.g., some emails fail to send) be reported to the editor? (Log failures, mark post status; API returns success if post stored).
*   Rate Limiting Strategy: If implemented, what are the basic limits per IP for public endpoints? (Defer detailed strategy).

## 14. RAID Log (Risks, Assumptions, Issues, Dependencies)

| Category     | Item                                                                | Impact | Likelihood | Mitigation / Action                                                                                                | Owner       | Status      |
| :----------- | :------------------------------------------------------------------ | :----- | :--------- | :----------------------------------------------------------------------------------------------------------------- | :---------- | :---------- |
| **Risk**     | **R1: Tight Timeline (2 weeks)**                                    | High   | High       | Simplify architecture (single binary), prioritize ruthlessly (MVP), timebox features, rely on managed services.     | ProjectLead | Mitigated   |
| **Risk**     | **R2: Complexity of "Distributed" Concepts (even within 1 binary)** | Med    | Med        | Keep internal calls simple (avoid for MVP deletion), focus on clear package boundaries, use standard libraries.    | Dev Team    | Monitoring  |
| **Risk**     | **R3: External Service Limits/Failures (Firebase, Resend)**         | Med    | Low        | Use free tiers judiciously, implement basic error handling/logging for external calls, have fallback (log-only email). | Dev Team    | Planned     |
| **Risk**     | **R4: Deployment / Infrastructure Issues**                          | Med    | Med        | Use familiar PaaS (Render/Railway), containerize early (Docker), test deployment frequently.                     | Dev Team    | Planned     |
| **Risk**     | **R5: `sqlc` Learning Curve / Verbosity**                           | Low    | Med        | Allocate time for learning `sqlc` setup/syntax, create helper functions for mapping DTOs.                        | Dev Team    | Planned     |
| **Assumption** | **A1: Firebase/Resend Free Tiers Sufficient**                       | Med    | High       | Verify limits upfront. Minimal usage expected for development/demo.                                              | ProjectLead | Verified    |
| **Assumption** | **A2: Cloud Platform Compatibility (Render/Railway)**               | Low    | High       | Platforms are Go/Docker friendly. Precedent exists.                                                                | ProjectLead | Assumed     |
| **Assumption** | **A3: `sqlc` meets "no ORM" constraint**                          | Low    | High       | `sqlc` generates code from SQL, doesn't impose its own query abstractions like a typical ORM.                    | ProjectLead | Assumed     |
| **Assumption** | **A4: Basic Go/SQL/Docker/REST knowledge exists**                 | Low    | High       | Core premise of the project.                                                                                       | ProjectLead | Assumed     |
| **Issue**    | **I1: Define specific error response format**                       | Low    | N/A        | Agree on a standard JSON error structure (e.g., `{ "error": { "code": "...", "message": "..." } }`).              | Dev Team    | Open        |
| **Issue**    | **I2: Finalize subscription confirmation flow (welcome vs opt-in)** | Low    | N/A        | Decision: Simple welcome email for MVP due to time constraints.                                                    | ProjectLead | **Resolved** |
| **Issue**    | **I3: Confirm backend trigger mechanism for editor signup**         | Low    | N/A        | Decision: Frontend makes explicit call to backend after Firebase signup success.                                       | ProjectLead | **Resolved** |
| **Dependency** | **D1: Access to Firebase Project**                                | High   | N/A        | Requires setup and credential sharing.                                                                             | ProjectLead | Required    |
| **Dependency** | **D2: Access to Resend Account/API Key**                          | High   | N/A        | Requires signup and key management.                                                                                | ProjectLead | Required    |
| **Dependency** | **D3: Access to Cloud Platform Account**                          | High   | N/A        | Requires account for deployment.                                                                                   | ProjectLead | Required    |
| **Dependency** | **D4: Stability of external tools (`sqlc`, `chi`, etc.)**         | Low    | N/A        | Use stable versions. Monitor community channels if issues arise.                                                   | Dev Team    | Monitoring  |

---

*End of PRD*
```