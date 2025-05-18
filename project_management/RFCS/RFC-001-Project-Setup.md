# RFC-001: Project Setup & Tooling

## Summary
Establish the foundational project structure, repository, CI, Go modules, initial database schema, and base tooling for the Go Newsletter Platform.

## Included Features
- NFR-02: Modern Stack & Architecture
- NFR-07: GitHub Repository
- NFR-08: Deployed API URL

## Dependencies
- None

## Technical Approach
- Initialize GitHub repository and invite Marek Cermak (CermakM)
- Set up Go modules and base directory structure
- Configure CI (e.g., GitHub Actions) for linting, tests, and build
- Create initial PostgreSQL schema (editors, newsletters, posts tables)
- Set up Railway project for deployment
- Add basic README and documentation structure

## Acceptance Criteria
- Repo exists, CermakM invited
- Go module initialized, builds pass
- CI runs on push/PR
- DB schema committed (migrations)
- Railway deployment works, API base URL available
- README with setup instructions

## APIs/Data Models
- No functional APIs yet; only base structure and migrations
- Data models: editors, newsletters, posts (initial tables)

## State
- Project is ready for feature development

## Error Handling
- N/A (setup phase)

## Testing
- CI runs basic test suite (even if empty)

## Notes
- All future RFCs depend on this setup 