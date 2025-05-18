# Master RFC List & Implementation Roadmap: Go Newsletter Platform

## Table of Contents
1. Introduction
2. RFC List (Sequential Order)
3. Dependency Graph
4. Implementation Roadmap

---

## 1. Introduction

This document breaks down the Go Newsletter Platform into strictly sequential RFCs (Request for Comments), each representing a cohesive implementation unit. Each RFC references specific features from the features list and is designed to be implemented in order, respecting dependencies.

---

## 2. RFC List (Strict Sequential Order)

| RFC ID   | Title                                 | Summary                                                      | Included Features                        | Dependencies |
|----------|---------------------------------------|--------------------------------------------------------------|------------------------------------------|--------------|
| RFC-001  | Project Setup & Tooling               | Initialize repo, CI, Go modules, DB schema, base structure   | NFR-02, NFR-07, NFR-08                   | None         |
| RFC-002  | Editor Auth & Account Management      | Registration, login, JWT, password reset                     | FE-ED-01, FE-ED-02, FE-ED-03, FE-ED-04   | RFC-001      |
| RFC-003  | Newsletter CRUD                       | Create, rename, delete newsletter                            | FE-ED-06, FE-ED-07, FE-ED-08             | RFC-002      |
| RFC-004  | Subscriber Management                 | Subscribe, confirm, unsubscribe                              | FE-SB-01, FE-SB-02, FE-SB-03             | RFC-003      |
| RFC-005  | Publishing & Email Delivery           | Publish post, email delivery, archive posts                  | FE-PB-01, FE-PB-02, FE-PB-03             | RFC-004      |
| RFC-006  | List Subscribers                      | Editor can view newsletter subscribers                       | FE-ED-09                                 | RFC-004      |
| RFC-007  | Non-Functional: Docs, Quality, Naming | API docs, project docs, naming, production readiness         | NFR-01, NFR-03, NFR-04, NFR-06           | RFC-001-006  |
| RFC-008  | Optional: Social Auth                 | Social login via Firebase Auth                               | FE-ED-05                                 | RFC-002      |

---

## 3. Dependency Graph

- RFC-001 → RFC-002 → RFC-003 → RFC-004 → RFC-005
- RFC-006 depends on RFC-004 (can be parallel with RFC-005)
- RFC-007 is finalized after all core RFCs (RFC-001 to RFC-006)
- RFC-008 (optional) depends on RFC-002

---

## 4. Implementation Roadmap

1. **RFC-001:** Project setup, repo, CI, DB, base structure
2. **RFC-002:** Editor registration, login, JWT, password reset
3. **RFC-003:** Newsletter CRUD (create, rename, delete)
4. **RFC-004:** Subscriber management (subscribe, confirm, unsubscribe)
5. **RFC-005:** Publishing (post, email delivery, archive)
6. **RFC-006:** List subscribers (editor view)
7. **RFC-007:** Non-functional: docs, naming, production readiness
8. **RFC-008:** (Optional) Social authentication

---

Each RFC will be detailed in its own file in the `project_management/RFCS/` subfolder (e.g., `RFC-001-Project-Setup.md`). 