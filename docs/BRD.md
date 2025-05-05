# Business Requirements Document: Go Newsletter Platform (Semester Project)

**Version:** 1.0
**Status:** Draft
**Project:** Go Newsletter Platform Backend API (Semester Project)

---

## Table of Contents

1.  **Introduction & Project Background**
2.  **Business Objectives & Goals**
3.  **Stakeholder & User Analysis**
    *   3.1 RACI Matrix
    *   3.2 User Personas
4.  **Value Proposition & Differentiation**
    *   4.1 Value Proposition Canvas
    *   4.2 Unique Selling Points (USPs)
5.  **Business Model & Market Context**
    *   5.1 Business Model Canvas (Project Adaptation)
    *   5.2 Competitive Landscape / Market Context (Porter's Five Forces - Project Lens)
6.  **Business Requirements & Prioritization (MoSCoW)**
    *   6.1 Must Have
    *   6.2 Should Have
    *   6.3 Could Have
    *   6.4 Won't Have (This Iteration)
7.  **Risk & Assumption Analysis**
    *   7.1 SWOT Analysis
    *   7.2 Risk Register
8.  **Success Metrics & KPIs**
9.  **Next Steps & High-Level Timeline**
10. **Appendices (Optional)**

---

## 1. Introduction & Project Background

This document outlines the business requirements for the "Go Newsletter Platform" backend API, developed as a semester project. The primary driver for this project is educational: to design, implement, and deploy a backend system using specific technologies (Go, PostgreSQL without ORM, Firebase) within a constrained timeframe (2 weeks). The system aims to provide core functionalities for managing newsletters, publishing posts, and handling subscriber interactions via a RESTful API. While mimicking a real-world application, the focus remains on demonstrating technical proficiency, architectural understanding, and adherence to project specifications rather than achieving commercial viability.

The technical approach involves building a single Go binary (`api-server`) with internal logical separation (auth, newsletter packages), fronted by a Caddy API gateway, leveraging Firebase for Authentication and Firestore (subscriber data), PostgreSQL (editor/newsletter/post data via `sqlc`), and Resend for email delivery.

## 2. Business Objectives & Goals

*   **Primary Objective:** Successfully deliver a functional backend API meeting all core technical and functional requirements of the semester project specification within the 2-week deadline.
*   **Learning Goal:** Demonstrate understanding and practical application of Go for backend development, microservice architectural concepts (even if implemented as internal packages initially), API design (REST), database interaction (`sqlc`, PostgreSQL, Firestore), and integration with third-party services (Firebase Auth, Resend).
*   **Quality Goal:** Produce a well-documented, testable, and deployable artifact that reflects good software engineering practices, suitable for evaluation as a final product (within the project context).
*   **Portfolio Goal:** Create a tangible project demonstrating relevant skills for potential internships or job applications.

## 3. Stakeholder & User Analysis

### 3.1 RACI Matrix

This RACI matrix reflects the context of a typical semester project.

| Activity / Deliverable         | Developer (Student) | Project Lead (Student/Self) | Course Instructor | TA / Mentor | External Services (Firebase/Resend) |
| :----------------------------- | :------------------: | :-------------------------: | :---------------: | :---------: | :---------------------------------: |
| Define Project Scope           |          C           |              A              |         R         |      I      |                 N/A                 |
| Develop Backend API            |          **R**       |              **A**          |         I         |      C      |                 N/A                 |
| Choose Technology Stack        |          C           |              A              |    **R** (Spec)   |      C      |                 N/A                 |
| Implement Database Schema      |          **R**       |              **A**          |         I         |      C      |                 N/A                 |
| Integrate External Services    |          **R**       |              **A**          |         I         |      C      |                 N/A                 |
| Write Documentation (Code/API) |          **R**       |              **A**          |         I         |      C      |                 N/A                 |
| Write Project Report/README    |          **R**       |              **A**          |         I         |      C      |                 N/A                 |
| Test the Application           |          **R**       |              **A**          |         I         |      C      |                 N/A                 |
| Deploy Application             |          **R**       |              **A**          |         I         |      C      |                 N/A                 |
| Evaluate Final Submission      |          I           |              I              |    **A**, **R**   |      C      |                 N/A                 |
| Provide Project Support        |         N/A          |             N/A             |         C         |   **R**, C  |            **R** (Platform)         |

*(R=Responsible, A=Accountable, C=Consulted, I=Informed)*

### 3.2 User Personas

**Persona 1: Ella the Editor (Content Creator)**

*   **Role:** Student blogger, hobbyist writer using the platform for a personal newsletter.
*   **Goals:**
    *   Easily publish articles to her audience.
    *   Manage multiple newsletter topics simply.
    *   See who has subscribed (basic list).
    *   Not worry about the technical details of email delivery or user logins.
*   **Needs:**
    *   Simple API endpoints for creating newsletters and publishing posts.
    *   Reliable authentication mechanism (doesn't want to build her own).
    *   A straightforward way to view subscribers for a given newsletter.
*   **Pain Points (with other potential tools/manual methods):**
    *   Overly complex platforms with features she doesn't need.
    *   Unreliable email sending or getting marked as spam.
    *   Manually managing subscriber lists and opt-outs.
    *   Building and securing user authentication is difficult.

**Persona 2: Sam the Subscriber (Reader)**

*   **Role:** Follower of Ella's blog/content.
*   **Goals:**
    *   Receive interesting content from Ella via email.
    *   Easily subscribe to newsletters they are interested in.
    *   Easily unsubscribe if the content is no longer relevant.
*   **Needs:**
    *   A simple subscription process (email + newsletter identifier).
    *   Clear confirmation of subscription (implicit or explicit).
    *   A working, easy-to-find unsubscribe link in every email.
*   **Pain Points:**
    *   Receiving spam or irrelevant emails.
    *   Complicated or non-functional unsubscribe processes.
    *   Giving out email addresses without clear value or trust.

## 4. Value Proposition & Differentiation

### 4.1 Value Proposition Canvas

**Customer Profile (Ella - Editor):**

*   **Jobs:** Publish content, manage newsletters, build audience list, authenticate securely.
*   **Pains:** Complex UIs, managing auth, subscriber list maintenance, unreliable email delivery, technical overhead.
*   **Gains:** Simple API workflow, quick publishing, automated subscriber handling (add/remove), reliable auth/email via trusted providers.

**Customer Profile (Sam - Subscriber):**

*   **Jobs:** Discover & read content, manage email subscriptions.
*   **Pains:** Difficult unsubscribe, spam, unclear subscription process.
*   **Gains:** Easy subscribe/unsubscribe, relevant content delivery, trustworthy process.

**Value Map (Go Newsletter Platform API):**

*   **Products/Services:** REST API Endpoints (Newsletter/Post CRUD, Subscribe/Unsubscribe), Firebase Auth Integration, Firestore Subscriber Storage, Resend Email Integration, `sqlc`-based DB access.
*   **Pain Relievers:** Leverages Firebase Auth (removes auth dev burden), automated subscriber storage in Firestore, tokenized one-click unsubscribe, simplified API interactions, reliable email via Resend.
*   **Gain Creators:** Enables programmatic content publishing, provides core newsletter management via API, facilitates easy subscription capture, ensures reliable email dispatch (via Resend), meets specific technical project requirements (Go, no ORM).

### 4.2 Unique Selling Points (USPs) - Within Project Context

1.  **Focused Core Functionality:** Provides essential newsletter API features without bloat, suitable for programmatic integration.
2.  **Leverages Robust External Services:** Integrates Firebase (Auth/Firestore) and Resend, reducing development burden and increasing reliability for core functions like auth, subscriber storage, and email delivery.
3.  **Adherence to Technical Specifications:** Built specifically using Go, PostgreSQL with `sqlc` (no ORM), meeting the unique constraints and learning objectives of the semester project.
4.  **Simplified Architecture:** A single deployable binary (`api-server`) for ease of deployment and management within the project's timeframe and scope.

## 5. Business Model & Market Context

### 5.1 Business Model Canvas (Project Adaptation)

| Key Partners                      | Key Activities                   | Value Propositions                    | Customer Relationships | Customer Segments       |
| :-------------------------------- | :------------------------------- | :------------------------------------ | :------------------- | :---------------------- |
| Firebase (Auth, Firestore)        | Backend API Development          | Core Newsletter API Functionality     | Automated (via API)  | **Editors** (API Users) |
| Resend (Email API)                | Database Design & Implementation | Reliable Auth & Email Delivery      |                      | **Subscribers** (Data)  |
| Cloud Provider (Render/Railway) | Integration with 3rd Parties     | Simplified Management via API       |                      | **Course Instructor**   |
| GitHub (Code Hosting)             | Testing (Unit, Integration)    | Adherence to Project Specs (Go, etc.) |                      | (Evaluator)             |
| *(TA/Mentor)*                     | Documentation                    | Learning Outcome Achievement        |                      |                         |
|                                   | Deployment                       |                                       |                      |                         |
| **Key Resources**                 |                                  | **Channels**                          |                      |                         |
| Developer Time & Expertise (Go)   |                                  | API Endpoints                         |                      |                         |
| Go Toolchain, Docker              |                                  | Email (for Subscribers)               |                      |                         |
| Firebase/Resend Accounts          |                                  | Project Documentation (README, API Spec) |                      |                         |
| Cloud Hosting Environment         |                                  |                                       |                      |                         |
| **Cost Structure**                | **Revenue Streams**                                                                                   |                      |                         |
| **Developer Time (Opportunity Cost)** | **Project Grade / Successful Completion**                                                            |                      |                         |
| Cloud Hosting Fees (Free Tier?)   | **Learning & Portfolio Value**                                                                        |                      |                         |
| External Service Fees (Free Tier?) | *(None - Not a commercial product)*                                                                 |                      |                         |

### 5.2 Competitive Landscape / Market Context (Porter's Five Forces - Project Lens)

*   **Threat of New Entrants:** *High* in the real market, but *Low/Irrelevant* for this project's success criteria. The goal isn't market share, but meeting requirements.
*   **Bargaining Power of Buyers (Users/Instructor):** *High* for the Instructor (defines requirements and grades). *Low* for hypothetical end-users (no actual market transaction).
*   **Bargaining Power of Suppliers (Firebase, Resend):** *Medium*. Essential services, but alternatives exist if needed (though switching would impact the project timeline/scope). Reliant on their free tier limits.
*   **Threat of Substitutes:** *High*. Many established newsletter platforms and APIs exist (Substack, Mailchimp). However, they don't fulfill the specific *project learning requirement* of building it with Go/sqlc.
*   **Industry Rivalry:** *High* in the real market, but *N/A* for the project. The "competition" is against the project requirements and timeline.

**Conclusion:** Market forces are less relevant than the need to meet the specific educational objectives and technical constraints defined by the project specification within the given timeframe. The primary "business pressure" is delivering a high-quality, functional artifact for evaluation.

## 6. Business Requirements & Prioritization (MoSCoW)

These requirements translate the technical features into business needs for the project's success.

### 6.1 Must Have (Essential for MVP & Project Pass)

*   **BR-M-01:** The platform **must** allow new Editors to register using email/password (via Firebase Auth).
*   **BR-M-02:** The platform **must** allow registered Editors to authenticate (via Firebase Auth) and use the API via JWTs.
*   **BR-M-03:** The system **must** validate JWTs for protected API endpoints.
*   **BR-M-04:** Authenticated Editors **must** be able to create a Newsletter via the API.
*   **BR-M-05:** Authenticated Editors **must** be able to publish a Post to a Newsletter they own via the API.
*   **BR-M-06:** Any user **must** be able to subscribe to a newsletter via a public API endpoint, providing email and newsletter identifier.
*   **BR-M-07:** Subscribers **must** be able to unsubscribe using a unique link/token provided via API/email.
*   **BR-M-08:** The system **must** send published post content via email to active subscribers of that newsletter (using Resend).
*   **BR-M-09:** The system **must** use PostgreSQL with `sqlc` (no ORM) for storing Editor, Newsletter, and Post data.
*   **BR-M-10:** The system **must** use Firebase Firestore for storing Subscriber data.
*   **BR-M-11:** The core API functionality **must** be accessible via a configured API Gateway (Caddy).
*   **BR-M-12:** The application **must** be deployable (e.g., using Docker on a cloud platform).
*   **BR-M-13:** Basic project documentation (README) explaining setup and core concepts **must** be provided.
*   **BR-M-14:** A basic health check endpoint **must** be available.

### 6.2 Should Have (Important for Quality & Completeness)

*   **BR-S-01:** Authenticated Editors **should** be able to list the newsletters they own.
*   **BR-S-02:** Authenticated Editors **should** be able to list the posts belonging to a newsletter they own.
*   **BR-S-03:** Authenticated Editors **should** be able to retrieve their own basic details (`/me`).
*   **BR-S-04:** The API **should** be documented using OpenAPI/Swagger specification.
*   **BR-S-05:** The application **should** implement structured logging for requests and errors.
*   **BR-S-06:** Basic unit and integration tests **should** cover core functionality.

### 6.3 Could Have (Desirable if Time Permits)

*   **BR-C-01:** Authenticated Editors **could** be able to update Newsletter details (name, description).
*   **BR-C-02:** Authenticated Editors **could** be able to delete a Newsletter they own (cascading post deletion).
*   **BR-C-03:** Authenticated Editors **could** be able to retrieve details of a specific Newsletter or Post.
*   **BR-C-04:** Authenticated Editors **could** be able to list the emails of subscribers for their newsletters.
*   **BR-C-05:** The platform **could** support the full editor account deletion flow, including cleanup of related data (simplified approach acceptable for MVP).
*   **BR-C-06:** A password reset flow (leveraging Firebase) **could** be supported.
*   **BR-C-07:** A confirmation/welcome email **could** be sent upon subscription.
*   **BR-C-08:** Basic rate limiting **could** be applied to public API endpoints.
*   **BR-C-09:** Handling of email sending failures **could** be more robust (e.g., marking post status).

### 6.4 Won't Have (This Iteration - Explicitly Excluded)

*   **BR-W-01:** Post Scheduling functionality **won't** be implemented.
*   **BR-W-02:** A GraphQL API endpoint **won't** be provided.
*   **BR-W-03:** Advanced email analytics (opens, clicks) **won't** be implemented.
*   **BR-W-04:** Complex user roles or permissions beyond ownership **won't** be implemented.
*   **BR-W-05:** A frontend web or mobile client **won't** be developed as part of this backend project.

## 7. Risk & Assumption Analysis

### 7.1 SWOT Analysis

| Strengths                                     | Weaknesses                                             |
| :-------------------------------------------- | :----------------------------------------------------- |
| Clear Project Plan & Architecture Document    | **Strict 2-Week Timeline**                             |
| Modern & Relevant Tech Stack (Go, Firebase) | Solo Developer (Junior Experience Level Assumed)         |
| Defined Scope (via PRD & MoSCoW)              | Reliance on External Service Free Tiers (Limits?)      |
| Focus on Backend Skills Development           | Simplified Data Consistency Model (Editor Deletion MVP) |
| Use of Code Generation (`sqlc`) for Safety    | Limited Time for Thorough Testing & Refinement        |
|                                               | Single Point of Failure (Solo Developer)               |
| **Opportunities**                             | **Threats**                                            |
| Demonstrate In-Demand Technical Skills        | **Scope Creep / Over-Engineering**                     |
| Create a Valuable Portfolio Piece             | **External Service Outages or API Changes**            |
| Achieve High Marks for the Project            | Underestimation of Technical Complexity / Debug Time |
| Learn Microservice Concepts & Trade-offs      | Deployment Environment Issues / Configuration Hell   |
| Potential for Future Expansion (Post-Project) | **Failure to Meet Core 'Must Have' Requirements**      |
|                                               | Burnout / Lack of Buffer Time                          |

### 7.2 Risk Register

| Risk ID | Description                                                 | Impact | Likelihood | Mitigation Strategy                                                                                                             | Status     |
| :------ | :---------------------------------------------------------- | :----- | :--------- | :------------------------------------------------------------------------------------------------------------------------------ | :--------- |
| R01     | **Project Exceeds 2-Week Deadline**                         | High   | High       | Strict adherence to MoSCoW prioritization. Simplify architecture (single binary). Timebox features aggressively. Get help early. | Monitoring |
| R02     | **Core 'Must Have' Functionality Incomplete or Buggy**      | High   | Medium     | Focus development effort on 'Must Haves' first (Phase 1 MVP). Implement basic tests early for core flows. Demo frequently.    | Monitoring |
| R03     | **External Service Issues (Firebase/Resend Limits/Downtime)** | Medium | Low        | Understand free tier limits. Implement basic error handling/logging around API calls. Have testing fallbacks (log-only email). | Planned    |
| R04     | **Deployment Complexity / Environment Issues**              | Medium | Medium     | Use familiar PaaS (Render/Railway). Containerize early (Docker). Test deployment pipeline early and often. Keep config simple. | Planned    |
| R05     | **Underestimation of `sqlc` / Go / Integration Effort**     | Medium | Medium     | Allocate specific learning/buffer time. Start complex integrations early. Use established libraries and examples.               | Monitoring |
| R06     | **Scope Creep Introduced During Development**               | High   | Medium     | Rigorously evaluate any new ideas against MoSCoW and timeline. Defer all non-essentials to post-project phase.              | Planned    |

## 8. Success Metrics & KPIs

Success for this project is measured by meeting the educational and delivery objectives.

| KPI Category          | Success Metric                                                                 | Target                                                                                                | Measurement Method                                    | Linked Requirements       |
| :-------------------- | :----------------------------------------------------------------------------- | :---------------------------------------------------------------------------------------------------- | :---------------------------------------------------- | :------------------------ |
| **Project Completion**  | % of 'Must Have' Business Requirements (MoSCoW) Implemented                  | 100%                                                                                                  | Code Review, Functional Demo, Test Results            | All BR-M-\*             |
| **Functionality**       | Successful Execution of Core User Flows (Publish, Subscribe, Unsubscribe Demo) | Pass / Fail                                                                                           | Live Demonstration during Evaluation                | BR-M-05, BR-M-06, BR-M-07 |
| **Code Quality**        | Adherence to Go Standards (`gofmt`, `go vet`, Linter)                          | 0 Errors/Warnings from Standard Tools                                                               | Automated Tool Execution (e.g., in CI/local hook)     | Implied (Quality Goal)    |
|                       | Test Coverage for Core Logic Packages                                          | > 60% (Target, adjust based on complexity/time)                                                       | Code Coverage Reports (e.g., `go test -cover`)        | Implied (Quality Goal)    |
| **Documentation**     | Completion of README.md (Setup, Arch Overview)                                 | Complete & Clear                                                                                      | Manual Review                                         | BR-M-13                   |
|                       | Availability & Basic Completeness of OpenAPI Spec                            | Available & Documents core endpoints                                                                  | Check `/docs` endpoint / spec file existence & content | BR-S-04                   |
| **Deployment**        | Application successfully deployed and accessible on Cloud Platform             | Pass / Fail                                                                                           | Accessing deployed API endpoint / Health Check      | BR-M-12, BR-M-14          |
| **Constraint Adherence**| Use of specified technologies (Go, sqlc, PG, Firebase)                       | Confirmed                                                                                             | Code Review                                           | BR-M-09, BR-M-10          |

## 9. Next Steps & High-Level Timeline

The project will proceed in phases aligned with the PRD roadmap, targeting completion within the 2-week window.

| Phase                       | Key Activities                                                                     | Estimated Duration | Target Completion | Dependencies                                 |
| :-------------------------- | :--------------------------------------------------------------------------------- | :----------------- | :---------------- | :------------------------------------------- |
| **Phase 0: Foundation**     | Setup Repo, Go module, DB, `sqlc`, Firebase, Caddy, Docker, Basic Config/Logging    | 1-2 Days           | Day 2             | Access to Firebase, PG instance            |
| **Phase 1: MVP Core Build** | Implement Must-Have Auth, Newsletter Create, Post Publish, Subscribe/Unsubscribe, Email Send | 5-6 Days           | Day 7             | Phase 0 Complete, Resend API Key           |
|                             | Basic Tests for MVP Flows, Initial Cloud Deployment                                |                    |                   | Cloud Platform Account                     |
| **Phase 2: Enhance & Refine**| Implement Should/Could features based on priority, Increase Test Coverage          | 5-6 Days           | Day 13            | MVP Functioning                            |
|                             | Finalize Documentation (README, OpenAPI), Refactor Code, Final Deployment Testing  |                    |                   |                                              |
| **Final Delivery**          | Submit Codebase, Documentation, Deployed Application Link                            | 1 Day              | Day 14            | All previous phases complete               |

**Key Milestones:**

*   **End of Day 2:** Local environment setup complete, basic project structure runnable.
*   **End of Day 7:** Core MVP functionality (Publish, Subscribe, Unsubscribe) demonstrable, potentially deployed.
*   **End of Day 13:** All planned features implemented, documentation drafted, testing near complete.
*   **End of Day 14:** Final submission package ready.

---

*End of BRD*