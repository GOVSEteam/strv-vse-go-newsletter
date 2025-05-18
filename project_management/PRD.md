# Product Requirements Document: Go Newsletter Platform

**Version:** 1.0  
**Date:** May 17, 2025  
**Project:** Go Newsletter Platform (Semestral Project)

---

## 1. Overview

The Go Newsletter Platform is an API that enables registered users (Editors) to create, manage, and publish newsletters. Other users (Subscribers) can discover and subscribe to these newsletters. This project is a semestral assignment for the "Microservice Development W/ Go" course.

---

## 2. Goals & Objectives

- Design and implement a robust API for a newsletter platform.
- Allow Editors to manage newsletters and publish content.
- Enable Subscribers to subscribe/unsubscribe to newsletters.
- Demonstrate proficiency in Go, API design (REST), PostgreSQL (no ORMs), and cloud deployment.
- Produce a production-quality application with comprehensive documentation.

---

## 3. Scope

### In Scope
- API for newsletter management (creation, publishing, subscription, unsubscription).
- Editor registration, authentication (JWT), and password reset.
- Subscriber management (subscribe/unsubscribe via email link).
- Email delivery of published posts to subscribers.
- Data storage: Editors & Newsletters in PostgreSQL, Subscribers in Firebase.
- Deployment to Railway cloud platform.
- API designed for both mobile and web clients.

### Out of Scope
- Client-side (frontend) implementation.
- Payment processing, analytics, advanced scheduling.
- Use of ORMs with PostgreSQL.
- Public newsletter discovery/listing (unless specified later).
- Social authentication (optional only).

---

## 4. User Personas / Target Audience

### Editors
- **Goals:** Register, create/manage newsletters, publish posts, view subscribers.
- **Tech-savviness:** Comfortable with web interfaces for content creation/management.
- **Pain Points:** Need a simple way to reach an audience without handling email delivery or subscription management.

### Subscribers
- **Goals:** Subscribe to newsletters, receive content, easily unsubscribe.
- **Tech-savviness:** Able to use email and click links.
- **Pain Points:** Avoid unwanted emails, easy unsubscription, not missing desired content.

---

## 5. Functional Requirements

### 5.1. Editor Features
- Sign up (email & password) & sign in
- Stateless authorization using JWT (Firebase Auth)
- Password reset request
- Create, rename, delete newsletter (single owner, name required, description optional)
- Publish post to newsletter
- List subscribers of their newsletter
- Store editor accounts in PostgreSQL

### 5.2. Subscriber Features
- Subscribe to newsletter via unique link & email
- Receive confirmation email (with unsubscribe link)
- Unsubscribe from newsletter (via link in every email)
- Store subscriber info in Firebase

### 5.3. Publishing
- Send published messages to subscribers via mailing service (Resend, SendGrid, AWS SES)
- Store published messages in database

---

## 6. Non-Functional Requirements

- Production-ready quality (not a prototype)
- Use of modern Go packages, technologies, and architectures
- Good API documentation
- Sufficient project documentation for client hand-over and maintenance
- Transactional context for consistency and robustness
- Naming convention: strv-vse-go-newsletter-[last_name]-[first_name] for Firebase & Cloud
- Source code in GitHub, access for Marek Cermak (CermakM)
- Deployed API URL provided

---

## 7. User Journeys

### Editor Journeys
- **Registration & First Newsletter:** Sign up → Log in → Create newsletter (name, description)
- **Publish Post:** Log in → Select newsletter → Compose post → Publish → System emails subscribers & archives post
- **Manage Newsletters/Subscribers:** Log in → View newsletters → Rename/edit/delete → View subscribers
- **Password Reset:** Forgot password → Request reset → Receive email → Set new password

### Subscriber Journeys
- **Subscribe:** Find unique link → Enter email → Receive confirmation email
- **Receive Content:** Subscribed → Receive published posts via email
- **Unsubscribe:** Click unsubscribe link in email → Stop receiving emails

---

## 8. Success Metrics

- Number of newsletters created
- Number of active subscribers
- Email delivery success rate
- System uptime and reliability
- API response time and error rate
- Completion and clarity of documentation

---

## 9. Timeline (High-Level)

- Project start: [TBD]
- Core API implementation: [TBD]
- Testing & documentation: [TBD]
- Deployment: [TBD]
- Submission: End of semester (per course guidelines)

---

## 10. Open Questions / Assumptions

- **Newsletter Discovery:** How do subscribers find unique links? (Public listing not in scope unless specified)
- **Post Content:** What fields are required? (Assume title, body; HTML optional)
- **Scheduling:** Only immediate publishing is in scope; advanced scheduling is out.
- **Double Opt-In:** Not required unless specified; confirmation email is sufficient.
- **Error Handling:** Standard REST error responses; transactional context for critical ops.
- **Security:** JWT auth; further rate limiting or input validation as best practice.
- **Firebase for Subscribers:** Only for storing email/newsletter subscriptions, not auth.
- **Unique Link Generation:** System generates unique subscription links.
- **Single Editor Ownership:** Each newsletter is owned/managed by a single editor.
- **Team Discretion:** Team may make reasonable choices for unspecified details and document them.

---

## 11. Deliverables

- Source code in GitHub
- GitHub access for Marek Cermak (CermakM)
- Deployed API URL
- Project and API documentation

---

## 12. Product Requirements Checklist

- [ ] API Type: REST implemented
- [ ] Language: Go used for backend
- [ ] Database: PostgreSQL for editors/newsletters, Firebase for subscribers
- [ ] ORM Constraint: No ORMs used
- [ ] Deployment: Application deployed to Railway
- [ ] Client Support: API serves mobile and web clients
- [ ] Editor Features: Registration, login, JWT, password reset, newsletter CRUD, publish, list subscribers
- [ ] Subscriber Features: Subscribe, confirm, unsubscribe, receive emails
- [ ] Publishing: Email delivery, archive posts
- [ ] Non-Functional: Production-ready, modern stack, docs, transactional context, naming, GitHub, deployed URL
- [ ] Optional: Social authentication (if using Firebase Auth)
- [ ] Open questions/assumptions documented 