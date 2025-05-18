# RFC-003: Newsletter CRUD

## Summary
Implement create, rename, and delete operations for newsletters. Each newsletter is owned by a single editor.

## Included Features
- FE-ED-06: Create Newsletter
- FE-ED-07: Rename Newsletter
- FE-ED-08: Delete Newsletter

## Dependencies
- RFC-002

## Technical Approach
- Endpoints for create, update (rename/description), and delete newsletter
- Enforce single-editor ownership and name uniqueness per editor
- Cascade or handle deletion of related posts and subscribers
- Input validation and ownership checks

## Acceptance Criteria
- Editors can create, rename, and delete their newsletters
- Ownership and uniqueness enforced
- Related data handled on delete
- Tests for all flows

## APIs
- POST /newsletters
- PATCH /newsletters/{id}
- DELETE /newsletters/{id}

## Data Models
- newsletters (id, editor_id, name, description, created_at, ...)

## State
- Newsletter CRUD is functional and secure

## Error Handling
- Standardized error responses (JSON)
- Ownership and validation errors handled

## Testing
- Unit and integration tests for all endpoints

## Notes
- Deletion must address related posts and subscribers 