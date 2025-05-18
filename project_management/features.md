# Features List: Go Newsletter Platform

## Table of Contents
1. Summary
2. Feature List
    - Editor Features
    - Subscriber Features
    - Publishing Features
    - Non-Functional Requirements
3. MoSCoW Prioritization Key

---

## 1. Summary

This document extracts and organizes all features from the PRD for the Go Newsletter Platform. Features are categorized, uniquely identified, prioritized (MoSCoW), and include detailed descriptions, acceptance criteria, technical notes, and estimated complexity with rationale.

---

## 2. Feature List

### 2.1. Editor Features

#### FE-ED-01: Editor Registration
- **Priority:** Must
- **Description:**
  - Allow new editors to register using email and password. Persist editor data in PostgreSQL. Enforce unique email constraint.
- **Acceptance Criteria:**
  - Editor can register with valid email and password.
  - Duplicate emails are rejected.
  - Editor data is stored in PostgreSQL.
  - Returns success or error response.
- **Technical Notes:**
  - Endpoint: `POST /editors/register`
  - Data model: `editors (id, email, password_hash, created_at, ...)`
  - Input validation for email format and password strength.
- **Complexity:** Low (standard registration flow)

#### FE-ED-02: Editor Login
- **Priority:** Must
- **Description:**
  - Allow editors to log in with email and password. Issue JWT on successful authentication.
- **Acceptance Criteria:**
  - Editor can log in with correct credentials.
  - JWT is returned on success.
  - Invalid credentials return error.
- **Technical Notes:**
  - Endpoint: `POST /editors/login`
  - Passwords checked against hash in DB.
  - JWT secret/config managed securely.
- **Complexity:** Low (standard login flow)

#### FE-ED-03: Stateless Auth (JWT)
- **Priority:** Must
- **Description:**
  - Use JWT for stateless authentication for all protected endpoints. Validate JWT on each request.
- **Acceptance Criteria:**
  - JWT is required for protected endpoints.
  - Invalid/expired JWT returns 401 error.
- **Technical Notes:**
  - Middleware for JWT validation.
  - JWT includes editor ID and expiry.
- **Complexity:** Low (standard JWT middleware)

#### FE-ED-04: Password Reset
- **Priority:** Should
- **Description:**
  - Allow editors to request a password reset via email. Provide a secure reset link and allow password update.
- **Acceptance Criteria:**
  - Editor can request reset with email.
  - Receives email with reset link.
  - Can set new password via link.
  - Token is single-use and expires.
- **Technical Notes:**
  - Endpoints: `POST /editors/password-reset-request`, `POST /editors/password-reset`
  - Secure token generation and validation.
  - Email integration required.
- **Complexity:** Medium (email integration, token security)

#### FE-ED-05: Social Auth (Optional)
- **Priority:** Could
- **Description:**
  - Allow editors to register/login using Google or GitHub via Firebase Auth.
- **Acceptance Criteria:**
  - Editor can authenticate via supported social providers.
  - JWT issued on success.
- **Technical Notes:**
  - Integration with Firebase Auth social providers.
  - UI/UX for OAuth flow (if client exists).
- **Complexity:** Medium (third-party integration)

#### FE-ED-06: Create Newsletter
- **Priority:** Must
- **Description:**
  - Authenticated editors can create newsletters. Name is required, description is optional. Each newsletter is owned by a single editor.
- **Acceptance Criteria:**
  - Newsletter created with valid name.
  - Newsletter is linked to editor.
  - Name uniqueness enforced per editor.
- **Technical Notes:**
  - Endpoint: `POST /newsletters`
  - Data model: `newsletters (id, editor_id, name, description, created_at, ...)`
- **Complexity:** Low (simple create flow)

#### FE-ED-07: Rename Newsletter
- **Priority:** Should
- **Description:**
  - Editors can update the name and description of their newsletters.
- **Acceptance Criteria:**
  - Editor can update name/description.
  - Changes are persisted.
  - Only owner can update.
- **Technical Notes:**
  - Endpoint: `PATCH /newsletters/{id}`
  - Auth required; check ownership.
- **Complexity:** Low (simple update)

#### FE-ED-08: Delete Newsletter
- **Priority:** Must
- **Description:**
  - Editors can delete their newsletters. Associated posts and subscriber links must be handled (deleted or archived).
- **Acceptance Criteria:**
  - Editor can delete owned newsletter.
  - Associated data is handled per design.
  - Only owner can delete.
- **Technical Notes:**
  - Endpoint: `DELETE /newsletters/{id}`
  - Cascade or soft-delete related data.
- **Complexity:** Medium (data integrity, cascading)

#### FE-ED-09: List Subscribers
- **Priority:** Should
- **Description:**
  - Editors can view a list of subscribers for their newsletters.
- **Acceptance Criteria:**
  - Editor can retrieve subscriber list for owned newsletters.
  - Data is accurate and up-to-date.
- **Technical Notes:**
  - Endpoint: `GET /newsletters/{id}/subscribers`
  - Data fetched from Firebase.
- **Complexity:** Low (read-only, cross-service fetch)

---

### 2.2. Subscriber Features

#### FE-SB-01: Subscribe to Newsletter
- **Priority:** Must
- **Description:**
  - Users can subscribe to a newsletter by providing their email via a unique link. Subscription is stored in Firebase.
- **Acceptance Criteria:**
  - User can subscribe with valid email.
  - Subscription is persisted in Firebase.
  - Duplicate subscriptions are prevented.
- **Technical Notes:**
  - Endpoint: `POST /newsletters/{id}/subscribe`
  - Unique link generation per newsletter.
- **Complexity:** Low (simple create, external storage)

#### FE-SB-02: Subscription Confirmation
- **Priority:** Must
- **Description:**
  - Upon subscribing, user receives a confirmation email with an unsubscribe link.
- **Acceptance Criteria:**
  - Confirmation email sent to subscriber.
  - Email contains working unsubscribe link.
- **Technical Notes:**
  - Email service integration required.
  - Email template includes unsubscribe URL.
- **Complexity:** Low (email send)

#### FE-SB-03: Unsubscribe from Newsletter
- **Priority:** Must
- **Description:**
  - Subscribers can unsubscribe from a newsletter via a link in any email.
- **Acceptance Criteria:**
  - Unsubscribe link works and removes subscriber from newsletter.
  - Subscriber receives confirmation of unsubscription.
- **Technical Notes:**
  - Endpoint: `POST /newsletters/{id}/unsubscribe`
  - Tokenized link for security.
- **Complexity:** Low (simple delete)

---

### 2.3. Publishing Features

#### FE-PB-01: Publish Post to Newsletter
- **Priority:** Must
- **Description:**
  - Editors can publish posts to their newsletters. Posts are sent to all current subscribers and archived.
- **Acceptance Criteria:**
  - Editor can create and publish a post.
  - Post is sent to all subscribers.
  - Post is stored in DB for archival.
- **Technical Notes:**
  - Endpoint: `POST /newsletters/{id}/posts`
  - Data model: `posts (id, newsletter_id, title, body, created_at, ...)`
  - Email service integration for delivery.
- **Complexity:** Medium (multi-step, email + DB)

#### FE-PB-02: Email Delivery of Posts
- **Priority:** Must
- **Description:**
  - Each published post is emailed to all subscribers of the newsletter.
- **Acceptance Criteria:**
  - All subscribers receive the post via email.
  - Email delivery failures are logged/handled.
- **Technical Notes:**
  - Integration with chosen email service (Resend, SendGrid, AWS SES).
  - Batch or async sending for scale.
- **Complexity:** Medium (external service, error handling)

#### FE-PB-03: Archive Published Posts
- **Priority:** Must
- **Description:**
  - All published posts are stored in the database for future reference.
- **Acceptance Criteria:**
  - Posts are persisted and retrievable by editor/newsletter.
- **Technical Notes:**
  - Data model: `posts` table in PostgreSQL.
  - Endpoint: `GET /newsletters/{id}/posts`
- **Complexity:** Low (DB insert/read)

---

### 2.4. Non-Functional Requirements

#### NFR-01: Production-Ready Quality
- **Priority:** Must
- **Description:**
  - The system must be robust, reliable, and suitable for real-world use, not a prototype.
- **Acceptance Criteria:**
  - Meets reliability, robustness, and documentation goals.
  - No critical bugs or data loss in normal operation.
- **Technical Notes:**
  - Automated tests, error handling, monitoring.
- **Complexity:** High (system-wide)

#### NFR-02: Modern Stack & Architecture
- **Priority:** Must
- **Description:**
  - Use modern Go packages, idiomatic code, and best practices for maintainability and performance.
- **Acceptance Criteria:**
  - Codebase uses up-to-date, idiomatic Go.
  - Follows project and industry standards.
- **Technical Notes:**
  - Dependency management, modular design, code reviews.
- **Complexity:** Medium (ongoing)

#### NFR-03: API Documentation
- **Priority:** Must
- **Description:**
  - Provide comprehensive API documentation for client developers and maintainers.
- **Acceptance Criteria:**
  - All endpoints documented with request/response examples.
  - Docs are accessible and up-to-date.
- **Technical Notes:**
  - Use Swagger/OpenAPI or similar.
- **Complexity:** Medium (docs tooling)

#### NFR-04: Project Documentation
- **Priority:** Must
- **Description:**
  - Provide sufficient project documentation for hand-over and maintenance.
- **Acceptance Criteria:**
  - Setup, architecture, and maintenance docs are complete.
- **Technical Notes:**
  - README, architecture diagrams, setup scripts.
- **Complexity:** Medium (docs creation)

#### NFR-05: Transactional Context
- **Priority:** Should
- **Description:**
  - Use database transactions for multi-step or critical operations to ensure consistency.
- **Acceptance Criteria:**
  - Multi-step DB ops are atomic and consistent.
- **Technical Notes:**
  - Use SQL transactions in Go code.
- **Complexity:** Medium (transactional logic)

#### NFR-06: Naming Convention
- **Priority:** Must
- **Description:**
  - Use required naming for Firebase/Cloud resources: `strv-vse-go-newsletter-[last_name]-[first_name]`.
- **Acceptance Criteria:**
  - All accounts/projects follow naming convention.
- **Technical Notes:**
  - Naming enforced in setup/deployment scripts.
- **Complexity:** Low (convention)

#### NFR-07: GitHub Repository
- **Priority:** Must
- **Description:**
  - Source code is hosted in GitHub, with access for Marek Cermak (CermakM).
- **Acceptance Criteria:**
  - Repo exists, CermakM invited.
- **Technical Notes:**
  - GitHub permissions managed.
- **Complexity:** Low (repo setup)

#### NFR-08: Deployed API URL
- **Priority:** Must
- **Description:**
  - Provide a deployed API URL for client access and testing.
- **Acceptance Criteria:**
  - API is deployed and URL is shared.
- **Technical Notes:**
  - Deployment to Railway, URL in docs.
- **Complexity:** Low (deployment)

---

## 3. MoSCoW Prioritization Key
- **Must:** Essential for MVP and project success
- **Should:** Important but not strictly required for MVP
- **Could:** Nice to have, optional
- **Won't:** Out of scope 